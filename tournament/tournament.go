package tournament

import (
	"errors"
	"fmt"
	"github.com/GlenKelley/battleref/arena"
	"github.com/GlenKelley/battleref/git"
	"log"
	"strings"
	"time"
)

type Clock interface {
	Now() time.Time
}

//A clock which delegates to the system time
func SystemClock() Clock {
	return &systemClock{}
}

type systemClock struct {
}

func (c *systemClock) Now() time.Time {
	return time.Now()
}

type TournamentCategory string

const (
	CategoryTest           = TournamentCategory("battlecode2014")
	CategoryBattlecode2014 = TournamentCategory("battlecode2014")
	CategoryBattlecode2015 = TournamentCategory("battlecode2015")
)

type Tournament struct {
	Database  Database
	Arena     arena.Arena
	Bootstrap arena.Bootstrap
	GitHost   git.GitHost
	Remote    git.Remote
}

func NewTournament(database Database, arena arena.Arena, bootstrap arena.Bootstrap, gitHost git.GitHost, remote git.Remote) *Tournament {
	return &Tournament{database, arena, bootstrap, gitHost, remote}
}

func (t *Tournament) InstallDefaultMaps(resourcePath string, category TournamentCategory) error {
	if defaultMaps, err := arena.DefaultMaps(resourcePath, string(category)); err != nil {
		return err
	} else if maps, err := t.ListMaps(category); err != nil {
		return err
	} else {
		lookup := make(map[string]bool)
		for _, m := range maps {
			lookup[m] = true
		}
		for name, source := range defaultMaps {
			if !lookup[name] {
				if err := t.CreateMap(name, source, category); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func (t *Tournament) UserExists(name string) (bool, error) {
	exists, err := t.Database.UserExists(name)
	return exists, err
}

func (t *Tournament) ListUsers() ([]string, error) {
	users, err := t.Database.ListUsers()
	return users, err
}

func (t *Tournament) ListCategories() []TournamentCategory {
	return []TournamentCategory{CategoryBattlecode2014, CategoryBattlecode2015}
}

func (t *Tournament) deleteRepository(name string) error {
	if err := t.GitHost.DeleteRepository(name); err != nil {
		fmt.Printf("Failed to delete repo for user %n: %v\n", err)
		return err
	}
	return nil
}

func (t *Tournament) CreatePlayerRepository(name, publicKey string, category TournamentCategory) (string, error) {
	if publicKeys, err := t.Database.ListKeys(); err != nil {
		return "", err
	} else if playerKeys, err := t.Database.PlayerKeys(); err != nil {
		return "", err
	} else if err = t.GitHost.InitRepository(name, playerKeys, publicKeys); err != nil {
		return "", err
	} else if checkout, err := t.GitHost.CloneRepository(t.Remote, name); err != nil {
		defer t.deleteRepository(name)
		return "", err
	} else {
		defer checkout.Delete()
		if files, err := t.Bootstrap.PopulateRepository(name, checkout.Dir(), string(category)); err != nil {
			defer t.deleteRepository(name)
			return "", err
		} else if err := checkout.AddFiles(files); err != nil {
			defer t.deleteRepository(name)
			return "", err
		} else if err := checkout.CommitFiles(files, "Bootstrap_Code"); err != nil {
			defer t.deleteRepository(name)
			return "", err
		} else if err := checkout.Push(); err != nil {
			defer t.deleteRepository(name)
			return "", err
		} else if commitHash, err := checkout.Head(); err != nil {
			defer t.deleteRepository(name)
			return "", err
		} else {
			return commitHash, nil
		}
	}
}

func (t *Tournament) CreateUser(name, publicKey string, category TournamentCategory) (string, error) {
	var commitHash string
	//TODO:(gkelley) this didn't work with a transaction. There is a race condition without one.
	//	return commitHash, t.Database.TransactionBlock(func(tx Statements) error {
	if exists, err := t.Database.UserExists(name); err != nil {
		return "", err
	} else if exists {
		return "", errors.New("User already exists")
	} else if err := t.Database.CreateUser(name, publicKey); err != nil {
		return "", err
	} else if ch, err := t.CreatePlayerRepository(name, publicKey, category); err != nil {
		if err2 := t.Database.DeleteUser(name); err2 != nil {
			fmt.Println(err2)
		}
		return "", err
	} else {
		commitHash = ch
		return commitHash, nil
	}
	//	})
}

func (t *Tournament) CreateMap(name, source string, category TournamentCategory) error {
	return t.Database.CreateMap(name, source, category)
}

func (t *Tournament) GetMapSource(name string, category TournamentCategory) (string, error) {
	source, err := t.Database.GetMapSource(name, category)
	return source, err
}

func (t *Tournament) ListMaps(category TournamentCategory) ([]string, error) {
	users, err := t.Database.ListMaps(category)
	return users, err
}

type Match struct {
	Id       int64
	Player1  string
	Player2  string
	Commit1  string
	Commit2  string
	Map      string
	Category string
	Result   MatchResult
	Time     time.Time
}

func (t *Tournament) ListMatches(category TournamentCategory) ([]Match, error) {
	matches, err := t.Database.ListMatches(category)
	return matches, err
}

func (t *Tournament) MapExists(name string, category TournamentCategory) (bool, error) {
	exists, err := t.Database.MapExists(name, category)
	return exists, err
}

func (t *Tournament) SubmitCommit(name string, category TournamentCategory, commitHash string, time time.Time) error {
	return t.Database.CreateCommit(name, category, commitHash, time)
}

func (t *Tournament) ListCommits(name string, category TournamentCategory) ([]string, error) {
	commits, err := t.Database.ListCommits(name, category)
	return commits, err
}

type Submission struct {
	Name       string
	CommitHash string
}

type MatchResult string

const (
	MatchResultInProgress = "InProgress"
	MatchResultWinA       = "WinA"
	MatchResultWinB       = "WinB"
	MatchResultTieA       = "TieA"
	MatchResultTieB       = "TieB"
	MatchResultError      = "Error"
)

func (t *Tournament) CreateMatch(category TournamentCategory, mapName string, player1, player2 Submission, created time.Time) (int64, error) {
	id, err := t.Database.CreateMatch(category, mapName, player1, player2, created)
	return id, err
}

func (t *Tournament) UpdateMatch(category TournamentCategory, mapName string, player1, player2 Submission, finished time.Time, result MatchResult, replay []byte) error {
	return t.Database.UpdateMatch(category, mapName, player1, player2, finished, result, replay)
}

func (t *Tournament) GetMatchResult(id int64) (MatchResult, error) {
	result, err := t.Database.GetMatchResult(id)
	return result, err
}

func (t *Tournament) GetMatchReplay(id int64) ([]byte, error) {
	replay, err := t.Database.GetMatchReplay(id)
	return replay, err
}

func (t *Tournament) RunMatch(category TournamentCategory, mapName string, player1, player2 Submission, clock Clock) (int64, MatchResult, error) {
	if id, err := t.CreateMatch(category, mapName, player1, player2, clock.Now()); err != nil {
		return 0, MatchResultError, err
	} else {
		if mapSource, err := t.GetMapSource(mapName, category); err != nil {
			return id, MatchResultError, err
		} else if finished, result, err := t.Arena.RunMatch(arena.MatchProperties{
			mapName,
			strings.NewReader(mapSource),
			string(category),
			t.GitHost.RepositoryURL(player1.Name),
			t.GitHost.RepositoryURL(player2.Name),
			player1.CommitHash,
			player2.CommitHash,
		}, func() time.Time { return clock.Now() }); err != nil {
			if err2 := t.UpdateMatch(category, mapName, player1, player2, finished, MatchResultError, []byte{}); err2 != nil {
				log.Println(err2)
			}
			return id, MatchResultError, err
		} else {
			matchResult := GetMatchResult(result)
			if err := t.UpdateMatch(category, mapName, player1, player2, finished, matchResult, result.Replay); err != nil {
				return id, MatchResultError, err
			}
			return id, matchResult, nil
		}
	}
}

func (t *Tournament) LatestCommits(category TournamentCategory) ([]Submission, error) {
	if latestCommits, err := t.Database.LatestCommits(category); err != nil {
		return nil, err
	} else {
		return latestCommits, nil
	}
}

func (t *Tournament) RunLatestMatches(category TournamentCategory) error {
	if latestCommits, err := t.LatestCommits(category); err != nil {
		return err
	} else if maps, err := t.ListMaps(category); err != nil {
		return err
	} else {
		for _, submission1 := range latestCommits {
			for _, submission2 := range latestCommits {
				if submission1.Name != submission2.Name {
					for _, mapName := range maps {
						if _, _, err := t.RunMatch(category, mapName, submission1, submission2, SystemClock()); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func lookupMatch(matchLookup map[string]map[string]map[string]Match, commit1, commit2, mapName string) (Match, bool) {
	if v, ok := matchLookup[commit1]; !ok {
		return Match{}, false
	} else if v2, ok := v[commit2]; !ok {
		return Match{}, false
	} else {
		v3, ok := v2[mapName]
		return v3, ok
	}
}

func lookupInsert(matchLookup map[string]map[string]map[string]Match, match Match) {
	var m1 map[string]map[string]Match
	if v, ok := matchLookup[match.Commit1]; ok {
		m1 = v
	} else {
		m1 = map[string]map[string]Match{}
		matchLookup[match.Commit1] = m1
	}
	var m2 map[string]Match
	if v, ok := m1[match.Commit2]; ok {
		m2 = v
	} else {
		m2 = map[string]Match{}
		m1[match.Commit2] = m2
	}
	m2[match.Map] = match
}

type LeaderboardStats struct {
	Score  float64
	Wins   int
	Ties   int
	Losses int
}

func (l *LeaderboardStats) AddWin() {
	l.Wins++
}

func (l *LeaderboardStats) AddLoss() {
	l.Losses++
}

func (l *LeaderboardStats) AddTie() {
	l.Ties++
}

func (t *Tournament) CalculateLeaderboard(category TournamentCategory) error {
	if latestCommits, err := t.LatestCommits(category); err != nil {
		return err
	} else if matches, err := t.ListMatches(category); err != nil {
		return err
	} else if maps, err := t.ListMaps(category); err != nil {
		return err
	} else {

		matchLookup := map[string]map[string]map[string]Match{}
		stats := map[string]*LeaderboardStats{}
		commits := map[string]string{}
		for _, match := range matches {
			lookupInsert(matchLookup, match)
			stats[match.Player1] = &LeaderboardStats{}
			stats[match.Player2] = &LeaderboardStats{}
			commits[match.Player1] = match.Commit1
			commits[match.Player2] = match.Commit2
		}

		for _, submission1 := range latestCommits {
			for _, submission2 := range latestCommits {
				if submission1.Name != submission2.Name {
					for _, mapName := range maps {
						if match, ok := lookupMatch(matchLookup, submission1.CommitHash, submission2.CommitHash, mapName); ok {
							if result, err := t.GetMatchResult(match.Id); err != nil {
								return err
							} else {
								switch result {
								case MatchResultWinA:
									stats[submission1.Name].AddWin()
									stats[submission2.Name].AddLoss()
									break
								case MatchResultWinB:
									stats[submission1.Name].AddLoss()
									stats[submission2.Name].AddWin()
									break
								case MatchResultTieA:
								case MatchResultTieB:
									stats[submission1.Name].AddTie()
									stats[submission2.Name].AddTie()
									break
								}
							}
						}
					}
				}
			}
		}

		stats2 := map[string]LeaderboardStats{}
		for name, stat := range stats {
			stat.Score = float64(stat.Wins*3 + stat.Losses*-1)
			stats2[name] = *stat
		}

		return t.Database.UpdateLeaderboard(category, stats2, commits)
	}
}

func (t *Tournament) GetLeaderboard(category TournamentCategory) (map[string]LeaderboardStats, []Match, error) {
	ranks, matches, err := t.Database.GetLeaderboard(category)
	return ranks, matches, err
}

func GetMatchResult(a arena.MatchResult) MatchResult {
	if a.Reason == arena.ReasonVictory {
		if a.Winner == arena.WinnerA {
			return MatchResultWinA
		} else {
			return MatchResultWinB
		}
	} else if a.Reason == arena.ReasonTie {
		if a.Winner == arena.WinnerA {
			return MatchResultTieA
		} else {
			return MatchResultTieB
		}
	} else {
		return MatchResultError
	}
}
