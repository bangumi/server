MAKEFLAGS += --no-builtin-rules

define helpMessage
Building Targets:

  build:

Testing Targets:

  test: run Tests without e2e and driver.
  test-all: Run all tests, need database and redis.
            You can also run `make test` with env `TEST_MYSQL=1 TEST_REDIS=1`
  bench:    Run all benchmark.
  coverage: Test-all with coverage report './coverage.out'.

Others Targets:

  gen: Generate files like gorm-gen and mocks.
  mocks: Generate mocks.
  clean: Cleanup all auxiliary files.
  install: Install required binary

endef
export helpMessage

help:
	@echo "$$helpMessage"

# this is used in github ci with `make ${{ runner.os }}`
build: ./dist/chii.exe

./dist/chii.exe:
	env CGO_ENABLED=0 go build -o $@

mocks:
	for dir in domain cache; do \
		go run github.com/vektra/mockery/v2 --all --dir $$dir --with-expecter; \
	done

	go run github.com/vektra/mockery/v2 --dir  ./web/captcha --name Manager --filename CaptchaManager.go --structname CaptchaManager --with-expecter;
	go run github.com/vektra/mockery/v2 --dir  ./web/session --name Manager --filename SessionManager.go --structname SessionManager --with-expecter;
	go run github.com/vektra/mockery/v2 --dir  ./web/session --name Repo --filename SessionRepo.go --structname SessionRepo --with-expecter;
	go run github.com/vektra/mockery/v2 --dir  ./web/rate --name Manager --filename RateLimiter.go --structname RateLimiter --with-expecter;


gen: ./dal/query/gen.go mocks

# don't enable `-race` in test because it require cgo, only enable it at coverage.
test:
	go test ./...

test-all: .bin/dotenv.exe
	.bin/dotenv.exe env TEST_MYSQL=1 TEST_REDIS=1 go test ./...

bench:
	go test -bench=. -benchmem ./pkg/wiki ./internal/rand

./dal/query/gen.go: ./internal/cmd/gen/gorm.go internal/cmd/gen/method go.mod .bin/dotenv.exe
	.bin/dotenv.exe go run ./internal/cmd/gen/gorm.go

coverage: .bin/dotenv.exe
	.bin/dotenv.exe env TEST_MYSQL=1 TEST_REDIS=1 go test -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.out -count=1 ./...

.bin/dotenv.exe: go.mod
	go build -o $@ github.com/joho/godotenv/cmd/godotenv

install: .bin/dotenv.exe
	@mkdir -p ./.bin ./tmp

lint:
	golangci-lint run --fix

clean::
	rm -rf ./out
	rm -rf ./dist ./.bin

.PHONY:: install help build test test-all bench coverage clean gen lint mocks
