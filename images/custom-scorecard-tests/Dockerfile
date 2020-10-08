FROM --platform=$BUILDPLATFORM golang:1.15.2-alpine as builder

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GOPROXY="https://proxy.golang.org|direct"
ENV CGO_ENABLED=0

ENV WD=/app
WORKDIR $WD
COPY images/custom-scorecard-tests .

RUN go mod init custom-scorecard-tests && go mod download
RUN go build -o custom-scorecard-tests -gcflags "all=-trimpath=${WD}" -asmflags "all=-trimpath=${WD}"

# Base image
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV HOME=/opt/custom-scorecard-tests \
    USER_NAME=custom-scorecard-tests \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

WORKDIR ${HOME}
USER ${USER_UID}

COPY --from=builder /app/custom-scorecard-tests /usr/local/bin/custom-scorecard-tests
ENTRYPOINT ["/usr/local/bin/custom-scorecard-tests"]
