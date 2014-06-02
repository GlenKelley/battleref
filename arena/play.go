package arena

import (
	"os/exec"
	"time"
	"log"
	"regexp"
)

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

func PlayMatch(r1 Revision, r2 Revision, m, script, gitServer, battlecodePath string) MatchResult {
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