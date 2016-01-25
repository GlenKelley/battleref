package main

import (
	"flag"
	"github.com/GlenKelley/battleref/arena"
	"github.com/GlenKelley/battleref/git"
	"github.com/GlenKelley/battleref/server"
	"github.com/GlenKelley/battleref/tournament"
	"log"
)

func main() {
	var environment string
	var resourcePath string
	var clear bool
	flag.StringVar(&environment, "e", "", "environment parameters for application")
	flag.StringVar(&resourcePath, "r", ".", "root directory for resource files")
	flag.BoolVar(&clear, "c", false, "clear all state from the server")
	flag.Parse()
	if environment == "" {
		flag.Usage()
		log.Fatal("You must define a environment")
	}

	if properties, err := server.ReadProperties(environment, resourcePath); err != nil {
		log.Fatal(err)
	} else {
		if clear {
			if err := tournament.RemoveDatabase(properties.DatabaseURL); err != nil {
				log.Fatal(err)
			}
		}
		if webserver, err := CreateServer(properties); err != nil {
			log.Fatal(err)
		} else {
			if clear {
				if err := webserver.Tournament.GitHost.Reset(); err != nil {
					log.Fatal(err)
				}
			}
			categories := webserver.Tournament.ListCategories()
			for _, category := range categories {
				if err := webserver.Tournament.InstallDefaultMaps(properties.ArenaResourcePath(), category); err != nil {
					log.Fatal(err)
				}
			}
			log.Printf("Listening on port %v.", properties.ServerPort)
			log.Fatal(webserver.Serve())
		}
	}
}

func CreateServer(properties server.Properties) (*server.ServerState, error) {
	if database, err := tournament.OpenDatabase(properties.DatabaseURL); err != nil {
		return nil, err
	} else if err := database.MigrateSchema(); err != nil {
		return nil, err
	} else if host, err := git.CreateGitHost(properties.GitServerType, properties.GitServerConf); err != nil {
		return nil, err
	} else if err := host.Validate(); err != nil {
		return nil, err
	} else {
		matchArena := arena.NewArena(properties.ArenaResourcePath())
		remote := git.TempRemote{}
		bootstrap := arena.MinimalBootstrap{properties.ArenaResourcePath()}
		tm := tournament.NewTournament(database, matchArena, bootstrap, host, remote)
		webserver := server.NewServer(tm, properties)
		return webserver, nil
	}

}
