package server

import (
	"strings"
	"io/ioutil"
	"net/http"
	"testing"
	"runtime"
	"net/http/httptest"
	_ "github.com/mattn/go-sqlite3"
	"github.com/GlenKelley/battleref/tournament"
)
// import "code.google.com/p/gomock/gomock"

func ErrorNow(t *testing.T, arg ... interface{}) {
	t.Error(arg ... )
	trace := make([]byte, 1024)
	count := runtime.Stack(trace, false)
	t.Errorf("Stack of %d bytes: %s", count, trace)
	t.FailNow()
}

func createServer(t * testing.T) (*ServerState) {
	if database, err := tournament.NewInMemoryDatabase(); err != nil {
		ErrorNow(t, err)
		return nil
	} else if err = database.MigrateSchema(); err != nil {
		ErrorNow(t, err)
		return nil
	} else {
		tournament := tournament.NewTournament(database)
		properties := Properties {
			":memory:",
			"8080",
		}
		return NewServer(tournament, properties)
	}

}

func getResponse(t *testing.T, server *ServerState, url string) string {
	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	server.Handler.ServeHTTP(resp, req)
	if p, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Error(err)
		t.FailNow()
		return ""
	} else {
		return string(p)
	}
}

func TestVersion(t *testing.T) {
	server := createServer(t)
	ps := getResponse(t, server, "/version")
	if !strings.Contains(ps, "SchemaVersion") {
		t.FailNow()
	}
	if !strings.Contains(ps, "SourceVersion") {
		t.FailNow()
	}
}


