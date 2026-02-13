@task
func build()
    $go build -o bin/rage cmd/rage/main.go

@task
func tst()
    $go test ./...

@task
func rage()
    $go run ./cmd/rage/main.go

@task
func integration()
    $go run ./test/integration/integration_test_runner.go
