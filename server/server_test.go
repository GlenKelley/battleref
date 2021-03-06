package server

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"github.com/GlenKelley/battleref/arena"
	"github.com/GlenKelley/battleref/git"
	"github.com/GlenKelley/battleref/testing"
	"github.com/GlenKelley/battleref/tournament"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	UnescapedSamplePublicKey = "ssh-rsa+AAAAB4NzaC1yc2EAAAADAQABAAABAQCTxaDi3ImnIVHDeu3Gy/qjB/P2Bnv2JSiJa12b8obRAHhdE0cA3D5i26fnBQtssixapgwtDeADkeyKm+KhCtGbXdObQFDiDnWmUAxhjPyXwIHfvWwjYSIoPB9w8137wtOEVh9L2FtU3gL948VO589a5PsTeNTLmyJP07KcOdOtdKzgg14/rfRv6/jfzPKfRCz4b36siYdeLYc4Qg4L2TjGiP/4UtwfkkvrEBisw54v2hNCjzqBfKzwq3gzwZk8/KKQvYChdqypMN6GP18JDMv4ztroJt9awcEbk43iuQiMwDBE73ePs6ColoPKHB+OFCa/cQBS6ZzaNJd2OL3AUy1==+sample@public.key.com"
	SamplePublicKey          = "ssh-rsa AAAAB4NzaC1yc2EAAAADAQABAAABAQCTxaDi3ImnIVHDeu3Gy/qjB/P2Bnv2JSiJa12b8obRAHhdE0cA3D5i26fnBQtssixapgwtDeADkeyKm+KhCtGbXdObQFDiDnWmUAxhjPyXwIHfvWwjYSIoPB9w8137wtOEVh9L2FtU3gL948VO589a5PsTeNTLmyJP07KcOdOtdKzgg14/rfRv6/jfzPKfRCz4b36siYdeLYc4Qg4L2TjGiP/4UtwfkkvrEBisw54v2hNCjzqBfKzwq3gzwZk8/KKQvYChdqypMN6GP18JDMv4ztroJt9awcEbk43iuQiMwDBE73ePs6ColoPKHB+OFCa/cQBS6ZzaNJd2OL3AUy1== sample@public.key.com"
	SamplePublicKey2         = "ssh-rsa AAAAB4NzaC1yc2EAAAADAQABAAABAQCTxaDi3ImnIVHDeu3Gy/qjB/P2Bnv2JSiJa12b8obRAHhdE0cA3D5i26fnBQtssixapgwtDeADkeyKm+KhCtGbXdObQFDiDnWmUAxhjPyXwIHfvWwjYSIoPB9w8137wtOEVh9L2FtU3gL948VO589a5PsTeNTLmyJP07KcOdOtdKzgg14/rfRv6/jfzPKfRCz4b36siYdeLYc4Qg4L2TjGiP/4UtwfkkvrEBisw54v2hNCjzqBfKzwq3gzwZk8/KKQvYChdqypMN6GP18JDMv4ztroJt9awcEbk43iuQiMwDBE73ePs6ColoPKHB+OFCa/cQBS6ZzaNJd2OL3AUy2== sample@public.key.com"
	SimilarPublicKey         = "ssh-rsa AAAAB4NzaC1yc2EAAAADAQABAAABAQCTxaDi3ImnIVHDeu3Gy/qjB/P2Bnv2JSiJa12b8obRAHhdE0cA3D5i26fnBQtssixapgwtDeADkeyKm+KhCtGbXdObQFDiDnWmUAxhjPyXwIHfvWwjYSIoPB9w8137wtOEVh9L2FtU3gL948VO589a5PsTeNTLmyJP07KcOdOtdKzgg14/rfRv6/jfzPKfRCz4b36siYdeLYc4Qg4L2TjGiP/4UtwfkkvrEBisw54v2hNCjzqBfKzwq3gzwZk8/KKQvYChdqypMN6GP18JDMv4ztroJt9awcEbk43iuQiMwDBE73ePs6ColoPKHB+OFCa/cQBS6ZzaNJd2OL3AUy1== similar@public.key.com"
	SampleCommitHash         = "012345"
)

func ServerTest(test *testing.T, f func(*testutil.T, *ServerState)) {
	t := (*testutil.T)(test)
	if host, err := git.CreateGitHost(":temp:", nil); err != nil {
		t.ErrorNow(err)
	} else {
		defer host.Cleanup()
		if replay, err := ioutil.ReadFile("../simulator/" + string(tournament.CategoryBattlecode2015) + "/replay.xml.gz"); err != nil {
			t.ErrorNow()
		} else {
			dummyArena := arena.DummyArena{time.Now(), arena.MatchResult{arena.WinnerA, arena.ReasonVictory, replay}, nil}
			remote := &git.TempRemote{}
			bootstrap := &arena.MinimalBootstrap{"../arena/internal/categories"}
			if database, err := tournament.NewInMemoryDatabase(); err != nil {
				t.ErrorNow(err)
			} else if err = database.MigrateSchema(); err != nil {
				t.ErrorNow(err)
			} else {
				tournament := tournament.NewTournament(database, dummyArena, bootstrap, host, remote)
				properties := Properties{
					":memory:",
					"8081",
					":temp:",
					nil,
					"../arena",
				}
				server := NewServer(tournament, properties)
				f(t, server)
			}
		}
	}
}

func sendGet(t *testutil.T, server *ServerState, url string) JSONResponse {
	if req, err := http.NewRequest("GET", url, nil); err != nil {
		t.ErrorNow(err)
		return nil
	} else {
		return sendRequest(t, server, http.StatusOK, req)
	}
}

func sendRawGet(t *testutil.T, server *ServerState, url string) []byte {
	if req, err := http.NewRequest("GET", url, nil); err != nil {
		t.ErrorNow(err)
		return nil
	} else {
		body := sendRawRequest(t, server, http.StatusOK, req)
		if p, err := ioutil.ReadAll(body); err != nil {
			t.ErrorNow(err)
			return nil
		} else {
			return p
		}
	}
}

func sendPost(t *testutil.T, server *ServerState, url string, body io.Reader) JSONResponse {
	if req, err := http.NewRequest("POST", url, body); err != nil {
		t.ErrorNow(err)
		return nil
	} else {
		return sendRequest(t, server, http.StatusOK, req)
	}
}

func sendPostExpectStatus(t *testutil.T, server *ServerState, expectedCode int, url string, body io.Reader) JSONResponse {
	if req, err := http.NewRequest("POST", url, body); err != nil {
		t.ErrorNow(err)
		return nil
	} else {
		return sendRequest(t, server, expectedCode, req)
	}
}

func sendJSONPost(t *testutil.T, server *ServerState, url string, body interface{}) JSONResponse {
	return sendJSONPostExpectStatus(t, server, http.StatusOK, url, body)
}

func sendGetExpectStatus(t *testutil.T, server *ServerState, expectedCode int, url string) JSONResponse {
	if req, err := http.NewRequest("GET", url, nil); err != nil {
		t.ErrorNow(err)
		return nil
	} else {
		return sendRequest(t, server, expectedCode, req)
	}
}

func sendJSONPostExpectStatus(t *testutil.T, server *ServerState, expectedCode int, url string, body interface{}) JSONResponse {
	if bs, err := json.Marshal(body); err != nil {
		t.ErrorNow(err)
		return nil
	} else if req, err := http.NewRequest("POST", url, bytes.NewReader(bs)); err != nil {
		t.ErrorNow(err)
		return nil
	} else {
		req.Header.Set("Content-Type", "application/json")
		return sendRequest(t, server, expectedCode, req)
	}
}

func sendRawRequest(t *testutil.T, server *ServerState, expectedCode int, req *http.Request) *bytes.Buffer {
	resp := httptest.NewRecorder()
	server.HttpServer.Handler.ServeHTTP(resp, req)
	if resp.Code != expectedCode {
		t.Error("Status", resp.Code, "expected", expectedCode)
		if p, err := ioutil.ReadAll(resp.Body); err == nil {
			t.Log(string(p))
		}
		t.FailNow()
	}
	return resp.Body
}

func sendRequest(t *testutil.T, server *ServerState, expectedCode int, req *http.Request) JSONResponse {
	var jsonResponse JSONResponse
	body := sendRawRequest(t, server, expectedCode, req)
	t.CheckError(json.NewDecoder(body).Decode(&jsonResponse))
	return jsonResponse
}

func TestVersion(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		r := sendGet(t, server, "/version")
		if r["schemaVersion"] == "" {
			t.FailNow()
		}
		if r["sourceVersion"] == "" {
			t.FailNow()
		}
	})
}

func TestShutdown(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		go server.Serve()
		//Race condition of server not starting
		time.Sleep(time.Millisecond)
		r := sendPost(t, server, "/shutdown", nil)
		if r["shutdown"] == "" {
			t.FailNow()
		}
		sendPostExpectStatus(t, server, http.StatusInternalServerError, "/shutdown", nil)
	})
}

func TestParseForm(test *testing.T) {
	t := (*testutil.T)(test)
	data := url.Values{}
	data.Set("foo", "x")
	data.Add("bar", "y")
	if req, err := http.NewRequest("POST", "/?a=b", bytes.NewBufferString(data.Encode())); err != nil {
		t.ErrorNow(err)
	} else {
		var form struct {
			A string `json:"foo" form:"foo" validate:"required"`
			B string `json:"bar" form:"bar"`
			C string `json:"moo"`
		}
		t.CheckError(parseForm(req, &form))
		if form.A != "x" {
			t.Error(form.A, "expected x")
		}
		if form.B != "y" {
			t.Error(form.B, "expected y")
		}
		if form.C != "" {
			t.Error(form.C, "expected ''")
		}
	}
}

func TestParseFormJSON(test *testing.T) {
	t := (*testutil.T)(test)
	if req, err := http.NewRequest("POST", "/", strings.NewReader("{\"foo\":\"x\",\"bar\":\"y\"}")); err != nil {
		t.ErrorNow(err)
	} else {
		req.Header.Set("Content-Type", "application/json")
		var form struct {
			A string `json:"foo" form:"foo" validate:"required"`
			B string `json:"bar" form:"bar"`
			C string `json:"moo"`
		}
		t.CheckError(parseForm(req, &form))
		if form.A != "x" {
			t.Error(form.A, "expected x")
		}
		if form.B != "y" {
			t.Error(form.B, "expected y")
		}
		if form.C != "" {
			t.Error(form.C, "expected ''")
		}
	}
}

func TestRegisterForm(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		r := sendPost(t, server, "/register", strings.NewReader("name=NameFoo&category="+string(tournament.CategoryTest)+"&public_key="+url.QueryEscape(SamplePublicKey)))
		if Json(t, r).Key("data").Key("name").String() != "NameFoo" {
			t.FailNow()
		}
		if Json(t, r).Key("data").Key("public_key").String() != SamplePublicKey {
			t.FailNow()
		}
	})
}

func TestUnescaptedParsingFails(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		if r := sendPostExpectStatus(t, server, http.StatusInternalServerError, "/register", strings.NewReader("name=NameFoo&category="+string(tournament.CategoryTest)+"&public_key="+UnescapedSamplePublicKey)); Json(t, r).Key("error").Key("message").String() != "Invalid Public Key" {
			t.ErrorNow("expected 'Invalid Public Key'")
		}
	})
}

func TestSimilarPublicKeys(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		r := sendPost(t, server, "/register", strings.NewReader("name=NameFoo&category="+string(tournament.CategoryTest)+"&public_key="+url.QueryEscape(SamplePublicKey)))
		if Json(t, r).Key("data").Key("name").String() != "NameFoo" {
			t.FailNow()
		}
		if Json(t, r).Key("data").Key("public_key").String() != SamplePublicKey {
			t.FailNow()
		}
		r = sendPost(t, server, "/register", strings.NewReader("name=NameBar&category="+string(tournament.CategoryTest)+"&public_key="+url.QueryEscape(SimilarPublicKey)))
		if Json(t, r).Key("data").Key("name").String() != "NameBar" {
			t.FailNow()
		}
		if Json(t, r).Key("data").Key("public_key").String() != SimilarPublicKey {
			t.FailNow()
		}
	})
}

func TestRegisterQuery(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		r := sendPost(t, server, "/register?name=NameFoo&category="+string(tournament.CategoryTest)+"&public_key="+url.QueryEscape(SamplePublicKey), nil)
		if Json(t, r).Key("data").Key("name").String() != "NameFoo" {
			t.FailNow()
		}
		if Json(t, r).Key("data").Key("public_key").String() != SamplePublicKey {
			t.FailNow()
		}
	})
}

func TestRegisterJSON(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		r := sendJSONPost(t, server, "/register", map[string]string{"name": "NameFoo", "public_key": SamplePublicKey, "category": string(tournament.CategoryTest)})
		if Json(t, r).Key("data").Key("name").String() != "NameFoo" {
			t.FailNow()
		}
		if Json(t, r).Key("data").Key("public_key").String() != SamplePublicKey {
			t.FailNow()
		}
	})
}

func compareStrings(a []interface{}, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, as := range a {
		if as != b[i] {
			return false
		}
	}
	return true
}

func compareStringsUnordered(a []interface{}, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	c := make(map[string]int)
	for _, s := range a {
		c[s.(string)]++
	}
	for _, s := range b {
		c[s]--
	}
	for _, i := range c {
		if i > 0 {
			return false
		}
	}
	return true
}

type JsonWrapper struct {
	T    *testutil.T
	Node interface{}
}

func Json(t *testutil.T, node interface{}) JsonWrapper {
	return JsonWrapper{t, node}
}

func (j JsonWrapper) Key(key string) JsonWrapper {
	if j.Node == nil {
		j.T.ErrorNow("json is null")
	}
	switch n := j.Node.(type) {
	case map[string]interface{}:
		if _, ok := n[key]; !ok {
			j.T.ErrorNow("no value for key", key, n)
		}
		return JsonWrapper{j.T, n[key]}
	case JSONResponse:
		if _, ok := n[key]; !ok {
			j.T.ErrorNow("no value for key", key, n)
		}
		return JsonWrapper{j.T, n[key]}
	default:
		j.T.ErrorNow("invalid json map", reflect.TypeOf(n))
		return JsonWrapper{}
	}
}

func (j JsonWrapper) Array() []interface{} {
	if j.Node == nil {
		j.T.ErrorNow("json map is null")
	}
	switch n := j.Node.(type) {
	case []interface{}:
		return n
	default:
		j.T.ErrorNow("invalid json array", reflect.TypeOf(n))
		return nil
	}
}

func (j JsonWrapper) At(i int) JsonWrapper {
	if j.Node == nil {
		j.T.ErrorNow("json array is null")
	}
	switch n := j.Node.(type) {
	case []interface{}:
		return JsonWrapper{j.T, n[i]}
	default:
		j.T.ErrorNow("invalid json array", n)
		return JsonWrapper{}
	}
}

func (j JsonWrapper) Int() int {
	switch n := j.Node.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		j.T.ErrorNow("invalid json int", n)
		return 0
	}
}

func (j JsonWrapper) String() string {
	switch n := j.Node.(type) {
	case string:
		return n
	default:
		j.T.ErrorNow("invalid json string", n)
		return ""
	}
}

func (j JsonWrapper) Len() int {
	switch n := j.Node.(type) {
	case []interface{}:
		return len(n)
	default:
		j.T.ErrorNow("invalid json array", n)
		return 0
	}
}

func TestCategories(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		if r := sendGet(t, server, "/categories"); !compareStringsUnordered(Json(t, r).Key("data").Key("categories").Array(), []string{string(tournament.CategoryBattlecode2014), string(tournament.CategoryBattlecode2015), string(tournament.CategoryBattlecode2016)}) {
			t.ErrorNow("expected 3 categories", r)
		}
	})
}

func TestPlayers(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		if r := sendGet(t, server, "/players"); Json(t, r).Key("data").Key("players").Len() > 0 {
			t.ErrorNow("expected no players", r)
		}
		sendJSONPost(t, server, "/register", map[string]string{"name": "NameFoo", "public_key": SamplePublicKey, "category": string(tournament.CategoryTest)})
		if r := sendGet(t, server, "/players"); !compareStrings(Json(t, r).Key("data").Key("players").Array(), []string{"NameFoo"}) {
			t.ErrorNow("expected single player NameFoo", r)
		}
		sendJSONPost(t, server, "/register", map[string]string{"name": "NameBar", "public_key": SamplePublicKey2, "category": string(tournament.CategoryTest)})
		if r := sendGet(t, server, "/players"); !compareStringsUnordered(Json(t, r).Key("data").Key("players").Array(), []string{"NameFoo", "NameBar"}) {
			t.ErrorNow("expected two players NameFoo, NameBar", r)
		}
	})
}

func TestMaps(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		if r := sendGet(t, server, "/maps?category="+string(tournament.CategoryTest)); Json(t, r).Key("data").Key("maps").Len() > 0 {
			t.Error("expected no maps", r)
			t.FailNow()
		}
		sendJSONPost(t, server, "/map/create", map[string]string{"name": "NameFoo", "source": "SourceFoo", "category": string(tournament.CategoryTest)})
		if r := sendGet(t, server, "/maps?category="+string(tournament.CategoryTest)); !compareStrings(Json(t, r).Key("data").Key("maps").Array(), []string{"NameFoo"}) {
			t.Error("expected single player NameFoo", r)
			t.FailNow()
		}
		sendJSONPost(t, server, "/map/create", map[string]string{"name": "NameBar", "source": "SourceBar", "category": string(tournament.CategoryTest)})
		if r := sendGet(t, server, "/maps?category="+string(tournament.CategoryTest)); !compareStringsUnordered(Json(t, r).Key("data").Key("maps").Array(), []string{"NameFoo", "NameBar"}) {
			t.ErrorNow("expected two maps NameFoo, NameBar", r)
		}
	})
}

func TestSubmitHash(test *testing.T) {
	t := (*testutil.T)(test)
	if !CommitHashRegex.MatchString(SampleCommitHash) {
		t.ErrorNowf("Expected match from <%v>", SampleCommitHash)
	}
	for _, invalidKey := range []string{"FooBar"} {
		if CommitHashRegex.MatchString(invalidKey) {
			t.ErrorNowf("Expected match from <%v>", invalidKey)
		}
	}
}

func TestSubmit(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		sendJSONPost(t, server, "/register", map[string]string{"name": "NameFoo", "public_key": SamplePublicKey, "category": string(tournament.CategoryTest)})
		if r := sendJSONPost(t, server, "/submit", map[string]string{"name": "NameFoo", "category": string(tournament.CategoryTest), "commit_hash": SampleCommitHash}); Json(t, r).Key("data").Key("name").String() != "NameFoo" {
			t.ErrorNow(r["name"], " expected ", "NameFoo")
		} else if Json(t, r).Key("data").Key("category").String() != string(tournament.CategoryTest) {
			t.ErrorNow(r["category"], " expected ", string(tournament.CategoryTest))
		} else if Json(t, r).Key("data").Key("commit_hash").String() != SampleCommitHash {
			t.ErrorNow(r["commit_hash"], " expected ", SampleCommitHash)
		}
	})
}

func TestSubmitPlayerNameError(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		if r := sendJSONPostExpectStatus(t, server, http.StatusInternalServerError, "/submit", map[string]string{"name": "NameFoo", "category": string(tournament.CategoryTest), "commit_hash": SampleCommitHash}); Json(t, r).Key("error").Key("message").String() != "Unknown player" {
			t.ErrorNow(r, "expected 'Unknown player'")
		}
	})
}

func TestSubmitCommitHashError(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		sendJSONPost(t, server, "/register", map[string]string{"name": "NameFoo", "public_key": SamplePublicKey, "category": string(tournament.CategoryTest)})
		if r := sendJSONPostExpectStatus(t, server, http.StatusInternalServerError, "/submit", map[string]string{"name": "NameFoo", "category": string(tournament.CategoryTest), "commit_hash": "InvalidCommitHash"}); Json(t, r).Key("error").Key("message").String() != "Invalid commit hash" {
			t.ErrorNow(r, "expected 'Unknown player'")
		}
	})
}

func TestSubmitDuplicateCommitError(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		sendJSONPost(t, server, "/register", map[string]string{"name": "NameFoo", "public_key": SamplePublicKey, "category": string(tournament.CategoryTest)})
		sendJSONPost(t, server, "/submit", map[string]string{"name": "NameFoo", "category": string(tournament.CategoryTest), "commit_hash": SampleCommitHash})
		sendJSONPostExpectStatus(t, server, http.StatusInternalServerError, "/submit", map[string]string{"name": "NameFoo", "category": string(tournament.CategoryTest), "commit_hash": SampleCommitHash})
	})
}

func TestCommits(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		sendJSONPost(t, server, "/register", map[string]string{"name": "NameFoo", "public_key": SamplePublicKey, "category": string(tournament.CategoryTest)})
		if r := sendGet(t, server, "/commits?name=NameFoo&category=General"); Json(t, r).Key("data").Key("commits").Len() > 0 {
			t.ErrorNow("expected no commits", r)
		}
		sendJSONPost(t, server, "/submit", map[string]string{"name": "NameFoo", "category": "General", "commit_hash": "abcdef"})
		if r := sendGet(t, server, "/commits?name=NameFoo&category=General"); !compareStringsUnordered(Json(t, r).Key("data").Key("commits").Array(), []string{"abcdef"}) {
			t.ErrorNow("expected single commit abcdef", r)
		}
		sendJSONPost(t, server, "/submit", map[string]string{"name": "NameFoo", "category": "General", "commit_hash": "012345"})
		if r := sendGet(t, server, "/commits?name=NameFoo&category=General"); !compareStringsUnordered(Json(t, r).Key("data").Key("commits").Array(), []string{"abcdef", "012345"}) {
			t.ErrorNow("expected two commits abcdef, 012345", r)
		}
	})
}

func TestReplay(t *testing.T) {
	ServerTest(t, func(t *testutil.T, server *ServerState) {
		sendJSONPost(t, server, "/register", map[string]string{"name": "NameFoo", "public_key": SamplePublicKey, "category": string(tournament.CategoryTest)})
		sendJSONPost(t, server, "/map/create", map[string]string{"name": "NameBar", "source": "SourceBar", "category": string(tournament.CategoryTest)})
		r := sendGet(t, server, "/commits?name=NameFoo&category="+string(tournament.CategoryTest))
		commit := Json(t, r).Key("data").Key("commits").At(0).String()
		r = sendJSONPost(t, server, "/match/run", map[string]string{"player1": "NameFoo", "player2": "NameFoo", "category": string(tournament.CategoryTest), "commit1": commit, "commit2": commit, "map": "NameBar"})
		id := Json(t, r).Key("data").Key("id").Int()
		sendRawGet(t, server, "/replay?id="+strconv.Itoa(id))
	})
}

func TestReplayStream(test *testing.T) {
	ServerTest(test, func(t *testutil.T, server *ServerState) {
		go server.Serve()
		//Race condition of server not starting
		time.Sleep(time.Microsecond)
		defer sendPost(t, server, "/shutdown", nil)
		sendJSONPost(t, server, "/register", map[string]string{"name": "NameFoo", "public_key": SamplePublicKey, "category": string(tournament.CategoryBattlecode2016)})
		sendJSONPost(t, server, "/map/create", map[string]string{"name": "NameBar", "source": "SourceBar", "category": string(tournament.CategoryBattlecode2016)})
		r := sendGet(t, server, "/commits?name=NameFoo&category="+string(tournament.CategoryBattlecode2016))
		commit := Json(t, r).Key("data").Key("commits").At(0).String()
		r = sendJSONPost(t, server, "/match/run", map[string]string{"player1": "NameFoo", "player2": "NameFoo", "category": string(tournament.CategoryBattlecode2016), "commit1": commit, "commit2": commit, "map": "NameBar"})
		id := Json(t, r).Key("data").Key("id").Int()
		msg := make([]byte, 4096)
		if ws, err := websocket.Dial(fmt.Sprintf("ws://localhost:8081/replay/stream?id=%v", id), "", "http://localhost"); err != nil {
			t.ErrorNow(err)
		} else {
			var err error
			for err == nil {
				if _, err = ws.Write([]byte("")); err != nil {
					t.ErrorNow(err)
				} else if _, err = ws.Read(msg); err != io.EOF && err != nil {
					t.ErrorNow(err)
				}
			}
		}
	})
}
