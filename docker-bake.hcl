group "default" {
    targets = ["lint", "build", "test"]
}

target "lint" {
    dockerfile = "Dockerfile"
    context = "."
    tags = ["lint"]
    target = "lint"
    cache-to = ["type=gha,mode=max"]
    cache-from = ["type=gha"]
}

target "build" {
    dockerfile = "Dockerfile"
    context = "."
    tags = ["build"]
    target = "build"
    cache-to = ["type=gha,mode=max"]
    cache-from = ["type=gha"]
}

target "test" {
    dockerfile = "Dockerfile"
    context = "."
    tags = ["test"]
    target = "test"
    cache-to = ["type=gha,mode=max"]
    cache-from = ["type=gha"]
}
