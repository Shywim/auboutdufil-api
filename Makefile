BINARY="abdf"

.PHONY: all
all: abdf

abdf: main.go parser.go
	go build .

.PHONY: clean
clean:
	rm -r ${BINARY}
