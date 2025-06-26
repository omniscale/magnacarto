.PHONY: test

test:
	go generate .
	go test .
