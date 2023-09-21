FROM scratch

ARG application

ENV APP=/usr/local/bin/${application} \
    USER_UID=1101 \
    USER_NAME=mohamedaly

# install engine binary
COPY build/_output/bin/${application} /
USER ${USER_UID}:${USER_UID}
ENTRYPOINT [ "./master" ]
