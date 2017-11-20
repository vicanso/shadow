#!/bin/sh

GOOS=linux go build
docker build -t vicanso/shadow .
rm ./shadow
