package arena

import (
	"log"
	"os"
	"os/exec"
	"time"
	"io"
	"bytes"
	"io/ioutil"
	"runtime/debug"
	"encoding/json"
)

const (
	ReasonTie = "TIE"
	ReasonVictory = "VICTORY"

	WinnerA = "A"
	WinnerB = "B"
)

type MatchProperties struct {
	MapName string
	MapSource io.Reader
	Category string
	PlayerRepo1 string
	PlayerRepo2 string
	Commit1 string
	Commit2 string
}

type Arena interface {
	RunMatch(properties MatchProperties, clock func()time.Time) (time.Time, MatchResult, error)
}

type MatchResult struct {
	Winner string `json:"winner"`
	Reason string `json:"reason"`
	Replay string `json:"replay"`
}

type LocalArena struct {
	ResourceDir string
}

func (a LocalArena) RunMatch(p MatchProperties, clock func()time.Time) (time.Time, MatchResult, error) {
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
		"-R",
	)
	cmd.Dir = a.ResourceDir
	buffer := bytes.Buffer{}
	cmd.Stderr = &buffer
	if out, err := cmd.Output(); err != nil {
		debug.PrintStack()
		log.Println("runMatch Error: ", string(buffer.Bytes()))
		log.Println("runMatch Output: ", string(out))
		return clock(), result, err
	} else if err := json.NewDecoder(bytes.NewReader(out)).Decode(&result); err != nil {
		log.Println("runMatch Output: ", string(out))
		return clock(), result, err
	} else {
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

func (a DummyArena) RunMatch(p MatchProperties, clock func()time.Time) (time.Time, MatchResult, error) {
	return a.Finish, a.Result, a.Err
}

/*
type MatchResult string

var (
	WinRegex = regexp.MustCompile("\\[java\\] \\[server\\]\\s+([\\w\\d-]+)\\s+\\((A|B)\\) wins \\(round (\\d+)\\)")
	ReasonRegex = regexp.MustCompile("\\[java\\] Reason: ([^\n]+)")
)

type Revision struct {
	Name string `json:"name"`
	GitHash string `json:"git_hash"`
	Date time.Time `json:"date"`
	IsHead bool `json:"date"`
}

const (
	ResultWinA MatchResult = "A"
	ResultWinB MatchResult = "B"
	ResultTieA MatchResult = "TA"
	ResultTieB MatchResult = "TB"
	ResultFail MatchResult = "F"
)

func RunMatch(r1 Revision, r2 Revision, m, script, gitServer, battlecodePath string) MatchResult {
	cmd := exec.Command(script, r1.Name, r1.GitHash, r2.Name, r2.GitHash, m, gitServer, battlecodePath)
	b, err := cmd.CombinedOutput()
	s := string(b)
	if err != nil {
		log.Println(err)
		log.Println(s)
		return ResultFail
	}
	log.Println(s)
	var winA bool
	if m := WinRegex.FindStringSubmatch(s); m != nil {
		winA = m[2] == "A"
	}
	var reason string
	if m := ReasonRegex.FindStringSubmatch(s); m != nil {
		reason = m[1]
	}
	var result MatchResult
	switch reason {
	case "The winning team won by getting a lot of milk.":
		if winA {
			result = ResultWinA
		} else {
			result = ResultWinB
		}
	case "The winning team won on tiebreakers.":
		if winA {
			result = ResultTieA
		} else {
			result = ResultTieB
		}
	default: result = ResultFail
	}
	return result
}
*/
