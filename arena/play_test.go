package arena

import (
	"time"
	"strings"
	"bytes"
	"io/ioutil"
	"os/exec"
	"os"
	"path/filepath"
	"testing"
	"go/build"
	"github.com/GlenKelley/battleref/testing"
)

var SampleMap = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<map height="21" width="20">
	<game seed="0" rounds="1"/>
	<symbols>
		<symbol terrain="NORMAL" type="TERRAIN" character="_"/>
		<symbol terrain="VOID" type="TERRAIN" character="v"/>
		<symbol terrain="ROAD" type="TERRAIN" character="#"/>
		<symbol team="A" type="HQ" character="a"/>
		<symbol team="B" type="HQ" character="b"/>
	</symbols>
	<data>
<![CDATA[
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 a0 _0 _0 _0 _0 b0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
_0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0 _0
]]></data>
</map>`)

var SamplePlayer = []byte(`package samplePlayer;
import battlecode.common.RobotController;
public class RobotPlayer {
	public static void run(RobotController rc) {
		while (true) {
			rc.yield();
		}
	}
}`)

func RunCommand(t *testutil.T, cmd *exec.Cmd) {
	if bs, err := cmd.CombinedOutput(); err != nil {
		t.Error(string(bs))
		t.ErrorNow(err)
	}
}

func TestRunMatch(test *testing.T) {
	t := (*testutil.T)(test)
	gitDir, err := ioutil.TempDir(os.TempDir(), "samplePlayer")
	if err != nil { t.ErrorNow(err) }
	defer os.RemoveAll(gitDir)

	sourceFile := filepath.Join(gitDir, "RobotPlayer.java")
	t.CheckError(ioutil.WriteFile(sourceFile, SamplePlayer, os.ModePerm))

	cmd := exec.Command("git","init")
	cmd.Dir = gitDir
	RunCommand(t, cmd)

	cmd = exec.Command("git","add",sourceFile)
	cmd.Dir = gitDir
	RunCommand(t, cmd)

	cmd = exec.Command("git","commit","-m","init commit")
	cmd.Dir = gitDir
	RunCommand(t, cmd)

	cmd = exec.Command("git","clone","--bare","./","samplePlayer.git")
	cmd.Dir = gitDir
	RunCommand(t, cmd)

	cmd = exec.Command("git","log","-n1","--pretty=%H")
	cmd.Dir = gitDir
	bs, err := cmd.Output()
	if err != nil {
		t.ErrorNow(err)
	}
	commitHash := strings.TrimSpace(string(bs))

	pkg, err := build.ImportDir("github.com/GlenKelley/battleref/arena", build.FindOnly)
	if err != nil {
		t.ErrorNow(err)
	}
	packageDir := filepath.Join(build.Default.GOPATH, "src", pkg.Dir)
	arena := NewArena(packageDir)
	finishedTime := time.Now()
	if finished, result, err := arena.RunMatch(MatchProperties{
		"sampleMap",
		bytes.NewReader(SampleMap),
		"CategoryFoo",
		filepath.Join(gitDir, "samplePlayer.git"),
		filepath.Join(gitDir, "samplePlayer.git"),
		commitHash,
		commitHash,
	}, func()time.Time{ return finishedTime }); err != nil {
		t.ErrorNow(err)
	} else if finished != finishedTime {
		t.ErrorNow(err)
	} else if result.Winner != WinnerA {
		t.ErrorNow(err)
	} else if result.Reason != ReasonTie {
		t.ErrorNow(err)
	} else if len(result.Replay) == 0 {
		t.ErrorNow(err)
	}
}

