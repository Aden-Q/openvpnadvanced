@help
    just -l

@vendor
    @echo "Running 'go mod vendor' to ensure all dependencies are vendored..."
    go mod vendor

@build
    @echo "Building the project..."
    docker buildx bake build

@lint
    @echo "Running linter..."
    # Using golangci-lint for linting
    docker buildx bake lint

@test
    @echo "Running tests..."
    # Running tests with coverage
    docker buildx bake test
