# Build CLI
build:
	@go build -o bin/simple-i18n ./cmd/cli/main.go

# Test integration
test: build
	@./bin/simple-i18n -i ./cmd/test/toml -o ./cmd/test/i18n
	@go build -o bin/test  ./cmd/test/main.go
	@./bin/test
