package main

import (
	"log"
	"flag"
	"path/filepath"
	"github.com/GlenKelley/battleref/server"
	"github.com/GlenKelley/battleref/tournament"
	"github.com/GlenKelley/battleref/arena"
)

//TODO:(glen) Validate public keys
//TODO:(glen) upload maps
//TODO:(glen) run checkouts in sandbox

func main() {
	var environment string
	var resetDatabase bool
	var resourcePath string
	flag.StringVar(&environment, "e", "", "environment parameters for application")
	flag.StringVar(&resourcePath, "r", ".", "root directory for resource files")
	flag.BoolVar(&resetDatabase, "d", false, "reset database")
	flag.Parse()
	if environment == "" {
		flag.Usage()
		log.Fatal("You must define a environment")
	}

	properties, err := server.ReadProperties(environment, resourcePath)
	if err != nil { log.Fatal(err) }

	if resetDatabase {
		if err := tournament.ResetDatabase(properties.DatabaseURL); err != nil { log.Fatal(err) }
	}

	database, err := tournament.OpenDatabase(properties.DatabaseURL)
	if err != nil { log.Fatal(err) }

	if err := database.MigrateSchema(); err != nil { log.Fatal(err) }

	arena := arena.NewArena(filepath.Join(resourcePath, properties.ArenaResourcePath), properties.GitURL)
	tournament := tournament.NewTournament(database, arena)

	server := server.NewServer(tournament, properties)
	log.Fatal(server.Serve())
}
