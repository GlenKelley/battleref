package tournament

import (
	"os/user"
	"time"
	"testing"
	_ "github.com/mattn/go-sqlite3"
	"github.com/GlenKelley/battleref/arena"
	"github.com/GlenKelley/battleref/git"
	"github.com/GlenKelley/battleref/testing"
)

func TournamentTest(test * testing.T, f func(*testutil.T, *Tournament)) {
	t := (*testutil.T)(test)
	if host, err := git.CreateGitHost(":temp:", nil); err != nil {
		t.ErrorNow(err)
	} else {
		defer host.Cleanup()
		dummyArena := arena.DummyArena{time.Now(), arena.MatchResult{arena.WinnerA, arena.ReasonVictory, "UkVQTEFZIEJBU0U2NAo="}, nil}
		remote := git.TempRemote{}
		bootstrap := &arena.MinimalBootstrap{}
		if database, err := NewInMemoryDatabase(); err != nil {
			t.ErrorNow(err)
		} else if err = database.MigrateSchema(); err != nil {
			t.ErrorNow(err)
		} else {
			tournament := NewTournament(database, dummyArena, bootstrap, host, remote)
			f(t, tournament)
		}
	}
}

var gitoliteTestConf = git.GitoliteConf {
	"localhost",
	"foobar",
	"git-test",
	".ssh/webserver",
	".ssh/git",
}

func GitoliteTournamentTest(test * testing.T, f func(*testutil.T, *Tournament)) {
	t := (*testutil.T)(test)
	conf := gitoliteTestConf
	conf.AdminKey = testutil.PathRelativeToUserHome(t, conf.AdminKey)
	conf.SSHKey = testutil.PathRelativeToUserHome(t, conf.SSHKey)
	if host, err := git.CreateGitoliteHost(conf); err != nil {
		t.ErrorNow(err)
	} else if _, err := user.Lookup(host.User); err != nil {
		switch err.(type) {
		case user.UnknownUserError:
			t.Skipf("%v, skipping gitolite tests", err)
			t.SkipNow()
			default: t.ErrorNow(err)
		}
	} else if err := host.Reset(); err != nil {
		t.ErrorNow(err)
	} else {
		defer host.Cleanup()
		dummyArena := arena.DummyArena{time.Now(), arena.MatchResult{arena.WinnerA, arena.ReasonVictory, "UkVQTEFZIEJBU0U2NAo="}, nil}
		remote := git.TempRemote{}
		bootstrap := &arena.MinimalBootstrap{}
		if database, err := NewInMemoryDatabase(); err != nil {
			t.ErrorNow(err)
		} else if err = database.MigrateSchema(); err != nil {
			t.ErrorNow(err)
		} else {
			tournament := NewTournament(database, dummyArena, bootstrap, host, remote)
			f(t, tournament)
		}
	}
}

func TestListCategories(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		if categories, err := tm.ListCategories(); err != nil {
			t.ErrorNow(err)
		} else if len(categories) != 1 {
			t.ErrorNowf("expected 1 category, got %v", len(categories))
		} else if categories[0] != CategoryGeneral {
			t.ErrorNowf("expected %v category, got %v", CategoryGeneral, categories[0])
		}
	})
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
		if _, err := tm.CreateUser("NameFoo", "PublicKey"); err != nil {
			t.ErrorNow(err)
		}
		isUser, err := tm.UserExists("NameFoo")
		if err != nil { t.ErrorNow(err) }
		if !isUser { t.FailNow() }
		users, err := tm.ListUsers()
		if err != nil { t.ErrorNow(err) }
		if len(users) != 1 { t.FailNow() }
		if users[0] != "NameFoo" { t.FailNow() }
	})
}

func TestCreateDuplicateUserError(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		if _, err := tm.CreateUser("NameFoo", "PublicKeyFoo"); err != nil {
			t.ErrorNow(err)
		}
		if _, err := tm.CreateUser("NameFoo", "PublicKeyFoo"); err == nil {
			t.ErrorNow("expected error")
		}
	})
}

func TestCreateExistingUserError(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		if _, err := tm.CreateUser("NameFoo", "PublicKeyFoo"); err != nil {
			t.ErrorNow(err)
		}
		if _, err := tm.CreateUser("NameFoo", "PublicKeyBar"); err == nil {
			t.ErrorNow("expected error")
		}
	})
}

func TestCreateExistingKeyError(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		if _, err := tm.CreateUser("NameFoo", "PublicKeyFoo"); err != nil {
			t.ErrorNow(err)
		}
		if _, err := tm.CreateUser("NameBar", "PublicKeyFoo"); err == nil {
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
		if _, err := tm.CreateUser("NameFoo", "PublicKeyFoo"); err != nil {
			t.ErrorNow(err)
		}
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
		p1 := Submission{"p1","c1"}
		p2 := Submission{"p2","c2"}
		t.CheckError(tm.CreateMap("MapFoo", "SourceFoo"))
		if result, err := tm.RunMatch(CategoryGeneral, "MapFoo", p1, p2, SystemClock()); err != nil {
			t.ErrorNow(err)
		} else if result != "WinA" {
			t.ErrorNowf("Expected WinA not %v\n", result)
		} else if result2, err := tm.GetMatchResult(CategoryGeneral, "MapFoo", p1, p2); err != nil {
			t.ErrorNow(err)
		} else if result != result2 {
			t.ErrorNowf("Expected %v not %v\n", result, result2)
		} else if matches, err := tm.ListMatches(); err != nil {
			t.ErrorNow(err)
		} else if len(matches) != 1 {
			t.ErrorNowf("Expected 1 match not %v\n", len(matches))
		} else if matches[0].Result != result {
			t.ErrorNowf("Expected %v not %v\n", result, matches[0].Result)
		}
	})
}

func TestRunLatestMatches(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		if _, err := tm.CreateUser("Name1", "PublicKey1"); err != nil {
			t.ErrorNow(err)
		}
		if _, err := tm.CreateUser("Name2", "PublicKey2"); err != nil {
			t.ErrorNow(err)
		}
		if _, err := tm.CreateUser("Name3", "PublicKey3"); err != nil {
			t.ErrorNow(err)
		}
		t.CheckError(tm.CreateMap("Map1", "MapSource"))
		t.CheckError(tm.CreateMap("Map2", "MapSource"))
		t.CheckError(tm.CreateMap("Map3", "MapSource"))
		date := time.Now()
		t.CheckError(tm.SubmitCommit("Name1", CategoryGeneral, "a1", date))
		t.CheckError(tm.SubmitCommit("Name1", CategoryGeneral, "a2", date.Add(time.Hour)))
		t.CheckError(tm.SubmitCommit("Name2", CategoryGeneral, "b1", date))
		t.CheckError(tm.SubmitCommit("Name2", CategoryGeneral, "b2", date.Add(time.Hour)))
		t.CheckError(tm.SubmitCommit("Name3", CategoryGeneral, "c1", date))
		t.CheckError(tm.SubmitCommit("Name3", CategoryGeneral, "c2", date.Add(time.Hour)))
		if err := tm.RunLatestMatches(CategoryGeneral); err != nil {
			t.ErrorNow(err)
		}
		if matches, err := tm.ListMatches(); err != nil {
			t.ErrorNow(err)
		} else if len(matches) != 18 {
			t.ErrorNow("Expected 1 got", len(matches))
		}

	})
}

func TestDuplicatePublicKey(t *testing.T) {
	GitoliteTournamentTest(t, func(t *testutil.T, tm *Tournament) {
	})
}

