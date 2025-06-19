build:
	@go build -o bin/simple-i18n ./cmd/simple-i18n/main.go

# Test integration
test: build
	rm -f ./cmd/test/toml/generated/*
	@./bin/simple-i18n -i ./cmd/test/toml -o ./cmd/test/generated -p i18n -b sv
	@go build -o bin/test  ./cmd/test/main.go
	@./bin/test
