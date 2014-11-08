package git

import (
	"os"
	"fmt"
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
}

func (g *GitoliteHost) InitRepository(name, publicKey string) error {
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


