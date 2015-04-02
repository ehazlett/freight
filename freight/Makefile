CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
TAG=${TAG:-latest}
OS="darwin windows linux"
ARCH="amd64 386"
COMMIT=`git rev-parse --short HEAD`

all: build

clean:
	@rm -rf freight_*

build:
	@godep go build -a -tags 'netgo' -ldflags "-w -X github.com/ehazlett/freight/version.Gitcommit $(COMMIT) -linkmode external -extldflags -static" .

build-cross:
	@gox -os=$(OS) -arch=$(ARCH) -ldflags "-w -X github.com/ehazlett/freight/version.Gitcommit $(COMMIT)" -output="freight_{{.OS}}_{{.Arch}}"

image: build
	@echo Building image $(TAG)
	@docker build -t ehazlett/freight:$(TAG) .

release: build image
	@docker push ehazlett/freight:$(TAG)


.PHONY: all deps build clean image test release
