package tournament

import (
	"os"
	"time"
	"io/ioutil"
	"testing"
	_ "github.com/mattn/go-sqlite3"
	"github.com/GlenKelley/battleref/arena"
	"github.com/GlenKelley/battleref/git"
	"github.com/GlenKelley/battleref/testing"
)

func TournamentTest(test * testing.T, f func(*testutil.T, *Tournament)) {
	t := (*testutil.T)(test)
	if tempDir, err := ioutil.TempDir("", "battleref_test_git_repo"); err != nil {
		t.ErrorNow(err)
	} else {
		defer os.RemoveAll(tempDir)
		gitHost := git.NewLocalDirHost(tempDir)
		dummyArena := arena.DummyArena{}
		remote := &git.TempRemote{}
		bootstrap := &arena.MinimalBootstrap{}
		if database, err := NewInMemoryDatabase(); err != nil {
			t.ErrorNow(err)
		} else if err = database.MigrateSchema(); err != nil {
			t.ErrorNow(err)
		} else {
			tournament := NewTournament(database, dummyArena, bootstrap, gitHost, remote)
			f(t, tournament)
		}
	}
}

func TestCreateUser(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		if isUser, err := tm.UserExists("NameFoo"); err != nil {
			t.ErrorNow(err)
		} else if isUser {
			t.FailNow()
		}
		if users, err := tm.ListUsers(); err != nil {
			t.ErrorNow(err)
		} else if len(users) != 0 {
			t.FailNow()
		}
		t.CheckError(tm.CreateUser("NameFoo", "PublicKey"))
		isUser, err := tm.UserExists("NameFoo")
		if err != nil { t.ErrorNow(err) }
		if !isUser { t.FailNow() }
		users, err := tm.ListUsers()
		if err != nil { t.ErrorNow(err) }
		if len(users) != 1 { t.FailNow() }
		if users[0] != "NameFoo" { t.FailNow() }
	})
}

func TestCreateExistingUserError(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		t.CheckError(tm.CreateUser("NameFoo", "PublicKeyFoo"))
		if err := tm.CreateUser("NameFoo", "PublicKeyFoo"); err == nil {
			t.ErrorNow("expected error")
		}
	})
}

func TestCreateMap(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		t.CheckError(tm.CreateMap("MapFoo", "MapString"))
		if maps, err := tm.ListMaps(); err != nil {
			t.ErrorNow(err)
		} else if len(maps) != 1 {
			t.ErrorNow(len(maps), "but expected", 1)
		} else if maps[0] != "MapFoo" {
			t.ErrorNow(maps[0], "but expected", "MapFoo")
		}
		if mapSource, err := tm.GetMapSource("MapFoo"); err != nil {
			t.ErrorNow(err)
		} else if mapSource != "MapString" {
			t.ErrorNow(mapSource, "but expected", "MapString")
		}
	})
}

func TestCreateExistingMapError(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		t.CheckError(tm.CreateMap("NameFoo", "SourceFoo"))
		if err := tm.CreateMap("NameFoo", "SourceFoo"); err == nil {
			t.ErrorNow("expected error")
		}
	})
}

func TestSubmitCommit(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		t.CheckError(tm.CreateUser("NameFoo","PublicKey"))
		t.CheckError(tm.SubmitCommit("NameFoo", CategoryGeneral, "abcdef", time.Now()))
		if commits, err := tm.ListCommits("NameFoo", CategoryGeneral); err != nil {
			t.ErrorNow(err)
		} else if len(commits) != 1 {
			t.ErrorNow(len(commits), "but expected", 1)
		} else if commits[0] != "abcdef" {
			t.ErrorNow(commits[0], "but expected", "abcdef")
		}
	})
}

func TestCreateMatch(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		p1 := Submission{"p1","c1"}
		p2 := Submission{"p2","c2"}
		t.CheckError(tm.CreateMatch(CategoryGeneral, "MapFoo", p1, p2, time.Now()))
		if result, err := tm.GetMatchResult(CategoryGeneral, "MapFoo", p1, p2); err != nil {
			t.ErrorNow(err)
		} else if result != MatchResultInProgress {
			t.FailNow()
		}
	})
}

func TestUpdateMatch(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		p1 := Submission{"p1","c1"}
		p2 := Submission{"p2","c2"}
		t.CheckError(tm.CreateMatch(CategoryGeneral, "MapFoo", p1, p2, time.Now()))
		t.CheckError(tm.UpdateMatch(CategoryGeneral, "MapFoo", p1, p2, time.Now(), MatchResultWinA, "LogFoo"))
		if result, err := tm.GetMatchResult(CategoryGeneral, "MapFoo", p1, p2); err != nil {
			t.ErrorNow(t, err)
		} else if result != MatchResultWinA {
			t.ErrorNow(result, " expected ", MatchResultWinA)
		} else if replay, err := tm.GetMatchReplay(CategoryGeneral, "MapFoo", p1, p2); err != nil {
			t.ErrorNow(err)
		} else if replay != "LogFoo" {
			t.ErrorNow(replay, " expected LogFoo")
		}
	})
}

func TestRunMatch(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		//TODO
	})
}



