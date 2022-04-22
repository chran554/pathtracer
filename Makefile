all: test vet fmt lint build_all

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l
	test -z $$(go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l)

lint:
	go list ./... | grep -v /vendor/ | xargs -L1 revive -set_exit_status
	# go list ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

build:
	go build -o bin/pathtracer ./cmd/pathtracer

build_all: build animations

animations: build_sphere_circle_rotation build_sphere_rotation_focaldistance build_cornellbox build_spherical_projection build_cylindrical_projection build_parallel_projection build_recursive_spheres

# Build animation scenes
# -----------------------------------

build_sphere_circle_rotation: build
	go build -o bin/sphere_circle_rotation ./cmd/sphere_circle_rotation

build_sphere_rotation_focaldistance: build
	go build -o bin/animation_sphere_circle_rotation_focaldistance ./cmd/animation_sphere_circle_rotation_focaldistance

build_cornellbox: build
	go build -o bin/cornellbox ./cmd/cornellbox

build_cylindrical_projection: build
	go build -o bin/cylindrical_projection ./cmd/cylindrical_projection

build_spherical_projection: build
	go build -o bin/spherical_projection ./cmd/spherical_projection

build_parallel_projection: build
	go build -o bin/parallel_projection ./cmd/parallel_projection

build_reflective_test: build
	go build -o bin/reflective_test ./cmd/reflective_test

build_refraction_test: build
	go build -o bin/refraction_test ./cmd/refraction_test

build_recursive_spheres: build
	go build -o bin/recursive_spheres ./cmd/recursive_spheres


