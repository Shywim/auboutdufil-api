BINARY="abdf"

.PHONY: all
all: abdf

abdf: main.go
	go build -o ${BINARY} main.go

.PHONY: clean
clean:
	rm -r ${BINARY}
