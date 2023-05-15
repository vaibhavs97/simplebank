postgres:
	docker run --name postgres15 -p 5432:5432 -e POSTGRES_PASSWORD=password -d postgres:15-alpine

up:
	docker start postgres15

down:
	docker stop postgres15

rm:
	docker rm postgres15

createdb:
	docker exec -it postgres15 createdb -U postgres -O postgres simple_bank

dropdb:
	docker exec -it postgres15 dropdb -U postgres simple_bank

migrateup:
	migrate -path db/migrations -database "postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migrations -database "postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres up down rm createdb dropdb migrateup migratedown sqlc test