.PHONY: build test install clean

build:
	go build -o ccgears ./cmd/ccgears

test:
	go test ./internal/... -v

install:
	go install ./cmd/ccgears

clean:
	rm -f ccgears
