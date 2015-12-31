package git

import (
	"os"
	"fmt"
	"bytes"
	"os/exec"
	"os/user"
	"encoding/json"
	"errors"
	"path/filepath"
	"regexp"
	"io/ioutil"
)

type GitHost interface {
	InitRepository(name, publicKey string) error
	CloneRepository(remote Remote, name string) (Repository, error)
	ForkRepository(source, fork, publicKey string) error
	DeleteRepository(name string) error
	RepositoryURL(name string) string
	ExternalRepositoryURL(name string) string
	Validate() error
	Cleanup() error
	Reset() error
}

type LocalDirHost struct {
	Dir string
	RemoveOnCleanup bool
}

var (
	RemoteRegexp = regexp.MustCompile("\\w+@\\w+(.\\w+)+")
	PublicKeyRegex = regexp.MustCompile("^(ssh-(r|d)sa AAAA[0-9A-Za-z+/]{256,}[=]{0,3})\\s*.*\\n?$")    //SSH public key
)


func CreateGitHost(hostType string, conf map[string]string) (GitHost, error) {
	switch hostType {
	case ":temp:":
		if tempDir, err := ioutil.TempDir(os.TempDir(), "battlecode_temp_host"); err != nil {
			return nil, err
		} else {
			return &LocalDirHost{tempDir, true}, nil
		}
	case ":gitolite:":
		var gitoliteConf GitoliteConf
		if err := MarshalMap(conf, &gitoliteConf); err != nil {
			return nil, err
		} else if host, err := CreateGitoliteHost(gitoliteConf); err != nil {
			return nil, err
		} else {
			return host, err
		}
	default: return nil, errors.New(fmt.Sprintf("Unkown host type %v", hostType))
	}
}

func MarshalMap(conf map[string]string, v interface{}) error {
	if bs, err := json.Marshal(conf); err != nil {
		return err
	} else {
		return json.Unmarshal(bs, v)
	}
}

type GitoliteConf struct {
	InternalHostname string `json:"internal_hostname"`
	ExternalHostname string `json:"external_hostname"`
	User		 string `json:"user"`
	AdminKey	 string `json:"admin_key"`
	SSHKey		 string `json:"ssh_key"`
}

func CreateGitoliteHost(conf GitoliteConf) (*GitoliteHost, error) {
	if conf.InternalHostname == "" { return nil, errors.New("Gitolite host missing internal hostname property.") }
	if conf.ExternalHostname == "" { return nil, errors.New("Gitolite host missing external hostname property.") }
	if conf.User == "" { return nil, errors.New("Gitolite host missing user property.") }
	if conf.AdminKey == "" { return nil, errors.New("Gitolite host missing admin key property.") }
	if conf.SSHKey == "" { return nil, errors.New("Gitolite host missing ssh key property.") }
	return &GitoliteHost{conf}, nil
}

func (g *LocalDirHost) InitRepository(name, publicKey string) error {
	repoURL := g.RepositoryURL(name)
	if _, err := os.Stat(repoURL); os.IsNotExist(err) {
		return RunCmd(exec.Command("git","init","--bare",repoURL))
	} else if err != nil {
		return err
	} else {
		return errors.New(fmt.Sprintf("%v exists", repoURL))
	}
}

func (g *LocalDirHost) CloneRepository(remote Remote, name string) (Repository, error) {
	repo, err := remote.CheckoutRepository(g.RepositoryURL(name))
	return repo, err
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

func (g *LocalDirHost) Validate() error {
	if info, err := os.Stat(g.Dir); err != nil {
		return err
	} else if !info.IsDir() {
		return errors.New("Local repo directory does not exist")
	}
	return nil
}

func (g *LocalDirHost) Cleanup() error {
	if g.RemoveOnCleanup {
		return os.RemoveAll(g.Dir)
	} else {
		return nil
	}
}

func (g *LocalDirHost) Reset() error {
	if err := os.RemoveAll(g.Dir); err != nil {
		return err
	} else {
		return os.Mkdir(g.Dir, 0644)
	}
}

type GitoliteHost struct {
	GitoliteConf
}

func (g *GitoliteHost) IsReservedKey(publicKey string) (bool, error) {
	for _, keyFile := range []string{g.AdminKey, g.SSHKey} {
		if keyFile == "" {
			return false, nil
		} else if key, err := ioutil.ReadFile(keyFile + ".pub"); err != nil {
			return false, err
		} else if match := PublicKeyRegex.FindStringSubmatch(string(key)); match == nil {
			return false, errors.New(fmt.Sprintf("Invalid Key Format %v", string(key)))
		} else if match[1] == publicKey {
			return true, nil
		}
	}
	return false, nil
}

func (g *GitoliteHost) checkoutAdminRepo() (Repository, error) {
	repo, err := g.CloneRepository(TempRemote{}, "gitolite-admin")
	return repo, err
}

func (g *GitoliteHost) InitRepository(name, publicKey string) error {
	if isReserved, err := g.IsReservedKey(publicKey); err != nil {
		return err
	} else if isReserved {
		return errors.New("Reserved Key")
	}
	if repo, err := g.checkoutAdminRepo(); err != nil {
		return err
	} else {
		defer repo.Delete()
		dir := repo.Dir()
		keyFile := filepath.Join(dir, "keydir", name + ".pub")
		confFile := filepath.Join(dir, "conf", "gitolite.conf")
		//TODO: check edge case where user key is duplicated
		files := []string{keyFile, confFile}
		confLine := fmt.Sprintf("repo %v\n    RW+    =   webserver %v\n", name, name)
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

func (g *GitoliteHost) CloneRepository(remote Remote, name string) (Repository, error) {
	if g.AdminKey == "" {
		repo, err := remote.CheckoutRepository(g.RepositoryURL(name))
		return repo, err
	} else {
		repo, err := remote.CheckoutRepositoryWithKeyFile(g.RepositoryURL(name), g.AdminKey)
		return repo, err
	}
}

func (g *GitoliteHost) ForkRepository(source, fork, publicKey string) error {
	return errors.New("Not implemented.")
}

func (g *GitoliteHost) DeleteRepository(name string) error {
	cmd := exec.Command("ssh", "-i", g.SSHKey, fmt.Sprintf("%v@%v", g.User, g.InternalHostname), fmt.Sprintf("rm -rf repositories/%v.git", name))
	return cmd.Run()
}

func (g *GitoliteHost) RepositoryURL(name string) string {
	return fmt.Sprintf("%s@%s:%s.git", g.User, g.InternalHostname, name)
}

func (g *GitoliteHost) ExternalRepositoryURL(name string) string {
	return fmt.Sprintf("%s@%s:%s.git", g.User, g.ExternalHostname, name)
}

func (g *GitoliteHost) Validate() error {
	if _, err := user.Lookup(g.User); err != nil {
		return err
	}
	if _, err := os.Stat(g.AdminKey); err != nil {
		return err
	}
	if _, err := os.Stat(g.SSHKey); err != nil {
		return err
	}
	return nil
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
		cmd = exec.Command("ssh", "-i", g.SSHKey, fmt.Sprintf("%v@%v", g.User, g.InternalHostname), "set -x ; find repositories -maxdepth 1 -mindepth 1 -type d | grep -v '/gitolite-admin.git$' | grep -v '/testing.git$' | xargs rm -rf")
		return cmd.Run()

	}
	return nil

}


