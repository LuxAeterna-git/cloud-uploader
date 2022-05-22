.PHONY: gen-mocks
gen-mocks: gen-mocks-gcp gen-mocks-azure gen-mocks-aws

.PHONY: test
test: gen-mocks
	go test -coverprofile "" ./...

.PHONY: gen-mocks-gcp
gen-mocks-gcp:
	docker run -v `pwd`:/src -w /src vektra/mockery:v2.9 --case snake --dir ./gcp --output mocks/gcp --outpkg mocks --all --exported

.PHONY: gen-mocks-azure
gen-mocks-azure:
	docker run -v `pwd`:/src -w /src vektra/mockery:v2.9 --case snake --dir ./azure --output mocks/azure --outpkg mocks --all --exported

.PHONY: gen-mocks-aws
gen-mocks-aws:
	docker run -v `pwd`:/src -w /src vektra/mockery:v2.9 --case snake --dir ./aws --output mocks/aws --outpkg mocks --all --exported
