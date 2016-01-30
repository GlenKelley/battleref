package arena

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Bootstrap interface {
	PopulateRepository(name, repoDir, category string) error
}

type MinimalBootstrap struct {
	ResourceDir string
}

func (b MinimalBootstrap) PopulateRepository(name, repoDir, category string) error {
	switch category {
	case "battlecode2014":
		return populateBattlecode2014Player(name, repoDir)
	case "battlecode2015":
		return populateBattlecode2015Player(name, repoDir)
	case "battlecode2016":
		return populateBattlecode2016Player(name, repoDir, b.ResourceDir)
	default:
		return fmt.Errorf("Can't create bootstrap for unkown category %s", category)
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

func populateBattlecode2014Player(name, sourceDir string) error {
	sourceFile := filepath.Join(sourceDir, "RobotPlayer.java")
	readmeFile := filepath.Join(sourceDir, "README")
	if err := ioutil.WriteFile(sourceFile, []byte(fmt.Sprintf(Battlecode2014Template, name)), os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(readmeFile, []byte(Battlecode2014Readme), os.ModePerm); err != nil {
		return err
	}
	return nil
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

func populateBattlecode2015Player(name, sourceDir string) error {
	sourceFile := filepath.Join(sourceDir, "RobotPlayer.java")
	readmeFile := filepath.Join(sourceDir, "README")
	if err := ioutil.WriteFile(sourceFile, []byte(fmt.Sprintf(Battlecode2015Template, name)), os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(readmeFile, []byte(Battlecode2015Readme), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func populateBattlecode2016Player(name, sourceDir, resourceDir string) error {
	cmd := exec.Command("./copySamplePlayer.sh", "-d", sourceDir, "-n", name)
	cmd.Dir = filepath.Join(resourceDir, "battlecode2016")
	fmt.Println(cmd)
	return cmd.Run()
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
