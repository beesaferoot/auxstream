
setup-db:
	psql -f ./db/setup.sql -d ${POSTGRES_DB} -U ${POSTGRES_USER}

teardown-db:
	psql -f ./db/teardown.sql -d ${POSTGRES_DB} -U ${POSTGRES_USER}

test: 
	go test -v ./tests

run:
	go run main.go