.PHONY: schema

schema:
	go run github.com/harderthanitneedstobe/rest-api/v0/cmd/schemagen > schema/00_initial.sql

docker:
	GOOS=linux go build -o ./bin/rest-api github.com/harderthanitneedstobe/rest-api/v0/cmd/api
	docker build -t harder-rest-api -f Dockerfile .

clean-development:
	docker-compose -f ./docker-compose.development.yaml rm --stop --force || true

compose-development: schema
	docker-compose  -f ./docker-compose.development.yaml up