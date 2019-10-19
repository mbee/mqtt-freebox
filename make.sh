#!/usr/bin/env sh

VERSION=1.0.0

echo fetch dependencies
go get github.com/Sirupsen/logrus
go get github.com/juju2013/go-freebox
go get github.com/yosssi/gmq/mqtt
go get github.com/yosssi/gmq/mqtt/client

echo build linux/arm/5
mkdir -p release/1.0.0/linux/arm
GOOS=linux GOARCH=arm GOARM=5 go build -o release/1.0.0/linux/arm/mqtt-freebox

echo build linux/amd64
mkdir -p release/1.0.0/linux/amd64
GOOS=linux GOARCH=amd64 go build -o release/1.0.0/linux/amd64/mqtt-freebox
