package testutil

import (
	"os"
	"testing"
	"path/filepath"
	"os/exec"
	"io/ioutil"
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


