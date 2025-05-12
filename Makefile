go_xx:
	go build -o chain-account.exe

clean:
	del /Q chain-account.exe

test:
	go test -v ./...

lint:
	golangci-lint run ./...

.PHONY: \
	chain-account \
	clean \
	test \
	lint