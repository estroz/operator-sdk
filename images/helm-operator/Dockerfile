FROM golang:1.15 as builder

WORKDIR /go/src
COPY . .

RUN GOPROXY="https://proxy.golang.org|direct" go mod download
RUN make build/helm-operator

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV HOME=/opt/helm \
    USER_NAME=helm \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

WORKDIR ${HOME}
USER ${USER_UID}

COPY --from=builder /go/src/build/helm-operator /usr/local/bin/helm-operator

ENTRYPOINT ["/usr/local/bin/helm-operator", "run", "--watches-file=./watches.yaml"]
