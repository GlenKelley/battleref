package testutil

import (
	"testing"
)

type T testing.T

func (t *T) ErrorNow(args ... interface{}) {
	t.Error(args ...)
	t.FailNow()
}

func (t *T) ErrorNowf(format string, args ... interface{}) {
	t.Errorf(format, args ...)
	t.FailNow()
}

func (t *T) CheckError(err error, args ... interface{}) {
	if err != nil {
		t.ErrorNow(append([]interface{}{err},args...) ...)
	}
}

func (t *T) ExpectEqual(a, b interface{}) {
	if (a != b) {
		t.ErrorNowf("Expected <%v> = <%v>", a, b)
	}
}

func (t *T) CompareStringsUnsorted(as, bs []string) {
	counts := map[string]int{}
	for _, a := range as { counts[a]++ }
	for _, b := range bs { counts[b]-- }
	for k, c := range counts {
		if c != 0 {
			t.ErrorNowf("Different element <%v>: <%v> != <%v>", k, as, bs)
		}
	}
}

