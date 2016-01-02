package arena

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Bootstrap interface {
	PopulateRepository(name, repoURL, category string) ([]string, error)
}

type MinimalBootstrap struct {
}

func (b MinimalBootstrap) PopulateRepository(name, sourceDir, category string) ([]string, error) {
	switch category {
	case "battlecode2014":
		{
			files, err := populateBattlecode2014Player(name, sourceDir)
			return files, err
		}
	case "battlecode2015":
		{
			files, err := populateBattlecode2015Player(name, sourceDir)
			return files, err
		}
	default:
		return []string{}, fmt.Errorf("Can't create bootstrap for unkown category %s", category)
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
	if err := ioutil.WriteFile(readmeFile, []byte(Battlecode2014Readme), os.ModePerm); err != nil {
		return []string{}, err
	}
	return files, nil
}

const Battlecode2015Template = `package %s;
import battlecode.common.RobotController;
public class RobotPlayer {
	public static void run(RobotController rc) {
		while (true) {
			rc.yield();
		}
	}
}`

const Battlecode2015Readme = ``

func populateBattlecode2015Player(name, sourceDir string) ([]string, error) {
	sourceFile := filepath.Join(sourceDir, "RobotPlayer.java")
	readmeFile := filepath.Join(sourceDir, "README")
	files := []string{sourceFile, readmeFile}
	if err := ioutil.WriteFile(sourceFile, []byte(fmt.Sprintf(Battlecode2015Template, name)), os.ModePerm); err != nil {
		return []string{}, err
	}
	if err := ioutil.WriteFile(readmeFile, []byte(Battlecode2015Readme), os.ModePerm); err != nil {
		return []string{}, err
	}
	return files, nil
}

func DefaultMaps(resourcePath, category string) (map[string]string, error) {
	mapsPath := filepath.Join(resourcePath, category, "maps")
	if files, err := ioutil.ReadDir(mapsPath); err != nil {
		return nil, err
	} else {
		maps := make(map[string]string, len(files))
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".xml" {
				filename := filepath.Join(mapsPath, file.Name())
				if bs, err := ioutil.ReadFile(filename); err != nil {
					return nil, err
				} else {
					basename := filepath.Base(file.Name())
					mapName := strings.TrimSuffix(basename, filepath.Ext(basename))
					maps[mapName] = string(bs)
				}
			}
		}
		return maps, nil
	}
}
