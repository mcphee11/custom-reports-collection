#!/bin/bash
#!/bin/bash
    go build -ldflags="-s -w" -o linux/customer-insights \
      && env GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o windowsARM64/customer-insights.exe \
      && env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o windows/customer-insights.exe \
      && env GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o macos/customer-insights \
      && env GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o macosARM64/customer-insights
