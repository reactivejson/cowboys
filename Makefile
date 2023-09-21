# Go env variables
export GOFLAGS		?= -mod=vendor
export GO111MODULE	?= on

# Service env variables
export METRICS_ADDR			?= :8081


# Build variables
.DEFAULT_GOAL		:= help
PLAYER_IMAGE_NAME	?= player
MASTER_IMAGE_NAME	?= master
PLAYER_APP_NAME	    ?= player
MASTER_APP_NAME	    ?= master
COMMIT_ID			?= snapshot
BUILD_VERSION		?= 0.0.0-snapshot
DOCKER_REGISTRY		?= 127.0.0.1:5000

TEST_REPORT_DIR			:= target/test
COVERAGE_FILE_SUFFIX	:= coverage.txt

# List of tools that can be installed with go get
TOOLS_DIR		:= .tools/
GOTESTSUM		:= ${TOOLS_DIR}gotest.tools/gotestsum@v1.7.0

${GOTESTSUM}:
	$(eval TOOL=$(@:%=%))
	@echo Installing ${TOOL}...
	go install $(TOOL:${TOOLS_DIR}%=%)
	@mkdir -p $(dir ${TOOL})
	@cp ${GOBIN}/$(firstword $(subst @, ,$(notdir ${TOOL}))) ${TOOL}

COVER_PKGS			= $(subst ${SPACE},${COMMA},$(shell go list ./...))
UNIT_TEST_FLAGS		= -race -cover
UNIT_TEST_FLAGS		+= -coverprofile=${TEST_REPORT_DIR}/unit_test_${COVERAGE_FILE_SUFFIX}
INT_TEST_FLAGS		= -race -cover -coverpkg=${COVER_PKGS} -tags=integration -run='^TestIntegration'
INT_TEST_FLAGS		+= -coverprofile=${TEST_REPORT_DIR}/integration_test_${COVERAGE_FILE_SUFFIX}

.PHONY: help
help:  ## Display help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: run
run: ## Run service locally with default values
	go run cmd/main.go

builds/${PLAYER_APP_NAME}:
	env GOOS=linux CGO_ENABLED=0 go build -o build/_output/bin/${PLAYER_APP_NAME} cmd/${PLAYER_APP_NAME}/main.go

builds/${MASTER_APP_NAME}:
	env GOOS=linux CGO_ENABLED=0 go build -o build/_output/bin/${MASTER_APP_NAME} cmd/${MASTER_APP_NAME}/main.go


.PHONY: build
build:
	env GOOS=linux CGO_ENABLED=0 go build -o build/_output/bin/${PLAYER_APP_NAME} cmd/${PLAYER_APP_NAME}/main.go
	env GOOS=linux CGO_ENABLED=0 go build -o build/_output/bin/${MASTER_APP_NAME} cmd/${MASTER_APP_NAME}/main.go

.PHONY: test
test: ${GOTESTSUM} ## Run unit tests
	@mkdir -p ${TEST_REPORT_DIR}
	${GOTESTSUM} --jsonfile ${TEST_REPORT_DIR}/units-tests-output.log --junitfile=${TEST_REPORT_DIR}/unit-junit.xml -- ${UNIT_TEST_FLAGS} ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	@mkdir -p ${TEST_REPORT_DIR}
	${GOTESTSUM} --jsonfile ${TEST_REPORT_DIR}/integration-tests-output.log --junitfile=${TEST_REPORT_DIR}/integration-junit.xml -- ${INT_TEST_FLAGS} ./...



.PHONY: generate-docker-compose
generate-docker-compose: ## Start environment for running integration type of tests
	pip install jinja-docker-compose
	jinja-docker-compose -f docker-compose.yml.j2 -D $(players)

.PHONY: run-app
run-app: generate-docker-compose ## Start environment for running integration type of tests
	docker compose up everything

.PHONY: environment-stop
environment-stop: ## Stop environment for running integration type of tests
	docker compose down

.PHONY: docker
docker: builds/${PLAYER_APP_NAME} builds/${MASTER_APP_NAME}
	docker build --rm -t ${PLAYER_APP_NAME} --build-arg application=${PLAYER_APP_NAME}  build
	docker build --rm -t ${MASTER_APP_NAME} --build-arg application=${MASTER_APP_NAME}  build

.PHONY: lint
lint:
	mkdir -p target
ifeq (, $(shell which golangci-lint))
	docker run --rm -v $(shell pwd):/work -w /work neo-docker-releases.repo.lab.pl.alcatel-lucent.com/go-tools:1.15.5-39 golangci-lint run --out-format checkstyle 2>&1 | tee target/lint-report.xml
else
	golangci-lint run --out-format checkstyle 2>&1 | tee target/lint-report.xml
endif


.PHONY: docker-build
docker-build: build ## Build service docker image
	docker build --rm -t ${PLAYER_APP_NAME}  --build-arg application=${PLAYER_APP_NAME} --file build/${PLAYER_APP_NAME}/Dockerfile .
	docker tag ${PLAYER_APP_NAME} ${DOCKER_REGISTRY}/${PLAYER_APP_NAME}:latest
	docker tag ${PLAYER_APP_NAME} ${DOCKER_REGISTRY}/${PLAYER_APP_NAME}:${BUILD_VERSION}
	docker build --rm -t ${MASTER_APP_NAME}  --build-arg application=${MASTER_APP_NAME}  --file build/${MASTER_APP_NAME}/Dockerfile .
	docker tag ${MASTER_APP_NAME} ${DOCKER_REGISTRY}/${MASTER_APP_NAME}:latest
	docker tag ${MASTER_APP_NAME} ${DOCKER_REGISTRY}/${MASTER_APP_NAME}:${BUILD_VERSION}

.PHONY: docker-push
docker-push: ## Publish docker image
	docker push ${DOCKER_REGISTRY}/${PLAYER_APP_NAME}:latest
	docker push ${DOCKER_REGISTRY}/${PLAYER_APP_NAME}:${BUILD_VERSION}
	docker push ${DOCKER_REGISTRY}/${MASTER_APP_NAME}:latest
	docker push ${DOCKER_REGISTRY}/${MASTER_APP_NAME}:${BUILD_VERSION}

.PHONY: vendor
vendor: ## Update vendor folder to match go.mod
	go mod tidy
	go mod vendor

.PHONY: helm-create
helm-create: ## Create Helm package
	rm -rf target/helm && mkdir -p target/helm
	cp -r helm/$(PLAYER_APP_NAME) target/helm/$(PLAYER_APP_NAME)
	cp -r helm/$(MASTER_APP_NAME) target/helm/$(MASTER_APP_NAME)
	cp -f README.md target/helm/$(PLAYER_APP_NAME)
	cp -f README.md target/helm/$(MASTER_APP_NAME)
	docker run --rm -v "$(CURDIR)/target:/workdir" \
		-e APP_NAME=$(PLAYER_APP_NAME) \
		-e CUSTOM_SCHEMA_VALIDATION="false"  \
		-e SKIP_SCHEMA_VALIDATION_TEST="false" \
		-e COMMIT_ID=$(COMMIT_ID) \
		-e BUILD_VERSION=$(BUILD_VERSION) \
		-e DOCKER_REGISTRY=$(DOCKER_REGISTRY) \
		-e DOCKER_RELEASE_REGISTRY=$(DOCKER_RELEASE_REGISTRY) \
		-e TARGET_K8S_VERSIONS="$(TARGET_K8S_VERSIONS)" \
		neo-docker-releases.repo.lab.pl.alcatel-lucent.com/helm-builder:latest
	docker run --rm -v "$(CURDIR)/target:/workdir" \
		-e APP_NAME=$(MASTER_APP_NAME) \
		-e CUSTOM_SCHEMA_VALIDATION="false"  \
		-e SKIP_SCHEMA_VALIDATION_TEST="false" \
		-e COMMIT_ID=$(COMMIT_ID) \
		-e BUILD_VERSION=$(BUILD_VERSION) \
		-e DOCKER_REGISTRY=$(DOCKER_REGISTRY) \
		-e DOCKER_RELEASE_REGISTRY=$(DOCKER_RELEASE_REGISTRY) \
		-e TARGET_K8S_VERSIONS="$(TARGET_K8S_VERSIONS)" \
		neo-docker-releases.repo.lab.pl.alcatel-lucent.com/helm-builder:latest
