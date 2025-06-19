# Build CLI
build:
	@go build -o bin/simple-i18n ./main.go

# Test integration
test: build
	@./bin/simple-i18n -i ./cmd/test/toml -o ./cmd/test/generated -p i18n -b sv
	@go build -o bin/test  ./cmd/test/main.go
	@./bin/test
