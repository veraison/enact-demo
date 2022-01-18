# Prerequisites

## go (1.17.x)

```sh
brew install go
```

## mockgen

```sh
go get github.com/golang/mock/mockgen@v1.5.0
```

## golangci-lint

```sh
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.43.0
```

## protoc (3.x)

```sh
brew install protobuf
```

## protoc-gen

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
go install github.com/mitchellh/protoc-gen-go-json@latest
```

## Caveats

* Make sure that `~/go/bin` is in your shell `PATH`

# Veraison build

**For now, these instructions are limited to the provisioning frontend side only.**

* checkout code 

```sh
git clone https://github.com/veraison/veraison
```

* build the provisioning frontend

```sh
cd veraison/provisioning && make
```

* run the provisioning backend with the TPM/EnactTrust and PSA plugins in

```sh
( cd cmd && ./provisioning )
```

## Caveats

The provisioning frontend wants to run on `localhost:8888` so make sure that the TCP server port is available or change it in the `veraison/provisioning/cmd/main.go` file.

## EnactTrust plugin code

The TPM EnactTrust provisioning plugin resides in the `veraison/provisioning/plugins/corim-tpm-enacttrust-decoder` folder.

The `decoder.go` file is the ultimate source of truth in terms of the supported media type(s).

## Shell tester

There is a `curl(1)`-based test script at `veraison/provisioning/test/req.sh` which can be used for checking the health of the service.
