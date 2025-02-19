test:
	GO_ENV=test go test ./...
run:
	@go run main.go -- perform -m demo -c ./gobackup_test.yml
start:
	@go run main.go -- start --config ./gobackup_test.yml -d
release:
	@rm -Rf dist/
	@goreleaser --skip-validate
