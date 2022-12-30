# Config

VERSION=$(version)


# Build

.PHONY: FORCE

build: go-build
.PHONY: build

clean: go-clean
.PHONY: clean

docker: go-docker
.PHONY: docker

lint: go-lint
.PHONY: lint

test: go-test
.PHONY: test


# Non-PHONY targets (real files)

go-build: FORCE
	./script/build.sh $(VERSION)

go-clean: FORCE
	./script/clean.sh

go-docker: FORCE
	./script/docker.sh $(VERSION)

go-lint: FORCE
	./script/lint.sh

go-test: FORCE
	./script/test.sh report
