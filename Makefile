all: test vet fmt lint build_all

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

build_all: build animations

animations: build_sphere_rotation build_sphere_rotation_focaldistance build_cornellbox build_cylindrical_projection

# Build animation scenes
# -----------------------------------

build_sphere_rotation: build
	go build -o bin/animation_sphere_circle_rotation ./cmd/animation_sphere_circle_rotation

build_sphere_rotation_focaldistance: build
	go build -o bin/animation_sphere_circle_rotation_focaldistance ./cmd/animation_sphere_circle_rotation_focaldistance

build_cornellbox: build
	go build -o bin/cornellbox ./cmd/cornellbox

build_cylindrical_projection: build
	go build -o bin/cylindrical_projection ./cmd/cylindrical_projection


