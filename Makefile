postgres:
	docker run --name postgres14-auxstream -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14-alpine

createdb:
	docker exec -it postgres14-auxstream createdb --username=root --owner=root auxstreamdb

init-migration-schema:
	go run cmd/migration/main.go init 
	
setup-db:
	go run cmd/migration/main.go up 

rollback-db:
	go run cmd/migration/main.go down

test: 
	go test -v ./tests -coverpkg=./...

run:
	go run main.go


build:
	go build -o build/auxstream

.PHONY: test run createdb setup-db teardown-db