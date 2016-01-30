package arena

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"time"
)

const (
	ReasonTie     = "TIE"
	ReasonVictory = "VICTORY"

	WinnerA = "A"
	WinnerB = "B"
)

type MatchProperties struct {
	MapName     string
	MapSource   io.Reader
	Category    string
	PlayerRepo1 string
	PlayerRepo2 string
	Commit1     string
	Commit2     string
}

type Arena interface {
	RunMatch(properties MatchProperties, clock func() time.Time) (time.Time, MatchResult, error)
}

type MatchResult struct {
	Winner string `json:"winner"`
	Reason string `json:"reason"`
	Replay []byte `json:"-"`
}

type LocalArena struct {
	ResourceDir string
}

func (a LocalArena) RunMatch(p MatchProperties, clock func() time.Time) (time.Time, MatchResult, error) {
	var result MatchResult
	tarFile := "battlecode.tar"
	mapFile, err := ioutil.TempFile(os.TempDir(), p.MapName)
	if err != nil {
		return clock(), result, err
	}
	defer os.Remove(mapFile.Name())
	if file, err := os.Create(mapFile.Name()); err != nil {
		return clock(), result, err
	} else if _, err := io.Copy(file, p.MapSource); err != nil {
		return clock(), result, err
	} else {
		file.Close()
	}

	tempDir, err := ioutil.TempDir("", "battlecode")
	if err != nil {
		return clock(), result, err
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("./runMatch.sh",
		"-r", tarFile,
		"-d", tempDir,
		"-p", p.PlayerRepo1,
		"-P", p.PlayerRepo2,
		"-c", p.Commit1,
		"-C", p.Commit2,
		"-m", p.MapName,
		"-M", mapFile.Name(),
	)
	cmd.Dir = filepath.Join(a.ResourceDir, p.Category)
	buffer := bytes.Buffer{}
	cmd.Stderr = &buffer
	log.Println(cmd)
	if out, err := cmd.Output(); err != nil {
		debug.PrintStack()
		log.Println("runMatch Error: ", string(buffer.Bytes()))
		log.Println("runMatch Output: ", string(out))
		return clock(), result, err
	} else if err := json.NewDecoder(bytes.NewReader(out)).Decode(&result); err != nil {
		log.Println("runMatch Error: ", string(buffer.Bytes()))
		return clock(), result, err
	} else if bs, err := ioutil.ReadFile(filepath.Join(tempDir, "replay.xml.gz")); err != nil {
		log.Println("runMatch Error: ", string(buffer.Bytes()))
		return clock(), result, err
	} else {
		log.Println("runMatch Error: ", string(buffer.Bytes()))
		log.Println("runMatch Output: ", string(out))
		log.Println("bs length", len(bs))
		result.Replay = bs
		return clock(), result, nil
	}
}

func NewArena(resourceDir string) Arena {
	return LocalArena{resourceDir}
}

type DummyArena struct {
	Finish time.Time
	Result MatchResult
	Err    error
}

func (a DummyArena) RunMatch(p MatchProperties, clock func() time.Time) (time.Time, MatchResult, error) {
	return a.Finish, a.Result, a.Err
}
