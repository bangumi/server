MAKEFLAGS += --no-builtin-rules

define helpMessage
Building Targets:

  build:

Testing Targets:

  test: run Tests without e2e and driver.
  test-all: Run all tests, need database and redis.
            You can also run `make test` with env `TEST_MYSQL=1 TEST_REDIS=1`
  bench:    Run all benchmark
  coverage: Test-all with coverage report './coverage.out'.

Others Targets:

  gen: generated files like protobuf.
  clean: cleanup all auxiliary files.
  install: install required binary

endef
export helpMessage

help:
	@echo "$$helpMessage"

# this is used in github ci with `make ${{ runner.os }}`
build: ./dist/chii.exe

./dist/chii.exe:
	env CGO_ENABLED=0 go build -o $@

mocks: .bin/mockery.exe
	.bin/mockery.exe --all --dir domain --inpackage --with-expecter
	.bin/mockery.exe --all --dir cache --inpackage --with-expecter

gen: ./dal/query/gen.go mocks

# don't enable `-race` in test because it require cgo, only enable it at coverage.
test: .bin/dotenv.exe
	.bin/dotenv.exe go test ./...

test-all: .bin/dotenv.exe
	.bin/dotenv.exe env TEST_MYSQL=1 TEST_REDIS=1 go test ./...

bench:
	go test -bench=. -benchmem ./pkg/wiki

./dal/query/gen.go: ./internal/cmd/gen/gorm.go internal/cmd/gen/method go.mod .bin/dotenv.exe
	.bin/dotenv.exe go run ./internal/cmd/gen/gorm.go

coverage: .bin/dotenv.exe
	.bin/dotenv.exe env TEST_MYSQL=1 TEST_REDIS=1 go test -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.out -count=1 ./...

.bin/dotenv.exe: go.mod
	go build -o $@ github.com/joho/godotenv/cmd/godotenv

.bin/mockery.exe: go.mod
	go build -o $@ github.com/vektra/mockery/v2

install: .bin/mockery.exe .bin/dotenv.exe
	@mkdir -p ./.bin ./tmp
	go get ./...

lint:
	golangci-lint run --fix

clean::
	rm -rf ./out
	rm -rf ./dist ./.bin

.PHONY:: install help build test test-all bench coverage clean gen lint mocks
