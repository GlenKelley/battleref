package arena

import (
	"path/filepath"
	"fmt"
	"errors"
	"os"
	"io/ioutil"
)

type Bootstrap interface {
	PopulateRepository(name, repoURL, category string) ([]string, error)
}

type MinimalBootstrap struct {
}

func (b MinimalBootstrap) PopulateRepository(name, sourceDir, category string) ([]string, error) {
	switch(category) {
		case "battlecode2014": {
			files, err := populateBattlecode2014Player(name, sourceDir)
			return files, err
		}
		default: return []string{}, errors.New(fmt.Sprintf("Can't create bootstrap for unkown category %s", category))
	}
}

const Battlecode2014Template = `package %s;
import battlecode.common.RobotController;
public class RobotPlayer {
	public static void run(RobotController rc) {
		while (true) {
			rc.yield();
		}
	}
}`

const Battlecode2014Readme = ``

func populateBattlecode2014Player(name, sourceDir string) ([]string, error) {
	sourceFile := filepath.Join(sourceDir, "RobotPlayer.java")
	readmeFile := filepath.Join(sourceDir, "README")
	files := []string{sourceFile, readmeFile}
	if err := ioutil.WriteFile(sourceFile, []byte(fmt.Sprintf(Battlecode2014Template, name)), os.ModePerm); err != nil {
		return []string{}, err
	}
	if err := ioutil.WriteFile(sourceFile, []byte(Battlecode2014Readme), os.ModePerm); err != nil {
		return []string{}, err
	}
	return files, nil
}

