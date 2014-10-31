package server

import (
	"fmt"
	"time"
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
			"8081",
			".",
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
	server.HttpServer.Handler.ServeHTTP(resp, req)
	if p, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Error(err)
		t.FailNow()
		return ""
	} else {
		return string(p)
	}
}

func errorResponse(t *testing.T, server *ServerState, url string) string {
	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	server.HttpServer.Handler.ServeHTTP(resp, req)
	if resp.Code == http.StatusOK {
		t.FailNow()
	}
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
	if !strings.Contains(ps, "schemaVersion") {
		t.FailNow()
	}
	if !strings.Contains(ps, "sourceVersion") {
		t.FailNow()
	}
}

func TestShutdown(t *testing.T) {
	server := createServer(t)
	go server.Serve()
	//Race condition of server not starting
	time.Sleep(time.Millisecond)
	ps := getResponse(t, server, "/shutdown")
	if !strings.Contains(ps, "Shutting Down") {
		t.FailNow()
	}
	ps = errorResponse(t, server, "/shutdown")
	fmt.Println(ps)
}


