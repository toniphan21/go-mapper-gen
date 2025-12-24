build-go-mapper-gen:
	go build -o bin/go-mapper-gen ./cmd/go-mapper-gen

build-binaries: build-go-mapper-gen
	go build -o bin/generator ./cmd/generator

test-features: build-go-mapper-gen
	./bin/go-mapper-gen test features/*.md testdata/*.md converters/grpc/features/*.md

generate-examples: build-go-mapper-gen
	rm -rf ./examples
	./bin/go-mapper-gen test features/*.md testdata/*.md converters/grpc/features/*.md -e -s

generate-testdata-golden-files: build-go-mapper-gen
	rm -rf ./testdata/golden
	./bin/go-mapper-gen test features/*.md testdata/*.md converters/grpc/features/*.md -e -s
