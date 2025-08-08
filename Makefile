MODULE_NAME := myapp
SRC := $(shell find . -name '*.go' 2>/dev/null || echo "")
BUILD_DIR := ./build
BINARY_NAME := $(BUILD_DIR)/$(MODULE_NAME)
SONAR_HOST_URL := https://sonarcloud.io
SONAR_SECRET := $(shell cat .sonar.secret 2>/dev/null || echo "")
BRANCH_NAME := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")
CHANGE_TARGET := $(shell git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null | sed 's/.*\///' || echo "main")
# CHANGE_ID := $(shell git rev-parse --short=8 HEAD)
CHANGE_ID := $(shell whoami 2>/dev/null || echo "unknown")

.PHONY: all docs

build:
	go build -v -ldflags="-X 'version.Version=v1.0.0' -X 'version.GitCommit=$(shell git rev-parse --short=8 HEAD 2>/dev/null || echo "unknown")' -X 'build.User=$(shell id -u -n 2>/dev/null || whoami 2>/dev/null || echo "unknown")' -X 'build.Time=$(shell date 2>/dev/null || echo "unknown")'" -o $(BINARY_NAME)

clean:
	go clean
	-@rm -f $(BINARY_NAME) 2>/dev/null || true

unit-test:
	@echo "Running unit tests"
	rm -f unit-test-coverage.out && \
	go test -v $(shell go list ./... | grep -v /test) \
	-count=1 \
	-cover \
	-coverpkg=./... \
	-coverprofile=./unit-test-coverage.out

api-test:
	@echo "Running api tests"
	rm -f api-test-coverage.out && \
	go test -v ./test/ \
	-count=1 \
	-cover \
	-coverpkg=./... \
	-coverprofile=./api-test-coverage.out

unit-test-xml:
	@echo "Running unit tests"
	rm -f unit-test-report.xml && \
	go test -v 2>&1 $(shell go list ./... | grep -v /test) \
	-count=1 \
	| go-junit-report -set-exit-code > unit-test-report.xml

api-test-xml:
	@echo "Running api tests"
	rm -f api-test-report.xml && \
	go test -v 2>&1 ./test/ \
	-count=1 \
	| go-junit-report -set-exit-code > api-test-report.xml

cov-html:
	go tool cover -html=api-test-coverage.out -html=unit-test-coverage.out -o merged-coverage.html

sonarqube-pr:
	rm -rf .scannerwork && \
	sonar-scanner \
		-Dsonar.host.url="$(SONAR_HOST_URL)" \
		-Dsonar.working.directory=".scannerwork" \
		-Dsonar.pullrequest.key="$(CHANGE_ID)" \
		-Dsonar.pullrequest.branch="$(BRANCH_NAME)" \
		-Dsonar.pullrequest.base="$(CHANGE_TARGET)" \
		-Dsonar.login="$(SONAR_SECRET)"

sonarqube-branch:
	rm -rf .scannerwork && \
	sonar-scanner \
		-Dsonar.host.url="$(SONAR_HOST_URL)" \
		-Dsonar.working.directory=".scannerwork" \
		-Dsonar.branch.name="$(BRANCH_NAME)" \
		-Dsonar.login="$(SONAR_SECRET)"

coverage:
	make unit-test && make api-test

vet:
	go vet $(SRC)

lint:
	go get golang.org/x/lint/golint
	$(GOPATH)/bin/golint ./...

# Docker commands
docker-build: docker-build-dev docker-build-staging docker-build-prod

docker-build-dev:
	docker build --target development -t go-restfull:dev .

docker-build-staging:
	docker build --target staging -t go-restfull:staging .

docker-build-prod:
	docker build --target production -t go-restfull:prod .

docker-run:
	docker run -p 8000:8000 go-restfull:prod

docker-run-dev:
	docker run -p 8000:8000 -v $(PWD):/app go-restfull:dev

docker-run-staging:
	docker run -p 8000:8000 go-restfull:staging

# Docker run with external configuration
docker-run-with-config:
	docker run -p 8000:8000 -v $(PWD)/custom.yml:/app/config.yml go-restfull:prod

docker-run-with-config-env:
	docker run -p 8000:8000 -v $(PWD)/custom.yml:/configs/user.yml -e CONFIG_PATH=/configs/user.yml go-restfull:prod

docker-run-dev-with-config:
	docker run -p 8000:8000 -v $(PWD):/app -v $(PWD)/custom.yml:/app/config.yml go-restfull:dev

docker-run-dev-with-config-env:
	docker run -p 8000:8000 -v $(PWD):/app -v $(PWD)/custom.yml:/configs/user.yml -e CONFIG_PATH=/configs/user.yml go-restfull:dev

# Docker Hub commands
docker-hub-login:
	docker login

docker-hub-build-and-push: docker-hub-build docker-hub-push

docker-hub-build:
	docker build --target production -t turahe/go-restfull:latest .
	docker build --target production -t turahe/go-restfull:prod .
	docker build --target staging -t turahe/go-restfull:staging .
	docker build --target development -t turahe/go-restfull:dev .

docker-hub-push:
	docker push turahe/go-restfull:latest
	docker push turahe/go-restfull:prod
	docker push turahe/go-restfull:staging
	docker push turahe/go-restfull:dev

docker-hub-push-latest:
	docker push turahe/go-restfull:latest

docker-hub-push-prod:
	docker push turahe/go-restfull:prod

docker-hub-push-staging:
	docker push turahe/go-restfull:staging

docker-hub-push-dev:
	docker push turahe/go-restfull:dev

# Development commands
run:
	go run main.go server

migrate:
	go run main.go migrate

seed:
	go run main.go seed

docs:
	@echo "Generating API documentation..."
	swag init -g main.go
	@echo "Documentation generated successfully!"
