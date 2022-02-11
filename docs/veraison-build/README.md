# Veraison EnactTrust Integration Diary

## Prerequisites

### go (1.17.x)

```sh
brew install go
```

### mockgen

```sh
go get github.com/golang/mock/mockgen@v1.5.0
```

### golangci-lint

```sh
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.43.0
```

### protoc (3.x)

```sh
brew install protobuf
```

### protoc-gen

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
go install github.com/mitchellh/protoc-gen-go-json@latest
```

### Caveats

* Make sure that `~/go/bin` is in your shell `PATH`

## Veraison build

**For now, these instructions are limited to the provisioning pipeline only.**

* checkout code and move to the project's root

```sh
git clone -b provisioning-part2 https://github.com/veraison/veraison veraison-integration-test && cd $_
```

* build the provisioning pipeline

```sh
for d in provisioning vts-temp ; do make -C $d ; done
```

* initialise the key-val store (once)

```sh
( cd vts-temp/cmd && ./init-kvstores.sh )
```

* In one shell, run the provisioning frontend:

```sh
( cd provisioning/cmd && ./provisioning-api )
```

* In another shell (assuming the same CWD as above), run the provisioning backend:

```sh
( cd vts-temp/cmd && ./vts-temp )
```

If there are no errors, the provisioning pipeline should be ready for testing.

### Caveats

The provisioning frontend listens on `localhost:8888` so make sure that the TCP server port is available or change it in the `provisioning/cmd/main.go` file and recompile.

The provisioning backend listens on `127.0.0.1:12345`, so the same applies (the file to modify in this case is `vts-temp/cmd/main.go`).

### EnactTrust plugin code

The TPM EnactTrust provisioning plugin resides in the `provisioning/plugins/corim-tpm-enacttrust-decoder` folder.

The `decoder.go` file is the ultimate source of truth in terms of the supported media type(s).

## Shell tester

There is a `curl(1)`-based test script in the `provisioning/test` folder which can be used for checking the health of the service:

```sh
cd provisioning/test
```

* Provision an EnactTrust AK and check the store is populated accordingly (you need to install `jq(1)`):

```sh
B=trustanchor T=tpm-enacttrust ./req.sh

sqlite3 ../../vts-temp/cmd/veraison-trustanchor.sql 'SELECT DISTINCT val FROM trustanchor'  | jq .
```

* Provision an EnactTrust golden value, and check the store is populated accordingly:

```sh
B=refvalue T=tpm-enacttrust ./req.sh

sqlite3 ../../vts-temp/cmd/veraison-endorsement.sql 'SELECT DISTINCT val FROM endorsement'  | jq .
```
