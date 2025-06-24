build:
	@go build -o bin/simple-i18n ./cmd/simple-i18n/main.go

test:
	@go test ./...

integration: build
	rm -f ./cmd/test/toml/generated/*
	@./bin/simple-i18n -i ./cmd/test/toml -o ./cmd/test/generated -p i18n -b sv -v
	@go build -o bin/test  ./cmd/test/main.go
	@./bin/test
