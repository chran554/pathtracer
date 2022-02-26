all: test vet fmt lint build

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l
	test -z $$(go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l)

lint:
	go list ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

build:
	go build -o bin/pathtracer ./cmd/pathtracer
	go build -o bin/animation_sphere_circle_rotation ./cmd/animation_sphere_circle_rotation
	go build -o bin/animation_sphere_circle_rotation_focaldistance ./cmd/animation_sphere_circle_rotation_focaldistance
	go build -o bin/cornellbox ./cmd/cornellbox
