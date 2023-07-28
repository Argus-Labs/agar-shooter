#!/bin/sh

set -o errexit -o nounset

cd nakama
rm -rf vendor/
go mod tidy
go mod vendor
cd ..

cd cardinal
go mod tidy
cd ..

docker compose up --build
