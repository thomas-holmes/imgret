build:
	go build -o imgret cmd/imgret/main.go

run:
	go run cmd/imgret/main.go

clean:
	rm -f imgret

test:
	go test ./...

vet:
	go vet ./...
