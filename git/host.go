package git

import (
	"os"
	"fmt"
	"os/exec"
	"errors"
	"path/filepath"
)

type GitHost interface {
	InitRepository(name, publicKey string) error
	ForkRepository(source, fork, publicKey string) error
	DeleteRepository(name string) error
	RepositoryURL(name string) string
	ExternalRepositoryURL(name string) string
}

type LocalDirHost struct {
	Dir string
}

func NewLocalDirHost(dir string) GitHost {
	return &LocalDirHost{dir}
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

type GitoliteHost struct {
	User string
	Hostname string
}

func NewGitoliteHost(user, host string) GitHost {
	return &GitoliteHost{user, host}
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
