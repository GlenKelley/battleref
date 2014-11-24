package git

import (
	"os"
	"fmt"
	"bytes"
	"os/exec"
	"errors"
	"path/filepath"
	"regexp"
	"io/ioutil"
)

type GitHost interface {
	InitRepository(name, publicKey string) error
	ForkRepository(source, fork, publicKey string) error
	DeleteRepository(name string) error
	RepositoryURL(name string) string
	ExternalRepositoryURL(name string) string
	Cleanup() error
}

type LocalDirHost struct {
	Dir string
	RemoveOnCleanup bool
}

var RemoteRegexp = regexp.MustCompile("\\w+@\\w+(.\\w+)+")

func CreateGitHost(path string) (GitHost, error) {
	if RemoteRegexp.MatchString(path) {
		return nil, errors.New("Not implemented")
	} else if path == ":temp:" {
		if tempDir, err := ioutil.TempDir(os.TempDir(), "battlecode"); err != nil {
			return nil, err
		} else {
			return &LocalDirHost{tempDir, true}, nil
		}
	} else {
		return &LocalDirHost{path, false}, nil
	}
}

func (g *LocalDirHost) InitRepository(name, publicKey string) error {
	repoURL := g.RepositoryURL(name)
	if _, err := os.Stat(repoURL); os.IsNotExist(err) {
		return exec.Command("git","init","--bare",repoURL).Run()
	} else if err != nil {
		return err
	} else {
		return errors.New(fmt.Sprintf("%v exists", repoURL))
	}
}

func (g *LocalDirHost) ForkRepository(source, fork, publicKey string) error {
	return exec.Command("git","clone","--bare",g.RepositoryURL(source),g.RepositoryURL(fork)).Run()
}

func (g *LocalDirHost) DeleteRepository(name string) error {
	return os.RemoveAll(g.RepositoryURL(name))
}

func (g *LocalDirHost) RepositoryURL(name string) string {
	return fmt.Sprintf("%s.git", filepath.Join(g.Dir, name))
}

func (g *LocalDirHost) ExternalRepositoryURL(name string) string {
	return fmt.Sprintf("%s.git", filepath.Join(g.Dir, name))
}

func (g *LocalDirHost) Cleanup() error {
	if g.RemoveOnCleanup {
		return os.RemoveAll(g.Dir)
	} else {
		return nil
	}
}

type GitoliteHost struct {
	User string
	Hostname string
	AdminKeyFile string
}

func (g *GitoliteHost) checkoutAdminRepo() (Repository, error) {
	if g.AdminKeyFile == "" {
		repo, err := TempRemote{}.CheckoutRepository(g.RepositoryURL("gitolite-admin"))
		return repo, err
	} else {
		repo, err := TempRemote{}.CheckoutRepositoryWithKeyFile(g.RepositoryURL("gitolite-admin"), g.AdminKeyFile)
		return repo, err
	}
}

func (g *GitoliteHost) InitRepository(name, publicKey string) error {
	if repo, err := g.checkoutAdminRepo(); err != nil {
		return err
	} else {
		defer repo.Delete()
		dir := repo.Dir()
		keyFile := filepath.Join(dir, "keydir", name + ".pub")
		confFile := filepath.Join(dir, "conf", "gitolite.conf")
		//TODO: check edge case where user key is duplicated
		files := []string{keyFile, confFile}
		confLine := fmt.Sprintf("\nrepo %v\n    RW+    =   webserver %v", name, name)
		if err := ioutil.WriteFile(keyFile, []byte(publicKey), 0644); err != nil {
			return err
		} else if conf, err := os.OpenFile(confFile, os.O_RDWR, 0644); err != nil {
			return err
		} else if _, err := conf.Seek(0, os.SEEK_END); err != nil {
			return err
		} else if _, err := conf.WriteString(confLine); err != nil {
			return err
		} else if err := repo.AddFiles(files); err != nil {
			return err
		} else if err := repo.CommitFiles(files, fmt.Sprintf("added repo %v", name)); err != nil {
			return err
		} else if err := repo.Push(); err != nil {
			return err
		}
	}
	return nil
}

func (g *GitoliteHost) ForkRepository(source, fork, publicKey string) error {
	return nil
}

func (g *GitoliteHost) DeleteRepository(name string) error {
	return nil
}

func (g *GitoliteHost) RepositoryURL(name string) string {
	return fmt.Sprintf("%s@%s:/%s.git", g.User, g.Hostname, name)
}

func (g *GitoliteHost) ExternalRepositoryURL(name string) string {
	return fmt.Sprintf("%s@%s:/%s.git", g.User, g.Hostname, name)
}

func (g *GitoliteHost) Cleanup() error {
	return nil
}

func (g *GitoliteHost) Reset() error {
	if repo, err := g.checkoutAdminRepo(); err != nil {
		return err
	} else {
		defer repo.Delete()
		cmd := exec.Command("git","rev-list","--max-parents=0","HEAD")
		cmd.Dir = repo.Dir()
		if initalCommit, err := CmdOutput(cmd); err != nil {
			return err
		} else if err := repo.HardReset(string(bytes.TrimSpace(initalCommit))); err != nil {
			return err
		} else if err := repo.ForcePush(); err != nil {
			return err
		}
	}
	return nil

}


