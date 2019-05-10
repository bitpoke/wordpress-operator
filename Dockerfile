# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/github.com/presslabs/wordpress-operator
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/presslabs/wordpress-operator/cmd/manager

# Copy the controller-manager into a thin image
FROM gcr.io/distroless/static:a4fd5de337e31911aeee2ad5248284cebeb6a6f4
COPY --from=builder /go/src/github.com/presslabs/wordpress-operator/manager /
USER nobody
ENTRYPOINT ["/manager"]
