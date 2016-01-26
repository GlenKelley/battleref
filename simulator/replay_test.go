package simulator

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"github.com/GlenKelley/battleref/testing"
	"io/ioutil"
	"regexp"
	"testing"
)

func TestReplay2015(t *testing.T) {
	testReplay((*testutil.T)(t), "battlecode2015")
}

func TestReplay2016(t *testing.T) {
	testReplay((*testutil.T)(t), "battlecode2016")
}

func testReplay(t *testutil.T, category string) {
	if bs, err := ioutil.ReadFile(category + "/replay.xml.gz"); err != nil {
		t.ErrorNow(err)
	} else if reader, err := gzip.NewReader(bytes.NewReader(bs)); err != nil {
		t.ErrorNow(err)
	} else if replay, err := NewReplay(reader, category); err != nil {
		t.ErrorNow(err)
	} else if regen, err := xml.MarshalIndent(replay, "", "  "); err != nil {
		t.ErrorNow(err)
	} else {
		r2, _ := gzip.NewReader(bytes.NewReader(bs))
		original, _ := ioutil.ReadAll(r2)
		regen = bytes.Replace(regen, []byte("&#xA;"), []byte("\n"), -1)
		regen = bytes.Replace(regen, []byte("&#39;"), []byte("&apos;"), -1)
		regen = regexp.MustCompile("></[\\w.-]+>").ReplaceAll(regen, []byte("/>"))
		original = regexp.MustCompile("\"(\\d+)\\.0\"").ReplaceAll(original, []byte("\"$1\""))
		regen = regexp.MustCompile("\"(\\d+)\\.0\"").ReplaceAll(regen, []byte("\"$1\""))
		t.StringCompare(string(original), string(regen))
	}
}
