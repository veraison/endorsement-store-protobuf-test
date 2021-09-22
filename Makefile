.PHONY: all
all: build

.PHONY: build
build: build/protogen build/server build/client test/endorsement.db

.PHONY: clean
clean:
	rm -rf build
	rm -f endorsementapi/*.pb.go
	rm -f test/endorsement.db

dir_guard=@mkdir -p $(@D)

build/server: server/*.go
	$(dir_guard)
	go build -o $@ $<

build/client: client/*.go
	$(dir_guard)
	go build -o $@ $<

build/protogen: endorsementapi/endorsementstore.proto
	$(dir_guard)
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative endorsementapi/endorsementstore.proto
	@echo ".pb.go generated in ../endorsementapi/" > build/protogen

test/endorsement.db: test/endorsement.sqlite
	sqlite3 $@ < $<
