package tournament

import (
	"time"
	"testing"
	"runtime"
	_ "github.com/mattn/go-sqlite3"
)

func ErrorNow(t *testing.T, arg ... interface{}) {
	t.Error(arg ... )
	trace := make([]byte, 1024)
	count := runtime.Stack(trace, false)
	t.Errorf("Stack of %d bytes: %s", count, trace)
	t.FailNow()
}

func Check(t *testing.T, err error) {
	if err != nil {
		ErrorNow(t, err)
	}
}

func createTournament(t * testing.T) (*Tournament) {
	if database, err := NewInMemoryDatabase(); err != nil {
		ErrorNow(t, err)
		return nil
	} else if err = database.MigrateSchema(); err != nil {
		ErrorNow(t, err)
		return nil
	} else {
		return NewTournament(database)
	}
}

func TestCreateUser(t *testing.T) {
	tm := createTournament(t)
	if isUser, err := tm.UserExists("NameFoo"); err != nil {
		ErrorNow(t, err)
	} else if isUser {
		t.FailNow()
	}
	if users, err := tm.ListUsers(); err != nil {
		ErrorNow(t, err)
	} else if len(users) != 0 {
		t.FailNow()
	}
	if err := tm.CreateUser("NameFoo", "PublicKey"); err != nil {
		ErrorNow(t, err)
	}
	isUser, err := tm.UserExists("NameFoo")
	if err != nil { ErrorNow(t, err) }
        if !isUser { t.FailNow() }
	users, err := tm.ListUsers()
	if err != nil { ErrorNow(t, err) }
	if len(users) != 1 { t.FailNow() }
	if users[0] != "NameFoo" { t.FailNow() }
}

func TestCreateExistingUserError(t *testing.T) {
	tm := createTournament(t)
	Check(t, tm.CreateUser("NameFoo", "PublicKeyFoo"))
	if err := tm.CreateUser("NameFoo", "PublicKeyFoo"); err == nil {
		ErrorNow(t, "expected error")
	}
}

func TestCreateMap(t *testing.T) {
	tm := createTournament(t)
	if err := tm.CreateMap("MapFoo", "MapString"); err != nil {
		ErrorNow(t, err)
	}
	if maps, err := tm.ListMaps(); err != nil {
		ErrorNow(t, err)
	} else if len(maps) != 1 {
		ErrorNow(t, len(maps), "but expected", 1)
	} else if maps[0] != "MapFoo" {
		ErrorNow(t, maps[0], "but expected", "MapFoo")
	}
	if mapSource, err := tm.GetMapSource("MapFoo"); err != nil {
		ErrorNow(t, err)
	} else if mapSource != "MapString" {
		ErrorNow(t, mapSource, "but expected", "MapString")
	}
}

func TestCreateExistingMapError(t *testing.T) {
	tm := createTournament(t)
	Check(t, tm.CreateMap("NameFoo", "SourceFoo"))
	if err := tm.CreateMap("NameFoo", "SourceFoo"); err == nil {
		ErrorNow(t, "expected error")
	}
}

func TestSubmitCommit(t *testing.T) {
	tm := createTournament(t)
	if err := tm.CreateUser("NameFoo","PublicKey"); err != nil {
		ErrorNow(t, err)
	}
	if err := tm.SubmitCommit("NameFoo", CategoryGeneral, "abcdef", time.Now()); err != nil {
		ErrorNow(t, err)
	}
	if commits, err := tm.ListCommits("NameFoo", CategoryGeneral); err != nil {
		ErrorNow(t, err)
	} else if len(commits) != 1 {
		ErrorNow(t, len(commits), "but expected", 1)
	} else if commits[0] != "abcdef" {
		ErrorNow(t, commits[0], "but expected", "abcdef")
	}
}

func TestRunMatch(t *testing.T) {

}



