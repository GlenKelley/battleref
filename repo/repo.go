package repo

import (
	"os"
	"log"
	"fmt"
	"os/exec"
	"io/ioutil"
	"path/filepath"
)

type GitServer interface {
	InitRepository(name, publicKey string) error
	ForkRepository(source, fork, publicKey string) error
	DeleteRepository(name string) error
	RepositoryURL(name string) string
}

type LocalDir struct {
	Dir string
}

func NewLocal(dir string) GitServer {
	return &LocalDir{dir}
}

func (g *LocalDir) InitRepository(name, publicKey string) error {
	return exec.Command("git","init","--bare",g.RepositoryURL(name)).Run()
}

func (g *LocalDir) ForkRepository(source, fork, publicKey string) error {
	return exec.Command("git","clone","--bare",g.RepositoryURL(source),g.RepositoryURL(fork)).Run()
}

func (g *LocalDir) DeleteRepository(name string) error {
	return os.RemoveAll(g.RepositoryURL(name))
}

func (g *LocalDir) RepositoryURL(name string) string {
	return fmt.Sprintf("%s.git", filepath.Join(g.Dir, name))
}

type Gitolite struct {
	User string
	Host string
}

func NewGitolite(user, host string) GitServer {
	return &Gitolite{user, host}
}

func (g *Gitolite) InitRepository(name, publicKey string) error {
	return nil
}

func (g *Gitolite) ForkRepository(source, fork, publicKey string) error {
	return nil
}

func (g *Gitolite) DeleteRepository(name string) error {
	return nil
}

func (g *Gitolite) RepositoryURL(name string) string {
	return fmt.Sprintf("%s@%s:/%s.git", g.User, g.Host, name)
}

type Repository interface {
	CommitFiles(files []string, message string) error
	Push() error
	Delete() error
	RepoDir() string
}

type Remote interface {
	CheckoutRepository(repoURL string) (Repository, error)
}

type TempRemote struct {
}

func (r TempRemote) CheckoutRepository(repoURL string) (Repository, error) {
	if tempDir, err := ioutil.TempDir(os.TempDir(), "battleref"); err != nil {
		return nil, err
	} else if err := exec.Command("git","clone","-m",repoURL,tempDir).Run(); err != nil {
		if err2 := os.RemoveAll(tempDir); err2 != nil { log.Println(err2) }
		return nil, err
	} else {
		return &SimpleRepository{tempDir}, nil
	}
}


type SimpleRepository struct {
	Dir string
}

func (r *SimpleRepository) RepoDir() string {
	return r.Dir
}

func (r *SimpleRepository) CommitFiles(files []string, message string) error {
	cmd := exec.Command("git", append([]string{"commit","-m",message}, files ...) ...)
	cmd.Dir = r.Dir
	return cmd.Run()
}

func (r *SimpleRepository) Push() error {
	cmd := exec.Command("git","push","origin","master")
	cmd.Dir = r.Dir
	return cmd.Run()
}

func (r SimpleRepository) Delete() error {
	return os.RemoveAll(r.Dir)
}
