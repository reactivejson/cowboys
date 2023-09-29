## Distributed Cowboys game


![Maintainer](https://img.shields.io/badge/maintainer-MohamedAly-blue)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/reactivejson/cowboys.svg)](https://github.com/reactivejson/cowboys)
[![Go Reference](https://pkg.go.dev/badge/github.com/reactivejson/cowboys)](https://pkg.go.dev/badge/github.com/reactivejson/cowboys)
[![Go](https://github.com/reactivejson/cowboys/actions/workflows/go.yml/badge.svg)](https://github.com/reactivejson/cowboys/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/reactivejson/cowboys)](https://goreportcard.com/report/github.com/reactivejson/cowboys)
![](https://img.shields.io/github/license/reactivejson/cowboys.svg)

Go implementation of Distributed cowboys shootout game.
You can run the game in two ways:

 - You can run this game using docker compose where we have a master to orchestrate the game and each player will run in it's own container.
 - You can run this game using Kubernetes platform with helm

### Project layout

This layout is following pattern:

```text
cowboys
└───
    ├── .github
    │   └── workflows
    │     └── go.yml
    ├── cmd
    │   └── master
        │   └── main.go
        │   └── app
        │     └── setup.go
        │     └── app.go
        │     └── context.go
    │   └── player
        │   └── main.go
        │   └── app
        │     └── setup.go
        │     └── app.go
        │     └── context.go
    ├── internal
    │   └── app
    │     └── master.go
    │     └── player.go
    │   └── domain
    │     └── master.go
    │     └── player.go
    │   └── game
    │     └── event.go
    │     └── game-state.go
    ├── build
    │   └── Dockerfile
    ├── helm
    │   └── <helm chart files>
    ├── docker-compose.yml.j2
    ├── Makefile
    ├── README.md
    └── <source packages>
```


## Setup

### Getting started
cowboys game is available in github
[cowboys](https://github.com/reactivejson/cowboys)

```shell
go get github.com/reactivejson/cowboys
```

#### Build

build the app and docker images
To build the components of the project, use the following commands:

```shell
make docker-build
```

This will build this application docker images so-called master and player

#### Deploy it and run it with docker compose
We use Jinja2 to run the application dynamically.

Use the following command to run the game and provide your list of players (or use the default one)
```shell
make run-app players=players.json
```

Example input:
````json
{
  "players": [
    {
      "name": "p1",
      "health": 10,
      "damage": 3
    },
    {
      "name": "p2",
      "health": 5,
      "damage": 4
    },
    {
      "name": "p3",
      "health": 10,
      "damage": 1
    },
    {
      "name": "p4",
      "health": 7,
      "damage": 2
    }
  ]
}

````

And check the logs to see who won:)
```shell
docker compose logs -f
```

#### Testing
```shell
make test
```

### Deploy with Helm chart in a Kubernetes environment
In order to deploy it in a Kubernetes platform with helm.
Create Helm package

```bash
make helm-create
helm upgrade --namespace neo --install master chart/master -f <your-custom-values>.yml
helm upgrade --namespace neo --install player chart/player -f <your-custom-values-with-players>.yml
```

## Test coverage

Test coverage is checked as a part of test execution with the gotestsum tool.

Test coverage is checked for unit tests and integration tests.

Coverage report files are available and stored as `*coverage.txt` and are also imported in the SonarQube for easier browsing.


## golangci-lint

In the effort of reducing errors and improving the overall quality of code, golangci-lint is run as a part of the pipeline. Linting is run for the services and packages that have changes since the previous green build (in master) or previous commit (in local or review).

Any issues found by golangci-lint for the changed code will lead to a failed build.

golangci-lint rules are configured in `.golangci.yml`.


### Requirements

- Go 1.18 or newer [https://golang.org/doc/install](https://golang.org/doc/install)
- Docker 18.09.6 or newer

### Variable names
Commonly used one letter variable names:

- i for index
- r for reader
- w for writer
- c for client


## License

Apache 2.0, see [LICENSE](LICENSE).
