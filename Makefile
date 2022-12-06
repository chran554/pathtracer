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

install-lint-revive:
	go install github.com/mgechev/revive@latest

build:
	go build -o bin/pathtracer ./cmd/pathtracer

build_all: build animations

animations: build_sphere_circle_rotation \
	build_sphere_rotation_focaldistance \
	build_cornellbox \
	build_spherical_projection \
	build_cylindrical_projection \
	build_parallel_projection \
	build_recursive_spheres \
	build_facetobj_test \
	build_objectfile_test \
	build_gordian_knot \
	build_diamondsR4ever \
	build_dop_test \
	build_primitive_display \
	build_aperture_shape_test \
	build_aperture_shape_test2 \
	build_ply_file_test \
	build_window_test \
	build_lamp_post

# Build animation scenes
# -----------------------------------

build_sphere_circle_rotation: build
	go build -o bin/sphere_circle_rotation ./cmd/scene/sphere_circle_rotation

build_sphere_rotation_focaldistance: build
	go build -o bin/animation_sphere_circle_rotation_focaldistance ./cmd/scene/animation_sphere_circle_rotation_focaldistance

build_cornellbox: build
	go build -o bin/cornellbox ./cmd/scene/cornellbox

build_cylindrical_projection: build
	go build -o bin/cylindrical_projection ./cmd/scene/test/cylindrical_projection

build_spherical_projection: build
	go build -o bin/spherical_projection ./cmd/scene/test/spherical_projection

build_parallel_projection: build
	go build -o bin/parallel_projection ./cmd/scene/test/parallel_projection

build_reflective_test: build
	go build -o bin/reflective_test ./cmd/scene/test/reflective_test

build_refraction_test: build
	go build -o bin/refraction_test ./cmd/scene/test/refraction_test

build_recursive_spheres: build
	go build -o bin/recursive_spheres ./cmd/scene/recursive_spheres

build_facetobj_test: build
	go build -o bin/facetobj_test ./cmd/scene/test/facetobj_test

build_objectfile_test: build
	go build -o bin/objectfile_test ./cmd/scene/test/objectfile_test

build_gordian_knot: build
	go build -o bin/gordian_knot ./cmd/scene/gordian_knot

build_diamondsR4ever: build
	go build -o bin/diamondsR4ever ./cmd/scene/diamondsR4ever

build_dop_test: build
	go build -o bin/dop_test ./cmd/scene/test/dop_test

build_primitive_display: build
	go build -o bin/primitive_display ./cmd/scene/primitive_display

build_aperture_shape_test: build
	go build -o bin/aperture_shape_test ./cmd/scene/test/aperture_shape_test

build_aperture_shape_test2: build
	go build -o bin/aperture_shape_test2 ./cmd/scene/test/aperture_shape_test2

build_ply_file_test: build
	go build -o bin/ply_file_test ./cmd/scene/test/ply_file_test

build_window_test: build
	go build -o bin/window_test ./cmd/scene/test/window_test

build_lamp_post: build
	go build -o bin/lamp_post ./cmd/scene/lamp_post
