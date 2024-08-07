DIST_FOLDER=./bin

build:
	@go build -o $(DIST_FOLDER)/export cmd/export/main.go

run:
	@go run cmd/export/main.go
