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

func TestReplay(test *testing.T) {
	t := (*testutil.T)(test)
	if bs, err := ioutil.ReadFile("replay.xml.gz"); err != nil {
		t.ErrorNow(err)
	} else if reader, err := gzip.NewReader(bytes.NewReader(bs)); err != nil {
		t.ErrorNow(err)
	} else if replay, err := NewReplay(reader); err != nil {
		t.ErrorNow(err)
	} else if regen, err := xml.MarshalIndent(replay, "", "  "); err != nil {
		t.ErrorNow(err)
	} else {
		r2, _ := gzip.NewReader(bytes.NewReader(bs))
		original, _ := ioutil.ReadAll(r2)
		regen = bytes.Replace(regen, []byte("&#xA;"), []byte("\n"), -1)
		regen = regexp.MustCompile("></[\\w.]+>").ReplaceAll(regen, []byte("/>"))
		original = regexp.MustCompile("\"(\\d+)\\.0\"").ReplaceAll(original, []byte("\"$1\""))
		regen = regexp.MustCompile("\"(\\d+)\\.0\"").ReplaceAll(regen, []byte("\"$1\""))
		t.StringCompare(string(original), string(regen))
	}
}
