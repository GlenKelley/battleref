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

func TestSubmitCommit(t *testing.T) {
	tm := createTournament(t)
	if err := tm.CreateUser("NameFoo","PublicKey"); err != nil {
		ErrorNow(t, err)
	}
	if err := tm.SubmitCommit("NameFoo","abcdef", time.Now()); err != nil {
		ErrorNow(t, err)
	}
	//TODO: test commit exists
}

func TestRunMatch(t *testing.T) {

}



