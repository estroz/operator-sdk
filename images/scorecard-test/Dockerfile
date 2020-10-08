ARG BUILDPLATFORM
FROM --platform=$BUILDPLATFORM golang:1.15-alpine as builder

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GOPROXY="https://proxy.golang.org|direct"
ENV CGO_ENABLED=0

ENV WD=/app
WORKDIR $WD
COPY . .

RUN go mod download
RUN go build -o scorecard-test -gcflags "all=-trimpath=${WD}" -asmflags "all=-trimpath=${WD}" ./images/scorecard-test

# Base image
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV HOME=/opt/scorecard-test \
    USER_NAME=scorecard-test \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

WORKDIR ${HOME}
USER ${USER_UID}

COPY --from=builder /app/scorecard-test /usr/local/bin/scorecard-test
ENTRYPOINT ["/usr/local/bin/scorecard-test"]
