FROM golang:1.15-alpine as builder

ENV WD=/go/src
WORKDIR $WD
COPY images/custom-scorecard-tests .

ENV CGO_ENABLED=0
RUN go mod init custom-scorecard-tests
RUN GOPROXY="https://proxy.golang.org|direct" go mod download
RUN go build -o custom-scorecard-tests -gcflags "all=-trimpath=${WD}" -asmflags "all=-trimpath=${WD}"

# Base image
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV HOME=/opt/custom-scorecard-tests \
    USER_NAME=custom-scorecard-tests \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

WORKDIR ${HOME}
USER ${USER_UID}

COPY --from=builder /go/src/custom-scorecard-tests /usr/local/bin/custom-scorecard-tests
ENTRYPOINT ["/usr/local/bin/custom-scorecard-tests"]
