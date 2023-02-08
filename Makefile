# Copyright 2023 Richard Kosegi
#
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

IMAGE_NAME  := "rkosegi/open-meteo-exporter"
IMAGE_TAG   := "1.0.0"
VER_CURRENT := $(shell git tag --sort=v:refname | tail -1 | sed -Ee 's/^v|-.*//')
VER_PARTS   := $(subst ., ,$(VER_CURRENT))
VER_MAJOR	:= $(word 1,$(VER_PARTS))
VER_MINOR   := $(word 2,$(VER_PARTS))
VER_PATCH   := $(word 3,$(VER_PARTS))
VER_NEXT    := $(VER_MAJOR).$(VER_MINOR).$(shell echo $$(($(VER_PATCH)+1)))

bump-version:
	@echo Current: $(VER_CURRENT)
	@echo Next: $(VER_NEXT)
	sed -i 's/^IMAGE_TAG   := .*/IMAGE_TAG   := "$(VER_NEXT)"/g' Makefile
	git add Makefile
	git commit -sm "Bump version to $(VER_NEXT)"

tag-version:
	git tag -am "Release $(VER_NEXT)" v$(VER_NEXT)

build-docker:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push-docker:
	docker push $(IMAGE_NAME):$(IMAGE_TAG)

build-local:
	go fmt
	go mod download
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o exporter . ; strip exporter

release: build-docker push-docker bump-version

