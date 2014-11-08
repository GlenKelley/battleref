package main

import (
	"log"
	"flag"
	"github.com/GlenKelley/battleref/server"
	"github.com/GlenKelley/battleref/tournament"
	"github.com/GlenKelley/battleref/arena"
	"github.com/GlenKelley/battleref/git"
)

//TODO:(glen) Validate public keys
//TODO:(glen) upload maps
//TODO:(glen) run checkouts in sandbox

func main() {
	var environment string
	var resourcePath string
	flag.StringVar(&environment, "e", "", "environment parameters for application")
	flag.StringVar(&resourcePath, "r", ".", "root directory for resource files")
	flag.Parse()
	if environment == "" {
		flag.Usage()
		log.Fatal("You must define a environment")
	}

	if properties, err := server.ReadProperties(environment, resourcePath); err != nil {
		log.Fatal(err)
	} else if database, err := tournament.OpenDatabase(properties.DatabaseURL); err != nil {
		log.Fatal(err)
	} else if err := database.MigrateSchema(); err != nil {
		log.Fatal(err)
	} else if host, err := git.CreateGitHost(properties.GitHost); err != nil {
		log.Fatal(err)
	} else {
		defer host.Cleanup()
		matchArena := arena.NewArena(properties.ArenaResourcePath())
		remote := git.TempRemote{}
		bootstrap := arena.MinimalBootstrap{}
		tournament := tournament.NewTournament(database, matchArena, bootstrap, host, remote)
		webServer := server.NewServer(tournament, properties)
		log.Fatal(webServer.Serve())
	}
}
