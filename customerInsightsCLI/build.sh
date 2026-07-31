#!/bin/bash
go build -o linux/customer-insights && env GOOS=windows GOARCH=arm64 go build -o windowsARM64/customer-insights.exe && env GOOS=windows GOARCH=amd64 go build -o windows/customer-insights.exe && env GOOS=darwin GOARCH=amd64 go build -o macos/customer-insights
