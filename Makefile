TEST?=$$(go list ./... | grep -v /vendor/)
NAME?=$(shell basename "$(CURDIR)")
VERSION = $(shell awk -F\" '/^const Version/ { print $$2; exit }' version.go)

default: dev

# bin generates the releaseable binaries for Vault
bin: generate
	@sh -c "'$(CURDIR)/scripts/build.sh'"

# dev creates binaries for testing Vault locally. These are put
# into ./bin/ as well as $GOPATH/bin
dev: generate
	@DEV=1 sh -c "'$(CURDIR)/scripts/build.sh'"

# dist creates the binaries for distibution
dist: bin
	@sh -c "'$(CURDIR)/scripts/dist.sh' $(VERSION)"

# test runs the unit tests and vets the code
test: generate
	TF_ACC= go test $(TEST) $(TESTARGS) -timeout=30s -parallel=4

# testacc runs acceptance tests
testacc: generate
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package"; \
		exit 1; \
	fi
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 45m

# testrace runs the race checker
testrace: generate
	TF_ACC= go test -race $(TEST) $(TESTARGS)

# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@go list -f '{{.Dir}}' ./... | grep -v /vendor/) \
		| grep -v '.*github.com/hashicorp/vault-ssh-helper$$' \
		| xargs go tool vet ; if [ $$? -eq 1 ]; then \
			echo ""; \
			echo "Vet found suspicious constructs. Please check the reported constructs"; \
			echo "and fix them if necessary before submitting the code for reviewal."; \
		fi

# generate runs `go generate` to build the dynamically generated
# source files.
generate:
	go generate $(go list ./... | grep -v /vendor/)

# updatedeps installs all the dependencies needed to run and build - this is
# specifically designed to only pull deps, but not self.
updatedeps:
	GO111MODULE=off go get -u github.com/mitchellh/gox
	echo $$(go list ./... | grep -v /vendor/) \
		| xargs go list -f '{{ join .Deps "\n" }}{{ printf "\n" }}{{ join .TestImports "\n" }}' \
		| grep -v github.com/hashicorp/$(NAME) \
		| xargs go get -f -u -v

install: dev
	@sudo cp bin/vault-ssh-helper /usr/local/bin

.PHONY: default bin dev dist generate test vet updatedeps testacc install
