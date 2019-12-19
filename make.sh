#!/usr/bin/env sh

VERSION=1.0.1

echo fetch dependencies
go get github.com/Sirupsen/logrus
go get github.com/yosssi/gmq/mqtt
go get github.com/yosssi/gmq/mqtt/client
go get github.com/juju2013/go-freebox
go get github.com/konsorten/go-windows-terminal-sequences

echo build linux/arm/5
mkdir -p release/${VERSION}/linux/arm
GOOS=linux GOARCH=arm GOARM=5 go build mqtt-freebox.go
mv mqtt-freebox release/${VERSION}/linux/arm/mqtt-freebox

echo build linux/amd64
mkdir -p release/${VERSION}/linux/amd64
GOOS=linux GOARCH=amd64 go build mqtt-freebox.go
mv mqtt-freebox release/${VERSION}/linux/amd64/mqtt-freebox

echo build windows/admd64
mkdir -p release/${VERSION}/windows/amd64
GOOS=windows GOARCH=amd64 go build mqtt-freebox.go
mv mqtt-freebox.exe release/${VERSION}/windows/amd64/mqtt-freebox.exe
