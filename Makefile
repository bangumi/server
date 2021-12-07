MAKEFLAGS += --no-builtin-rules

define helpMessage
Building Targets:

  build:

Testing Targets:

  test: simply run tests.
  coverage: test with coverage report './coverage.out'.

Others Targets:

  generate: generated files like protobuf.
  clean: cleanup all auxiliary files.
  install: install required binary

endef
export helpMessage

help:
	@echo "$$helpMessage"

# this is used in github ci with `make ${{ runner.os }}`
build: dist/app.exe

LDFLAGS = -X 'app/pkg/vars.Ref=${REF}'
LDFLAGS += -X 'app/pkg/vars.Commit=${SHA}'
LDFLAGS += -X 'app/pkg/vars.Builder=$(shell go version)'
LDFLAGS += -X 'app/pkg/vars.BuildTime=${TIME}'

GoBuildArgs = -ldflags "-s -w $(LDFLAGS)"

dist/app.exe: generate
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $(GoBuildArgs)

pkg/web/rice-box.go:
	rice embed-go -i ./pkg/web/

./ent/character.go: ./ent/schema/character.go ./.bin/ent.exe
	.bin/ent.exe generate ./ent/schema

generate: ./ent/character.go

test:
	go test -race -v ./...

coverage:
	go test -race -covermode=atomic -coverprofile=coverage.out -count=1 ./...

./.bin/ent.exe:
	go get -d entgo.io/ent/cmd/ent
	go build -o .bin/ent.exe entgo.io/ent/cmd/ent

install: .bin/ent.exe

clean::
	rm -rf ./out
	rm -rf ./dist ./.bin

.PHONY:: install help build test coverage clean generate
