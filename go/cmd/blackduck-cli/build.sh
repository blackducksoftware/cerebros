CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/blackduck-cli/blackduck-cli ./cmd/blackduck-cli/blackduck-cli.go

docker build -t blackducksoftware/cerebros/blackduck-cli:latest ./cmd/blackduck-cli/.