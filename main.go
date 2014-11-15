package main

import (
	"log"
	"flag"
	"github.com/GlenKelley/battleref/server"
	"github.com/GlenKelley/battleref/tournament"
	"github.com/GlenKelley/battleref/arena"
	"github.com/GlenKelley/battleref/git"
)

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
	} else if webserver, err := CreateServer(properties); err != nil {
		log.Fatal(err)
	} else if err := webserver.Tournament.InstallDefaultMaps(properties.ArenaResourcePath(), tournament.CategoryGeneral); err != nil {
		log.Fatal(err)
	} else {
		//TODO: Cleanup repo/host
		log.Fatal(webserver.Serve())
	}
}

func CreateServer(properties server.Properties) (*server.ServerState, error) {
	if database, err := tournament.OpenDatabase(properties.DatabaseURL); err != nil {
		return nil, err
	} else if err := database.MigrateSchema(); err != nil {
		return nil, err
	} else if host, err := git.CreateGitHost(properties.GitHost); err != nil {
		return nil, err
	} else {
		matchArena := arena.NewArena(properties.ArenaResourcePath())
		remote := git.TempRemote{}
		bootstrap := arena.MinimalBootstrap{}
		tm := tournament.NewTournament(database, matchArena, bootstrap, host, remote)
		webserver := server.NewServer(tm, properties)
		return webserver, nil
	}

}
