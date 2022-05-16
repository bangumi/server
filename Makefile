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

# we should use gomock once https://github.com/golang/mock/issues/622 is resolved.
mocks: web/session/repo_mock_test.go mocks/SessionManager.go mocks/CaptchaManager.go mocks/RateLimiter.go
	for dir in domain cache; do \
		mockery --all --dir $$dir --with-expecter; \
	done

web/session/repo_mock_test.go: web/session/repo.go
	mockery --inpackage --dir ./web/session --testonly --name Repo --filename repo_mock_test.go --structname MockRepo --with-expecter;

mocks/SessionManager.go: web/session/manager.go
	mockery --dir ./web/session --name Manager --filename SessionManager.go --structname SessionManager --with-expecter;

mocks/CaptchaManager.go: web/captcha/manager.go
	mockery --dir ./web/captcha --name Manager --filename CaptchaManager.go --structname CaptchaManager --with-expecter;

mocks/RateLimiter.go: web/rate/new.go
	mockery --dir ./web/rate --name Manager --filename RateLimiter.go --structname RateLimiter --with-expecter;

gen: ./dal/query/gen.go mocks

# don't enable `-race` in test because it require cgo, only enable it at coverage.
test: .bin/gotestfmt.exe
	go test -json -tags test ./... 2>&1 | .bin/gotestfmt.exe -hide empty-packages,successful-packages

test-all: .bin/dotenv.exe .bin/gotestfmt.exe
	.bin/dotenv.exe env TEST_MYSQL=1 TEST_REDIS=1 go test -json -tags test ./... 2>&1 | .bin/gotestfmt.exe -hide empty-packages,successful-packages

bench:
	go test -bench=. -benchmem ./pkg/wiki

./dal/query/gen.go: ./internal/cmd/gen/gorm.go go.mod .bin/dotenv.exe
	.bin/dotenv.exe go run ./internal/cmd/gen/gorm.go

coverage: .bin/dotenv.exe .bin/gotestfmt.exe
	.bin/dotenv.exe env TEST_MYSQL=1 TEST_REDIS=1 go test -json -tags test -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.out -count=1 ./... 2>&1 | .bin/gotestfmt.exe -hide empty-packages

.bin/gotestfmt.exe: go.mod
	go build -o $@ github.com/haveyoudebuggedit/gotestfmt/v2/cmd/gotestfmt

.bin/dotenv.exe: go.mod
	go build -o $@ github.com/joho/godotenv/cmd/godotenv

install: .bin/dotenv.exe .bin/gotestfmt.exe
	@mkdir -p ./.bin ./tmp

lint:
	golangci-lint run --fix

clean::
	rm -rf ./out
	rm -rf ./dist ./.bin

.PHONY:: install help build test test-all bench coverage clean gen lint mocks
