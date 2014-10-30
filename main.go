package main

import (
	"log"
	"flag"
	"github.com/GlenKelley/battleref/server"
	"github.com/GlenKelley/battleref/tournament"
)

//TODO:(glen) Validate public keys
//TODO:(glen) upload maps
//TODO:(glen) run checkouts in sandbox

func main() {
	var propertiesFile string
	var resetDatabase bool
	flag.StringVar(&propertiesFile, "p", "", "environment parameters for application")
	flag.BoolVar(&resetDatabase, "r", false, "reset database")
	flag.Parse()
	if propertiesFile == "" {
		flag.Usage()
		log.Fatal("You must provide a properties file")
	}

	properties, err := server.ReadProperties(propertiesFile)
	if err != nil { log.Fatal(err) }

	if resetDatabase {
		if err := tournament.ResetDatabase(properties.DatabaseURL); err != nil { log.Fatal(err) }
	}

	database, err := tournament.OpenDatabase(properties.DatabaseURL)
	if err != nil { log.Fatal(err) }

	if err := database.MigrateSchema(); err != nil { log.Fatal(err) }

	tournament := tournament.NewTournament(database)

	server := server.NewServer(tournament, properties)
	log.Fatal(server.Serve())
}
