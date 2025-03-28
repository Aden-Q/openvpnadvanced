group "default" {
    targets = ["lint", "build", "test"]
}

target "lint" {
    dockerfile = "Dockerfile"
    context = "."
    tags = ["lint"]
    target = "lint"
}

target "build" {
    dockerfile = "Dockerfile"
    context = "."
    tags = ["build"]
    target = "build"
}

target "test" {
    dockerfile = "Dockerfile"
    context = "."
    tags = ["test"]
    target = "test"
}
