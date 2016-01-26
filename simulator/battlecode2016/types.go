package bc2016

import (
	"encoding/xml"
)

type Replay struct {
	XMLName         xml.Name        `xml:"object-stream"`
	StoredConstants StoredConstants `xml:"ser.StoredConstants"`
	Header          Header          `xml:"ser.MatchHeader"`
	Metadata        Info            `xml:"ser.ExtensibleMetadata"`
	Round           []Round         `xml:"ser.RoundDelta"`
	GameStats       GameStats       `xml:"ser.GameStats"`
	Footer          Footer          `xml:"ser.MatchFooter"`
}

type Header struct {
	MatchNumber int         `xml:"matchNumber,attr"`
	MatchCount  int         `xml:"matchCount,attr"`
	Map         Map         `xml:"map"`
	State       LongArray2D `xml:"state>long-array"`
}

type Info struct {
	Type  string      `xml:"type,attr"`
	TeamA Team        `xml:"team-a,attr"`
	TeamB Team        `xml:"team-b,attr"`
	Maps  StringArray `xml:"maps,attr"`
}

type GameStats struct {
	DominationFactor DominationFactor `xml:"dominationFactor,attr",`
}

type Footer struct {
	Winner Team        `xml:"winner,attr"`
	State  LongArray2D `xml:"state>long-array"`
}

type Map struct {
	Width               int             `xml:"width,attr"`
	Height              int             `xml:"height,attr"`
	Origin              MapLoc          `xml:"origin,attr"`
	Seed                int             `xml:"seed,attr"`
	Rounds              int             `xml:"rounds,attr"`
	Name                string          `xml:"mapName,attr"`
	Armageddon          bool            `xml:"armageddon,attr"`
	InitialRubble       []FloatArray2D  `xml:"initialRubble>double-array"`
	InitialParts        []FloatArray2D  `xml:"initialParts>double-array"`
	ZombieSpawnSchedule []SpawnSchedule `xml:"zombieSpawnSchedule>round"`
	InitialRobots       []InitialRobot  `xml:"initialRobots>initial-robot"`
}

type SpawnSchedule struct {
	Number      int           `xml:"number,attr"`
	ZombieCount []ZombieCount `xml:"zombie-count"`
}

type ZombieCount struct {
	Type  string `xml:"type,attr"`
	Count int    `xml:"count,attr"`
}

type InitialRobot struct {
	OriginOffsetX int    `xml:"originOffsetX,attr"`
	OriginOffsetY int    `xml:"originOffsetY,attr"`
	Type          string `xml:"type,attr"`
	Team          Team   `xml:"team,attr"`
}

type Round struct {
	Signals []Signal `xml:",any"`
}

type Signal struct {
	XMLName xml.Name

	Location  *MapLoc    `xml:"location,attr" json:",omitempty"`
	RobotId   *RobotId   `xml:"robotID,attr" json:",omitempty"`
	TargetLoc *MapLoc    `xml:"targetLoc,attr" json:",omitempty"`
	RobotTeam *Team      `xml:"robotTeam,attr" json:",omitempty"`
	ParentId  *RobotId   `xml:"parentID,attr" json:",omitempty"`
	Loc       *MapLoc    `xml:"loc,attr" json:",omitempty"`
	Type      *RobotType `xml:"type,attr" json:",omitempty"`
	Team      *Team      `xml:"team,attr" json:",omitempty"`

	NewLoc       *MapLoc       `xml:"newLoc,attr" json:",omitempty"`
	Delay        *int          `xml:"delay,attr" json:",omitempty"`
	RobotIds     *RobotIdArray `xml:"robotIDs,attr" json:",omitempty"`
	NumByteCodes *IntArray     `xml:"numBytecodes,attr" json:",omitempty"`
	ObjectId     *RobotId      `xml:"objectID,attr" json:",omitempty"`
	Health       *string       `xml:"health,attr" json:",omitempty"`
	StringIndex  *int          `xml:"stringIndex,attr" json:",omitempty"`
	NewString    *string       `xml:"newString,attr" json:",omitempty"`

	CoreDelays          *FloatArray `xml:"coreDelays,attr" json:",omitempty"`
	WeaponDelays        *FloatArray `xml:"weaponDelays,attr" json:",omitempty"`
	Amount              *float64    `xml:"amount,attr" json:",omitempty"`
	DeathByActivation   *string     `xml:"deathByActivation,attr" json:",omitempty"`
	Message             *string     `xml:"message,attr" json:",omitempty"`
	Radius              *string     `xml:"radius,attr" json:",omitempty"`
	Resource            *string     `xml:"resource,attr" json:",omitempty"`
	ZombieInfectedRurns *string     `xml:"zombieInfectedTurns,attr" json:",omitempty"`
	ViperInfectedRurns  *string     `xml:"viperInfectedTurns,attr" json:",omitempty"`

	Component *Signal `xml:"signal,omitempty"`
	Cause     string  `xml:"cause,omitempty"`
}

type StoredConstants struct {
	EngineVersion string  `xml:"engineVersion,attr"`
	GameConstants []Entry `xml:"gameConstants>entry"`
	RobotTypes    []Entry `xml:"robotTypes>entry"`
}

type Entry struct {
	String string `xml:"string"`
	Enum   Value  `xml:",any"`
}

type Value struct {
	XMLName xml.Name
	Data    string `xml:",innerxml"`
}

const (
	Win0 = DominationFactor("DESTROYED")
	Win1 = DominationFactor("PWNED")
	Win2 = DominationFactor("BEAT")
	Win3 = DominationFactor("BARELY_BEAT")
	Win4 = DominationFactor("BARELY_BARELY_BEAT")
	Win5 = DominationFactor("WON_BY_DUBIOUS_REASONS")
)
