test:
	@go test -v ./...


cover:
	@go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

coverage:
	@go test -cover ./...

build_cli:
	@go build -o ../ladiwork ./cmd/cli


build:
	@go build -o ./dist/cli ./cmd/cli
