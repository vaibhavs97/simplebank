package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/vaibhavs97/simplebank/api"
	db "github.com/vaibhavs97/simplebank/db/sqlc"
	"github.com/vaibhavs97/simplebank/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	if err := server.Start(config.ServerAddress); err != nil {
		log.Fatal("cannot start server", err)
	}
}
