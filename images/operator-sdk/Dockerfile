FROM --platform=$BUILDPLATFORM golang:1.15.2 as builder

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GOPROXY="https://proxy.golang.org|direct"
ENV CGO_ENABLED=0

ENV WD=/app/operator-sdk
WORKDIR $WD
COPY . .

RUN go mod download
RUN make build/operator-sdk

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV CGO_ENABLED=0

RUN microdnf install -y golang make which

COPY --from=builder /app/operator-sdk/build/operator-sdk /usr/local/bin/operator-sdk

ENTRYPOINT ["/usr/local/bin/operator-sdk"]
