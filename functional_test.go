package main

import (
	"time"
	"testing"
	"github.com/GlenKelley/battleref/testing"
	"github.com/GlenKelley/battleref/tournament"
	"github.com/GlenKelley/battleref/server"
	"github.com/GlenKelley/battleref/web"
	"github.com/GlenKelley/battleref/git"
)

type JSONBody map[string]interface{}

func TestInitToRun(test *testing.T) {
	t := (*testutil.T)(test)
	if webserver, err := CreateServer(server.Properties{
		":memory:",
		"8080",
		":temp:",
		nil,
		".",
	}); err != nil {
		t.FailNow()
	} else {
		defer webserver.Tournament.GitHost.Cleanup()
		if err := webserver.Tournament.InstallDefaultMaps(webserver.Properties.ArenaResourcePath(), tournament.CategoryBattlecode2014); err != nil {
			t.ErrorNow(err)
		}
		go webserver.Serve()
		//Race condition of server not starting
		time.Sleep(time.Millisecond)

		if commit, err := CreatePlayer("playerFoo"); err != nil {
			t.ErrorNow(err)
		} else if maps, err := GetMaps(); err != nil {
			t.ErrorNow(err)
		} else if len(maps) == 0 {
			t.ErrorNowf("No default maps")
		} else {
			t.CheckError(RunMatch("playerFoo","playerFoo", commit, commit, maps[0]))
		}

	}
}

func CreatePlayer(name string) (string, error) {
	var response struct {
		Data struct {
			CommitHash string `json:"commit_hash"`
			RepoURL string `json:"repo_url"`
		} `json:"data"`
	}
	if _, pubKey, err := testutil.CreateKeyPair(); err != nil {
		return "", err
	} else if err := web.SendPostJson("http://localhost:8080/register", web.JsonBody{"name":name, "public_key":pubKey}, &response); err != nil {
		return "", err
	} else if repo, err := (git.TempRemote{}).CheckoutRepository(response.Data.RepoURL); err != nil {
		return "", err
	} else {
		defer repo.Delete()
		return response.Data.CommitHash, nil
	}
}

func GetMaps() ([]string, error) {
	var response struct {
		Data struct {
			Maps []string `json:"maps"`
		} `json:"data"`
	}
	if err := web.SendGetJson("http://localhost:8080/maps", web.JsonBody{}, &response); err != nil {
		return nil, err
	} else {
		return response.Data.Maps, nil
	}
}

func RunMatch(name, name2, commit, commit2, mapName string) error {
	response := struct{}{}
	if err := web.SendPostJson("http://localhost:8080/match/run", web.JsonBody{
		"player1":name,
		"player2":name2,
		"commit1":commit,
		"commit2":commit2,
		"category":tournament.CategoryBattlecode2014,
		"map":mapName,
	}, &response); err != nil {
		return err
	} else {
		return nil
	}
}


