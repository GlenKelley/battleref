package git

import (
	"strings"
	"os/exec"
	"testing"
	"io/ioutil"
	"os"
	"os/user"
	"github.com/GlenKelley/battleref/testing"
)

func TestPublicKeyParsing(t *testing.T) {
	padding := strings.Repeat("1",256)
	for _, key := range []string{
			"ssh-rsa AAAA1234"+padding,
			"ssh-rsa AAAA1234"+padding+" email@address.com",
			"ssh-rsa AAAA1234"+padding+" other text",
			"ssh-rsa AAAA1234"+padding+"\n",
			"ssh-rsa AAAA1234"+padding+" other text\n",
		} {
		if !PublicKeyRegex.MatchString(key) {
			t.Errorf("'%v' is not a public key", key)
		}
	}
	for _, key := range []string{
			"ssh-rsa AAAA1234",
			"ssh-rsa AAAA1234"+" email@address.com",
			"ssh-rsa AAAA12!34"+" email@address.com",
		} {
		if PublicKeyRegex.MatchString(key) {
			t.Errorf("Expected failure for invalid key '%v'", key)
		}
	}
}

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
	if local, err := CreateGitHost(":temp:", nil); err != nil {
		t.ErrorNow(err)
	} else {
		defer local.Cleanup()
		f(t, local.(*LocalDirHost))
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

var gitoliteTestConf = GitoliteConf {
	"localhost",
	"foobar",
	"git-test",
	".ssh/webserver",
	".ssh/git",
}

func GitoliteHostTest(test *testing.T, f func(*testutil.T, *GitoliteHost)) {
	t := (*testutil.T)(test)
	conf := gitoliteTestConf
	conf.AdminKey = testutil.PathRelativeToUserHome(t, conf.AdminKey)
	conf.SSHKey = testutil.PathRelativeToUserHome(t, conf.SSHKey)
	if host, err := CreateGitoliteHost(conf); err != nil {
		t.ErrorNow(err)
	} else if _, err := user.Lookup(host.User); err != nil {
		switch err.(type) {
		case user.UnknownUserError:
			t.Skipf("%v, skipping gitolite tests", err)
			t.SkipNow()
			default: t.ErrorNow(err)
		}
	} else if err := host.Reset(); err != nil {
		t.ErrorNow(err)
	} else {
		f(t, host)
	}
}

func createRepo(host *GitoliteHost, name string) error {
	if _, publicKey, err := testutil.CreateKeyPair(); err != nil {
		return err
	} else {
		return host.InitRepository(name, publicKey)
	}
}

func TestInitGitoliteRepo(t *testing.T) {
	GitoliteHostTest(t, func (t *testutil.T, host *GitoliteHost) {
		if privateKey, publicKey, err := testutil.CreateKeyPair(); err != nil {
		        t.ErrorNow(err)
		} else if file, err := ioutil.TempFile(os.TempDir(), "battlecode_private_key"); err != nil {
		        t.ErrorNow(err)
		} else if _, err := file.WriteString(privateKey); err != nil {
			t.ErrorNow(err)
		} else {
			file.Close()
			defer os.Remove(file.Name())
			if err := host.InitRepository("foo", publicKey); err != nil {
				t.ErrorNow(err)
			}
			defer host.DeleteRepository("foo")
			repoURL := host.RepositoryURL("foo")
			if repo, err := (TempRemote{}).CheckoutRepositoryWithKeyFile(repoURL, file.Name()); err != nil {
				t.ErrorNow(err)
			} else {
				defer repo.Delete()
				CheckDirectoryContent(t, repo.Dir(), []string{".git"})
			}
		}
	})
}

func TestDeleteGitoliteRepo(t *testing.T) {
	GitoliteHostTest(t, func (t *testutil.T, host *GitoliteHost) {
		t.CheckError(createRepo(host, "foo"))
		t.CheckError(host.DeleteRepository("foo"))
		cmd := exec.Command("ssh", "-v", "-v", "-i", host.SSHKey, host.User+"@"+host.InternalHostname, "[[ ! -d repositories/foo.git ]]")
		t.CheckError(cmd.Run())
	})
}

func TestDeleteGitoliteDuplicateRepo(t *testing.T) {
	GitoliteHostTest(t, func (t *testutil.T, host *GitoliteHost) {
		if privateKey, publicKey, err := testutil.CreateKeyPair(); err != nil {
			t.ErrorNow(err)
		} else if file, err := ioutil.TempFile(os.TempDir(), "battlecode_private_key"); err != nil {
		        t.ErrorNow(err)
		} else if _, err := file.WriteString(privateKey); err != nil {
			t.ErrorNow(err)
		} else {
			file.Close()
			defer os.Remove(file.Name())
			if err := host.InitRepository("foo", publicKey); err != nil {
				t.ErrorNow(err)
			}

			repoURL := host.RepositoryURL("foo")
			if repo, err := (TempRemote{}).CheckoutRepositoryWithKeyFile(repoURL, file.Name()); err != nil {
				t.ErrorNow(err)
			} else {
				defer repo.Delete()
				CheckDirectoryContent(t, repo.Dir(), []string{".git"})
			}

			defer host.DeleteRepository("foo")
			if err := host.InitRepository("bar", publicKey); err != nil {
				t.ErrorNow(err)
			}
			defer host.DeleteRepository("bar")

			repoURL2 := host.RepositoryURL("bar")
			if repo, err := (TempRemote{}).CheckoutRepositoryWithKeyFile(repoURL2, file.Name()); err != nil {
				t.ErrorNow(err)
			} else {
				defer repo.Delete()
				CheckDirectoryContent(t, repo.Dir(), []string{".git"})
			}

			if repo, err := (TempRemote{}).CheckoutRepositoryWithKeyFile(repoURL, file.Name()); err == nil {
				defer repo.Delete()
				t.ErrorNow("Expected failure")
			}
		}
	})
}


