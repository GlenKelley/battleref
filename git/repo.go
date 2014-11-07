package git

import (
	"os"
	"log"
	"fmt"
	"bytes"
	"strings"
	"os/exec"
	"io/ioutil"
)

type Repository interface {
	AddFiles(files []string) error
	CommitFiles(files []string, message string) error
	Push() error
	Delete() error
	Dir() string
	Log() ([]string, error)
	Head() (string, error)
}

type Remote interface {
	CheckoutRepository(repoURL string) (Repository, error)
}

type TempRemote struct {
}

func RunCmd(cmd *exec.Cmd) error {
	bs := bytes.Buffer{}
	cmd.Stderr = &bs
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running %v %v: %v\n", cmd.Path, cmd.Args, string(bs.Bytes()))
		return err
	} else {
		return nil
	}
}

func CmdOutput(cmd *exec.Cmd) ([]byte, error) {
	bs := bytes.Buffer{}
	cmd.Stderr = &bs
	if output, err := cmd.Output(); err != nil {
		fmt.Printf("Error running %v %v: %v\n", cmd.Path, cmd.Args, string(bs.Bytes()))
		return nil, err
	} else {
		return output, nil
	}
}

func (r TempRemote) CheckoutRepository(repoURL string) (Repository, error) {
	if tempDir, err := ioutil.TempDir(os.TempDir(), "battleref"); err != nil {
		return nil, err
	} else {
		cmd := exec.Command("git","clone",repoURL,tempDir)
		if err := RunCmd(cmd); err != nil {
			if err2 := os.RemoveAll(tempDir); err2 != nil { log.Println(err2) }
			return nil, err
		} else {
			return &SimpleRepository{tempDir}, nil
		}
	}
}

type SimpleRepository struct {
	dir string
}

func (r *SimpleRepository) Dir() string {
	return r.dir
}

func (r *SimpleRepository) AddFiles(files []string) error {
	cmd := exec.Command("git", append([]string{"add"}, files ...) ...)
	cmd.Dir = r.dir
	return RunCmd(cmd)
}

func (r *SimpleRepository) CommitFiles(files []string, message string) error {
	cmd := exec.Command("git", append([]string{"commit","-m",message}, files ...) ...)
	cmd.Dir = r.dir
	return RunCmd(cmd)
}

func (r *SimpleRepository) Push() error {
	cmd := exec.Command("git","push","origin","master")
	cmd.Dir = r.dir
	return RunCmd(cmd)
}

func (r SimpleRepository) Delete() error {
	return os.RemoveAll(r.dir)
}

func (r SimpleRepository) Log() ([]string, error) {
	cmd := exec.Command("git","log","--pretty=%H")
	cmd.Dir = r.dir
	if output, err := CmdOutput(cmd); err != nil {
		return []string{}, err
	} else {
		return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
	}
}

func (r SimpleRepository) Head() (string, error) {
	cmd := exec.Command("git","rev-parse","HEAD")
	cmd.Dir = r.dir
	if output, err := CmdOutput(cmd); err != nil {
		return "", err
	} else {
		return string(output), err
	}
}


