package testutil

import (
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"
)

type T testing.T

func (t *T) ErrorNow(args ...interface{}) {
	trace := make([]byte, 1024)
	count := runtime.Stack(trace, false)
	t.Errorf("Stack of %d bytes: %s", count, trace)
	t.Error(args...)
	t.FailNow()
}

func (t *T) ErrorNowf(format string, args ...interface{}) {
	trace := make([]byte, 1024)
	count := runtime.Stack(trace, false)
	t.Errorf("Stack of %d bytes: %s", count, trace)
	t.Errorf(format, args...)
	t.FailNow()
}

func (t *T) CheckError(err error, args ...interface{}) {
	if err != nil {
		t.ErrorNow(append([]interface{}{err}, args...)...)
	}
}

func (t *T) ExpectEqual(a, b interface{}) {
	if a != b {
		t.ErrorNowf("Expected <%v> = <%v>", a, b)
	}
}

func (t *T) CompareStringsUnsorted(as, bs []string) {
	counts := map[string]int{}
	for _, a := range as {
		counts[a]++
	}
	for _, b := range bs {
		counts[b]--
	}
	for k, c := range counts {
		if c != 0 {
			t.ErrorNowf("Different element <%v>: <%v> != <%v>", k, as, bs)
		}
	}
}

func prefix(s string, i int) string {
	if i < len(s) {
		return s[0:i]
	} else {
		return s
	}
}

func suffix(s string, i int) string {
	if i < len(s) {
		return s[len(s)-i-1:]
	} else {
		return s
	}

}

func (t *T) StringCompare(a, b string) {
	if a != b {
		n := len(a)
		if bn := len(b); bn < n {
			n = bn
		}
		for i := 0; i < n; i++ {
			if a[i] != b[i] {
				p := suffix(a[:i], 100)
				pa := prefix(a[i:], 512)
				pb := prefix(b[i:], 512)
				t.ErrorNowf("Strings don't match, differ at index %v: \n...%v[%v...]\n...%v[%v...]\n", i, p, pa, p, pb)
			}
		}
		if len(a) < len(b) {
			p := suffix(b[:n], 100)
			t.ErrorNowf("Strings don't match, differ at index %v, unexpected suffix: \n...%v[%v...]\n", n, p, prefix(b[n:], 100))
		} else {
			p := suffix(b[:n], 100)
			t.ErrorNowf("Strings don't match, differ at index %v, missing suffix: \n...%v[%v...]\n", n, p, prefix(a[n:], 100))
		}

	}
}

func CreateKeyPair() (string, string, error) {
	if dir, err := ioutil.TempDir(os.TempDir(), "battleref_keypair"); err != nil {
		return "", "", err
	} else {
		defer os.RemoveAll(dir)
		privateKeyFile := filepath.Join(dir, "key")
		publicKeyFile := privateKeyFile + ".pub"
		if err := exec.Command("ssh-keygen", "-t", "rsa", "-N", "", "-f", privateKeyFile).Run(); err != nil {
			return "", "", err
		}
		if privateKeyBytes, err := ioutil.ReadFile(privateKeyFile); err != nil {
			return "", "", err
		} else if publicKeyBytes, err := ioutil.ReadFile(publicKeyFile); err != nil {
			return "", "", err
		} else {
			return string(privateKeyBytes), string(publicKeyBytes), nil
		}
	}
}

func PathRelativeToUserHome(t *T, path string) string {
	if u, err := user.Current(); err != nil {
		t.ErrorNow(err)
		return ""
	} else {
		return filepath.Join(u.HomeDir, path)
	}
}
