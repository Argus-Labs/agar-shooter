#!/bin/sh

set -o errexit -o nounset

cd nakama
rm -rf vendor/
go mod tidy
go mod vendor
cd ..

cd cardinal
rm -rf vendor/
go mod tidy
go mod vendor
cd ..

if [[ ${1:-} == "build-only" ]] ; then
  docker compose build
  exit 0
fi

docker compose up --build
