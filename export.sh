#!/bin/bash
go build -v -o ./bin/tbotopenai ./cmd/gotbotopenai/main.go
./bin/tbotopenai -c ./etc.template/tbotopenai.yaml