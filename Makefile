# Copyright 2017 Giuseppe Pellegrino
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO       ?= GO15VENDOREXPERIMENT=1 go
GOPATH   := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
TARGET    = minio_exporter

VERSION             := $(shell cat VERSION)
OS                  := $(shell uname | tr A-Z a-z)
ARCH                := $(shell uname -p)
ARCHIVE             := $(TARGET)-$(VERSION)-$(OS)-$(ARCH).tar.gz

DOCKER_IMAGE_NAME   ?= joepll/minio-exporter
DOCKER_IMAGE_TAG    ?= v$(VERSION)

pkgs := $(shell $(GO) list)

all: format build test

build: get_dep
	@echo "... building binaries"
	@CGO_ENABLED=0 $(GO) build -o $(TARGET) -a -installsuffix cgo $(pkgs)

format:
	@echo "... formatting packages"
	@$(GO) fmt $(pkgs)

test:
	@echo "... testing binary"
	@$(GO) test -short $(pkgs)

get_dep:
	@echo "... getting dependencies"
	@$(GO) get -d

docker: docker_build docker_push
	@echo "... docker building and pushing"

docker_build: build
	@echo "... building docker image"
	@docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

docker_push:
	@echo "... pushing docker image"
	@docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

clean:
	@echo "... cleaning up"
	@rm -rf $(TARGET) $(ARCHIVE) .build

tarball: $(ARCHIVE)
$(ARCHIVE): $(TARGET)
	@echo "... creating tarball"
	@tar -czf $@ $<

.PHONY: all build format test docker_build docker_push tarball clean
