FROM --platform=$BUILDPLATFORM golang:1.15-alpine as builder

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GOPROXY="https://proxy.golang.org|direct"
ENV CGO_ENABLED=0

ENV WD=/app
WORKDIR $WD
COPY images/scorecard-test-kuttl .

RUN go mod init scorecard-test-kuttl && go mod download
RUN go build -o scorecard-test-kuttl -gcflags "all=-trimpath=${WD}" -asmflags "all=-trimpath=${WD}"

# Base image
#FROM docker.io/kudobuilder/kuttl@sha256:c9c5edc27ba6e5e994ffd8cea03368c0d4e274f3c7a00f51c1e360feb573f24e
FROM kudobuilder/kuttl:v0.6.1

ENV HOME=/opt/scorecard-test-kuttl \
    USER_NAME=scorecard-test-kuttl \
    USER_UID=1001 \
    KUBECONFIG=

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

WORKDIR ${HOME}
USER ${USER_UID}

COPY --from=builder /app/scorecard-test-kuttl /usr/local/bin/scorecard-test-kuttl
ENTRYPOINT ["/usr/local/bin/scorecard-test-kuttl"]
