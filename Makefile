all: test vet fmt lint

.PHONY: test
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

.PHONY: build
build:
	go build -o bin/pathtracer ./cmd/pathtracer

.PHONY: build_scene
build_scene: build
	@if [ -z "$(SCENE_NAME)" ]; then  \
		echo "You need to set SCENE_NAME parameter with a scene name before calling make."; \
		exit 2; \
	else \
	   	if [ -d "./cmd/scene/$(SCENE_NAME)" ]; then \
			echo "Making target for scene $(SCENE_NAME)"; \
			go build -o bin/$(SCENE_NAME) ./cmd/scene/$(SCENE_NAME); \
		elif [ -d "./cmd/scene/test/$(SCENE_NAME)" ]; then \
			echo "Making target for test scene $(SCENE_NAME)"; \
			go build -o bin/$(SCENE_NAME) ./cmd/scene/test/$(SCENE_NAME); \
		else \
			echo "Could not find any scene nor test scene directory for $(SCENE_NAME)"; \
			exit 2; \
		fi \
	fi
