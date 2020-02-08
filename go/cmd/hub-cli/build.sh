CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/hub-cli/hub-cli ./cmd/hub-cli/hub-cli.go

docker build -t blackducksoftware/cerebros/hub-cli:latest ./cmd/hub-cli/.