FROM golang:1.15 as builder

WORKDIR /go/src
COPY . .

RUN GOPROXY="https://proxy.golang.org|direct" go mod download
RUN make build/operator-sdk

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

RUN microdnf install -y golang make which

COPY --from=builder /go/src/build/operator-sdk /usr/local/bin/operator-sdk

ENTRYPOINT ["/usr/local/bin/operator-sdk"]
