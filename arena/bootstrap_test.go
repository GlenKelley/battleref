package arena

import (
	"github.com/GlenKelley/battleref/testing"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBootstrap2016(test *testing.T) {
	t := (*testutil.T)(test)
	gitDir, err := ioutil.TempDir(os.TempDir(), "sampleplayer")
	if err != nil {
		t.ErrorNow(err)
	}
	defer os.RemoveAll(gitDir)

	pkg, err := build.ImportDir("github.com/GlenKelley/battleref/arena", build.FindOnly)
	t.CheckError(err)
	resourceDir := filepath.Join(build.Default.GOPATH, "src", pkg.Dir, "internal", "categories")
	bootstrap := MinimalBootstrap{resourceDir}
	t.CheckError(bootstrap.PopulateRepository("glen", gitDir, "battlecode2016"))

	cmd := exec.Command("ant", "update")
	cmd.Dir = gitDir
	RunCommand(t, cmd)

	cmd = exec.Command("ant", "test")
	cmd.Dir = gitDir
	RunCommand(t, cmd)
}
