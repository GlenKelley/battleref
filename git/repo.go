package git

import (
	"os"
	"log"
	"fmt"
	"runtime/debug"
	"bytes"
	"strings"
	"os/exec"
	"io/ioutil"
)

type Repository interface {
	AddFiles(files []string) error
	CommitFiles(files []string, message string) error
	Push() error
	ForcePush() error
	Delete() error
	Dir() string
	Log() ([]string, error)
	Head() (string, error)
	HardReset(commit string) (error)
}

type Remote interface {
	CheckoutRepository(repoURL string) (Repository, error)
	CheckoutRepositoryWithKeyFile(repoURL string, privateKeyFile string) (Repository, error)
}

type TempRemote struct {
}

func DebugCmd(cmd *exec.Cmd) error {
	bs1 := bytes.Buffer{}
	bs2 := bytes.Buffer{}
	cmd.Stdout = &bs1
	cmd.Stderr = &bs2
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running %v %v %v:\n%v\n%v\n", cmd.Path, cmd.Args, cmd.Env, string(bs1.Bytes()), string(bs2.Bytes()))
		debug.PrintStack()
		return err
	} else {
		return nil
	}
}

func RunCmd(cmd *exec.Cmd) error {
	bs1 := bytes.Buffer{}
	bs2 := bytes.Buffer{}
	cmd.Stdout = &bs1
	cmd.Stderr = &bs2
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running %v %v %v: %v\n", cmd.Path, cmd.Args, cmd.Env, string(bs1.Bytes()), string(bs2.Bytes()))
		debug.PrintStack()
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
		debug.PrintStack()
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
			return &SimpleRepository{tempDir, ""}, nil
		}
	}
}

func CreateGitSSHWrapper(privateKeyFile string) (string, error) {
	if file, err := ioutil.TempFile(os.TempDir(), "battleref_ssh_wrapper"); err != nil {
		return "", err
	} else if err := file.Chmod(0755); err != nil {
		os.Remove(file.Name())
		return "", err
	} else if _, err := file.WriteString(fmt.Sprintf("ssh -i %v $@", privateKeyFile)); err != nil {
		os.Remove(file.Name())
		return "", err
	} else {
		file.Close()
		return file.Name(), nil
	}
}

func (r TempRemote) CheckoutRepositoryWithKeyFile(repoURL string, privateKeyFile string) (Repository, error) {
	if tempDir, err := ioutil.TempDir(os.TempDir(), "battleref"); err != nil {
		return nil, err
	} else if sshWrapper, err := CreateGitSSHWrapper(privateKeyFile); err != nil {
		os.RemoveAll(tempDir)
		return nil, err
	} else {
		cmd := exec.Command("git","clone",repoURL,tempDir)
		repo := SimpleRepository{tempDir, sshWrapper}
		repo.setWrapper(cmd)
		if err := RunCmd(cmd); err != nil {
			os.Remove(sshWrapper)
			os.RemoveAll(tempDir)
			return nil, err
		} else {
			return &repo, nil
		}
	}
}

type SimpleRepository struct {
	dir string
	sshWrapper string
}

func (r *SimpleRepository) setWrapper(cmd *exec.Cmd) {
	if r.sshWrapper != "" {
		cmd.Env = append(os.Environ(), "GIT_SSH=" + r.sshWrapper)
	}
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
	cmd0 := exec.Command("git", "status")
	cmd0.Dir = r.dir
	DebugCmd(cmd0)
	cmd := exec.Command("git", append([]string{"commit","-m",message}, files ...) ...)
	cmd.Dir = r.dir
	return RunCmd(cmd)
}

func (r *SimpleRepository) Push() error {
	cmd := exec.Command("git","push","origin","master")
	cmd.Dir = r.dir
	r.setWrapper(cmd)
	return RunCmd(cmd)
}

func (r *SimpleRepository) ForcePush() error {
	cmd := exec.Command("git","push","--force","origin","master")
	cmd.Dir = r.dir
	r.setWrapper(cmd)
	return RunCmd(cmd)
}

func (r SimpleRepository) Delete() error {
	if (r.sshWrapper != "") {
		os.Remove(r.sshWrapper)
	}
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
		return string(bytes.TrimSpace(output)), err
	}
}

func (r SimpleRepository) HardReset(commit string) error {
	cmd := exec.Command("git","reset",commit,"--hard")
	cmd.Dir = r.dir
	return RunCmd(cmd)
}

