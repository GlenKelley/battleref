package git

import (
	"github.com/GlenKelley/battleref/testing"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckoutRepository(t *testing.T) {
	LocalDirHostTest(t, func(t *testutil.T, host *LocalDirHost) {
		t.CheckError(host.InitRepository("foo", nil, nil))
		repoURL := host.RepositoryURL("foo")
		if repo, err := (TempRemote{}).CheckoutRepository(repoURL); err != nil {
			t.ErrorNow(err)
		} else {
			defer repo.Delete()
			CheckDirectoryContent(t, repo.Dir(), []string{".git"})
		}

	})
}

func TestCommitFiles(t *testing.T) {
	LocalDirHostTest(t, func(t *testutil.T, host *LocalDirHost) {
		t.CheckError(host.InitRepository("foo", nil, nil))
		repoURL := host.RepositoryURL("foo")
		if repo, err := (TempRemote{}).CheckoutRepository(repoURL); err != nil {
			t.ErrorNow(err)
		} else {
			defer repo.Delete()
			t.CheckError(ioutil.WriteFile(filepath.Join(repo.Dir(), "foo.txt"), []byte("hello"), os.ModePerm))
			t.CheckError(repo.AddFiles([]string{"foo.txt"}))
			t.CheckError(repo.CommitFiles([]string{"foo.txt"}, "commit message"))
			if log, err := repo.Log(); err != nil {
				t.ErrorNow(err)
			} else if len(log) != 1 {
				t.FailNow()
			}
		}

	})
}

func TestPush(t *testing.T) {
	LocalDirHostTest(t, func(t *testutil.T, host *LocalDirHost) {
		t.CheckError(host.InitRepository("foo", nil, nil))
		repoURL := host.RepositoryURL("foo")
		var head string
		if repo, err := (TempRemote{}).CheckoutRepository(repoURL); err != nil {
			t.ErrorNow(err)
		} else {
			defer repo.Delete()
			t.CheckError(ioutil.WriteFile(filepath.Join(repo.Dir(), "foo.txt"), []byte("hello"), os.ModePerm))
			t.CheckError(repo.AddFiles([]string{"foo.txt"}))
			t.CheckError(repo.CommitFiles([]string{"foo.txt"}, "commit message"))
			t.CheckError(repo.Push())
			if h, err := repo.Head(); err != nil {
				t.ErrorNow(err)
			} else {
				head = h
			}
		}
		if repo, err := (TempRemote{}).CheckoutRepository(repoURL); err != nil {
			t.ErrorNow(err)
		} else {
			defer repo.Delete()
			if head2, err := repo.Head(); err != nil {
				t.ErrorNow(err)
			} else if head != head2 {
				t.ErrorNowf("Expected <%v> != Actual <%v>", head, head2)
			}
		}

	})
}

func TestDeleteFiles(t *testing.T) {
	LocalDirHostTest(t, func(t *testutil.T, host *LocalDirHost) {
		t.CheckError(host.InitRepository("foo", nil, nil))
		repoURL := host.RepositoryURL("foo")
		if repo, err := (TempRemote{}).CheckoutRepository(repoURL); err != nil {
			t.ErrorNow(err)
		} else {
			defer repo.Delete()
			t.CheckError(ioutil.WriteFile(filepath.Join(repo.Dir(), "foo.txt"), []byte("hello"), os.ModePerm))
			t.CheckError(os.Mkdir(filepath.Join(repo.Dir(), "bar"), os.ModeDir|0755))
			t.CheckError(ioutil.WriteFile(filepath.Join(repo.Dir(), "bar", "moo.txt"), []byte("world"), os.ModePerm))
			t.CheckError(repo.AddFiles([]string{"foo.txt", "bar"}))
			t.CheckError(repo.CommitFiles(nil, "commit message"))
			if log, err := repo.Log(); err != nil {
				t.ErrorNow(err)
			} else if len(log) != 1 {
				t.FailNow()
			}

			CheckDirectoryContent(t, repo.Dir(), []string{".git", "foo.txt", "bar"})
			CheckDirectoryContent(t, filepath.Join(repo.Dir(), "bar"), []string{"moo.txt"})
			//Delete
			t.CheckError(repo.DeleteFiles([]string{"foo.txt"}))
			t.CheckError(repo.DeleteFiles([]string{"bar"}))
			if err := repo.DeleteFiles([]string{"other"}); err == nil {
				t.ErrorNow("Expected error")
			}
			CheckDirectoryContent(t, repo.Dir(), []string{".git"})
			t.CheckError(repo.CommitFiles([]string{}, "commit message"))
			if log, err := repo.Log(); err != nil {
				t.ErrorNow(err)
			} else if len(log) != 2 {
				t.FailNow()
			}
		}

	})
}
