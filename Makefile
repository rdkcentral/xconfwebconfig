#
# Copyright 2022 Comcast Cable Communications Management, LLC
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
#
# SPDX-License-Identifier: Apache-2.0
#
GOARCH = $(shell go env GOARCH)
GOOS = $(shell go env GOOS)
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
# must be "Version", NOT "VERSION" to be consistent with xpc jenkins env
Version ?= $(shell git log -1 --pretty=format:"%h")
BUILDTIME := $(shell date -u +"%F_%T_%Z")
GRAVBIN := xconfwebconfig-grav

all: build

build:  ## Build a version
	go build -v -ldflags="-X xconfwebconfig/common.BinaryBranch=${BRANCH} -X xconfwebconfig/common.BinaryVersion=${Version} -X xconfwebconfig/common.BinaryBuildTime=${BUILDTIME}" -o bin/xconfwebconfig-${GOOS}-${GOARCH} main.go

linux:
	GOOS=linux go build -v -ldflags="-X xconfwebconfig/common.BinaryBranch=${BRANCH} -X xconfwebconfig/common.BinaryVersion=${Version} -X xconfwebconfig/common.BinaryBuildTime=${BUILDTIME}" -o bin/xconfwebconfig-linux-amd64 main.go

test:
	ulimit -n 10000 ; go test ./... -cover -count=1

cover:
	go test ./... -count=1 -coverprofile=coverage.out

html:
	go tool cover -html=coverage.out

clean: ## Remove temporary files
	go clean

release:
	go build -v -ldflags="-X xconfwebconfig/common.BinaryBranch=${BRANCH} -X xconfwebconfig/common.BinaryVersion=${Version} -X xconfwebconfig/common.BinaryBuildTime=${BUILDTIME}" -o bin/xconfwebconfig-${GOOS}-${GOARCH} main.go

grav:
	@echo "Building Graviton binaries"
	export GOOS=linux ; export GOARCH=arm64 ; export CGO_ENABLED=0 ; go build -v -ldflags="-X ${REPO}/common.BinaryBranch=${BRANCH} -X ${REPO}/common.BinaryVersion=${Version} -X ${REPO}/common.BinaryBuildTime=${BUILDTIME}" -o bin/${GRAVBIN} main.go
