package main

import (
	"flag"
	"log"
	"github.com/GlenKelley/battleref/server"
	"github.com/GlenKelley/battleref/arena"
)

/**
TODO: validate public keys
TODO: upload maps
TODO: run checkouts in sandbox
*/

func main() {
	var configFilename string
	var cleanDatabase bool
	flag.StringVar(&configFilename, "config", "config.json", "environment parameters for application")
	flag.BoolVar(&cleanDatabase, "clean", false, "clean database")
	flag.Parse()

	config, err := LoadConfig(configFilename)
	if err != nil { log.Fatal(err) }

	dbFile := os.Getenv("BATTLEREF_DB")
	if cleanDatabase {
		os.Remove(dbFile)	
	}
	database, err := OpenDatabase(dbFile)
	if err != nil { log.Fatal(err) }

	err = database.InitTables(config)
	if err != nil { log.Fatal(err) }

	server := NewServer(config, database)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", config.ServerPort), server.Handler))
}

