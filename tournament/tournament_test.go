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
		dummyArena := arena.DummyArena{time.Now(), arena.MatchResult{arena.WinnerA, arena.ReasonVictory, []byte("MATCH_RESULT")}, nil}
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
		dummyArena := arena.DummyArena{time.Now(), arena.MatchResult{arena.WinnerA, arena.ReasonVictory, []byte("MATCH_RESULT")}, nil}
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
		categories := tm.ListCategories()
		if len(categories) != 2 {
			t.ErrorNowf("expected 2 category, got %v", len(categories))
		}
		as := []string{}
		bs := []string{string(CategoryBattlecode2014), string(CategoryBattlecode2015)}
		for _, category := range categories {
			as = append(as, string(category))
		}
		t.CompareStringsUnsorted(as, bs)
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
		if _, err := tm.CreateUser("NameFoo", "PublicKey", CategoryTest); err != nil {
			t.ErrorNow(err)
		}
		if isUser, err := tm.UserExists("NameFoo"); err != nil {
			t.ErrorNow(err)
		} else if !isUser {
			t.FailNow()
		} else if users, err := tm.ListUsers(); err != nil {
			t.ErrorNow(err)
		} else if len(users) != 1 {
			t.FailNow()
		} else if users[0] != "NameFoo" {
			t.FailNow()
		}
	})
}

func TestCreateDuplicateUserError(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		if _, err := tm.CreateUser("NameFoo", "PublicKeyFoo", CategoryTest); err != nil {
			t.ErrorNow(err)
		}
		if _, err := tm.CreateUser("NameFoo", "PublicKeyFoo", CategoryTest); err == nil {
			t.ErrorNow("expected error")
		}
	})
}

func TestCreateExistingUserError(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		if _, err := tm.CreateUser("NameFoo", "PublicKeyFoo", CategoryTest); err != nil {
			t.ErrorNow(err)
		}
		if _, err := tm.CreateUser("NameFoo", "PublicKeyBar", CategoryTest); err == nil {
			t.ErrorNow("expected error")
		}
	})
}

func TestCreateExistingKey(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		if _, err := tm.CreateUser("NameFoo", "PublicKeyFoo", CategoryTest); err != nil {
			t.ErrorNow(err)
		}
		if _, err := tm.CreateUser("NameBar", "PublicKeyFoo", CategoryTest); err != nil {
			t.ErrorNow(err)
		}
	})
}

func TestCreateMap(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		t.CheckError(tm.CreateMap("MapFoo", "MapString", CategoryTest))
		if maps, err := tm.ListMaps(CategoryTest); err != nil {
			t.ErrorNow(err)
		} else if len(maps) != 1 {
			t.ErrorNow(len(maps), "but expected", 1)
		} else if maps[0] != "MapFoo" {
			t.ErrorNow(maps[0], "but expected", "MapFoo")
		}
		if mapSource, err := tm.GetMapSource("MapFoo", CategoryTest); err != nil {
			t.ErrorNow(err)
		} else if mapSource != "MapString" {
			t.ErrorNow(mapSource, "but expected", "MapString")
		}
	})
}

func TestCreateExistingMapError(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		t.CheckError(tm.CreateMap("NameFoo", "SourceFoo", CategoryTest))
		if err := tm.CreateMap("NameFoo", "SourceFoo", CategoryTest); err == nil {
			t.ErrorNow("expected error")
		}
	})
}

func TestSubmitCommit(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		if _, err := tm.CreateUser("NameFoo", "PublicKeyFoo", CategoryTest); err != nil {
			t.ErrorNow(err)
		}
		t.CheckError(tm.SubmitCommit("NameFoo", CategoryTest, "abcdef", time.Now()))
		if commits, err := tm.ListCommits("NameFoo", CategoryTest); err != nil {
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
		if id, err := tm.CreateMatch(CategoryTest, "MapFoo", p1, p2, time.Now()); err != nil {
			t.FailNow()
		} else if result, err := tm.GetMatchResult(id); err != nil {
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
		if id, err := tm.CreateMatch(CategoryTest, "MapFoo", p1, p2, time.Now()); err != nil {
			t.ErrorNow(err)
		} else {
			t.CheckError(tm.UpdateMatch(CategoryTest, "MapFoo", p1, p2, time.Now(), MatchResultWinA, []byte("LogFoo")))
			if result, err := tm.GetMatchResult(id); err != nil {
				t.ErrorNow(t, err)
			} else if result != MatchResultWinA {
				t.ErrorNow(result, " expected ", MatchResultWinA)
			} else if replay, err := tm.GetMatchReplay(id); err != nil {
				t.ErrorNow(err)
			} else if string(replay) != "LogFoo" {
				t.ErrorNow(replay, " expected LogFoo")
			}
		}
	})
}

func TestRunMatch(t *testing.T) {
	TournamentTest(t, func(t *testutil.T, tm *Tournament) {
		p1 := Submission{"p1","c1"}
		p2 := Submission{"p2","c2"}
		t.CheckError(tm.CreateMap("MapFoo", "SourceFoo", CategoryTest))
		if id, result, err := tm.RunMatch(CategoryTest, "MapFoo", p1, p2, SystemClock()); err != nil {
			t.ErrorNow(err)
		} else if result != "WinA" {
			t.ErrorNowf("Expected WinA not %v\n", result)
		} else if result2, err := tm.GetMatchResult(id); err != nil {
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
		if _, err := tm.CreateUser("Name1", "PublicKey1", CategoryTest); err != nil {
			t.ErrorNow(err)
		}
		if _, err := tm.CreateUser("Name2", "PublicKey2", CategoryTest); err != nil {
			t.ErrorNow(err)
		}
		if _, err := tm.CreateUser("Name3", "PublicKey3", CategoryTest); err != nil {
			t.ErrorNow(err)
		}
		t.CheckError(tm.CreateMap("Map1", "MapSource", CategoryTest))
		t.CheckError(tm.CreateMap("Map2", "MapSource", CategoryTest))
		t.CheckError(tm.CreateMap("Map3", "MapSource", CategoryTest))
		date := time.Now()
		t.CheckError(tm.SubmitCommit("Name1", CategoryTest, "a1", date))
		t.CheckError(tm.SubmitCommit("Name1", CategoryTest, "a2", date.Add(time.Hour)))
		t.CheckError(tm.SubmitCommit("Name2", CategoryTest, "b1", date))
		t.CheckError(tm.SubmitCommit("Name2", CategoryTest, "b2", date.Add(time.Hour)))
		t.CheckError(tm.SubmitCommit("Name3", CategoryTest, "c1", date))
		t.CheckError(tm.SubmitCommit("Name3", CategoryTest, "c2", date.Add(time.Hour)))
		if err := tm.RunLatestMatches(CategoryTest); err != nil {
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

