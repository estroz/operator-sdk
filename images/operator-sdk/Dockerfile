FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

RUN microdnf install -y golang make which

ARG TARGETOS
ARG TARGETARCH
ARG BIN=dist/operator-sdk_${TARGETOS}_${TARGETARCH}/operator-sdk
COPY $BIN /usr/local/bin/operator-sdk

ENTRYPOINT ["/usr/local/bin/operator-sdk"]
