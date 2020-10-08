FROM --platform=$BUILDPLATFORM golang:1.15.2 as builder

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
RUN make build/helm-operator

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV HOME=/opt/helm \
    USER_NAME=helm \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

WORKDIR ${HOME}
USER ${USER_UID}

COPY --from=builder /app/build/helm-operator /usr/local/bin/helm-operator

ENTRYPOINT ["/usr/local/bin/helm-operator", "run", "--watches-file=./watches.yaml"]
