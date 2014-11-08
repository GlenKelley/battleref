package git

import (
	//"path/filepath"
	"fmt"
	"testing"
	"io/ioutil"
	"os"
	"os/user"
	"github.com/GlenKelley/battleref/testing"
)

func CheckDirectoryContent(t *testutil.T, dir string, expected []string) {
	if ls, err := ioutil.ReadDir(dir); err != nil {
		t.ErrorNow(err)
	} else {
		names := make([]string, 0, len(ls))
		for _, d := range ls { names = append(names, d.Name()) }
		t.CompareStringsUnsorted(names, expected)
	}
}

func LocalDirHostTest(test *testing.T, f func(*testutil.T, *LocalDirHost)) {
	t := (*testutil.T)(test)
	if dir, err := ioutil.TempDir("","battlref_test_repo_"); err != nil {
		t.ErrorNow(err)
	} else {
		defer os.RemoveAll(dir)
		local := NewLocalDirHost(dir).(*LocalDirHost)
		f(t, local)
	}
}

func TestInitLocalRepo(t *testing.T) {
	LocalDirHostTest(t, func (t *testutil.T, local *LocalDirHost) {
		t.CheckError(local.InitRepository("foo", "PublicKeyFoo"))
		repoURL := local.RepositoryURL("foo")
		if stat, err := os.Stat(repoURL); err != nil {
			t.ErrorNow(err)
		} else if ! stat.IsDir() {
			t.ErrorNowf("%s is not a directory", repoURL)
		} else {
			CheckDirectoryContent(t, repoURL, []string{"HEAD", "branches", "config", "description", "hooks", "info", "objects", "refs"})
		}
	})
}

func TestInitExitingLocalRepoFails(t *testing.T) {
	LocalDirHostTest(t, func (t *testutil.T, local *LocalDirHost) {
		t.CheckError(local.InitRepository("foo", "PublicKeyFoo"))
		if err := local.InitRepository("foo", "PublicKeyFoo"); err == nil {
			t.FailNow()
		}
	})
}

func TestForkLocalRepo(t *testing.T) {
	LocalDirHostTest(t, func (t *testutil.T, local *LocalDirHost) {
		t.CheckError(local.InitRepository("foo", "PublicKeyFoo"))
		t.CheckError(local.ForkRepository("foo", "bar", "PublicKeyFoo"))
		repoURL := local.RepositoryURL("bar")
		if stat, err := os.Stat(repoURL); err != nil {
			t.ErrorNow(err)
		} else if ! stat.IsDir() {
			t.ErrorNowf("%s is not a directory", repoURL)
		} else {
			CheckDirectoryContent(t, repoURL, []string{"HEAD", "branches", "config", "description", "hooks", "info", "objects", "refs"})
		}
	})
}

func TestDeleteLocalRepo(t *testing.T) {
	LocalDirHostTest(t, func (t *testutil.T, local *LocalDirHost) {
		t.CheckError(local.InitRepository("foo", "PublicKeyFoo"))
		t.CheckError(local.DeleteRepository("foo"))
		repoURL := local.RepositoryURL("foo")
		if _, err := os.Stat(repoURL); err == nil {
			t.FailNow()
		} else if ! os.IsNotExist(err) {
			t.ErrorNow(err)
		}
	})
}

var gitoliteHost = GitoliteHost { "git_test", "localhost" }

func GitoliteHostTest(test *testing.T, f func(*testutil.T, *GitoliteHost)) {
	t := (*testutil.T)(test)
	if gitoliteUser, err := user.Lookup(gitoliteHost.User); err != nil {
		switch err.(type) {
		case user.UnknownUserError:
			t.Skipf("%v, skipping gitolite tests", err)
			t.SkipNow()
		default: t.ErrorNow(err)
		}
	} else if err != nil {
		t.ErrorNow(err)
		fmt.Println(gitoliteUser, err)
	} else {
		f(t, &gitoliteHost)
	}
}

func TestInitGioliteRepo(t *testing.T) {
	GitoliteHostTest(t, func (t *testutil.T, host *GitoliteHost) {
	})
}

