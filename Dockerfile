ARG GOLANGCI_LINT_VERSION=v1.62.2
ARG GO_VERSION=1.23

FROM golang:${GO_VERSION} AS build-artifacts
WORKDIR /app
COPY . .
RUN go mod vendor

# Linting stage
FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION} AS lint
WORKDIR /app
COPY --from=build-artifacts /app /app
RUN golangci-lint run

# Build the application
FROM build-artifacts AS build
WORKDIR /app
RUN go build ./...

# Testing stage
FROM build-artifacts AS test
WORKDIR /app
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest
ENTRYPOINT [ "ginkgo" ]
CMD ["run", "-r", "-race", "-cover"]
