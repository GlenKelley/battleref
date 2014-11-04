package arena

import (
	"time"
	"strings"
	"bytes"
	"io/ioutil"
	"os/exec"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"go/build"
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

func ErrorNow(t *testing.T, arg ... interface{}) {
	t.Error(arg ... )
	trace := make([]byte, 1024)
	count := runtime.Stack(trace, false)
	t.Errorf("Stack of %d bytes: %s", count, trace)
	t.FailNow()
}

func Check(t *testing.T, err error) {
	if err != nil { ErrorNow(t, err) }
}

func RunCommand(t *testing.T, cmd *exec.Cmd) {
	if bs, err := cmd.CombinedOutput(); err != nil {
		t.Error(string(bs))
		ErrorNow(t, err)
	}
}

func TestRunMatch(t *testing.T) {
	gitDir, err := ioutil.TempDir(os.TempDir(), "samplePlayer")
	if err != nil { ErrorNow(t, err) }
	defer os.RemoveAll(gitDir)

	sourceFile := filepath.Join(gitDir, "RobotPlayer.java")
	Check(t, ioutil.WriteFile(sourceFile, SamplePlayer, os.ModePerm))

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
		ErrorNow(t, err)
	}
	commitHash := strings.TrimSpace(string(bs))

	pkg, err := build.ImportDir("github.com/GlenKelley/battleref/arena", build.FindOnly)
	if err != nil {
		ErrorNow(t, err)
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
		ErrorNow(t, err)
	} else if finished != finishedTime {
		ErrorNow(t, err)
	} else if result.Winner != WinnerA {
		ErrorNow(t, err)
	} else if result.Reason != ReasonTie {
		ErrorNow(t, err)
	} else if len(result.Replay) == 0 {
		ErrorNow(t, err)
	}
}

