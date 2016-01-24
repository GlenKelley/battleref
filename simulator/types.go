package simulator

import (
	"encoding/xml"
)

type Team string

const (
	TeamA = Team("A")
	TeamB = Team("B")
)

type DominationFactor string

const (
	Win0 = "DESTROYED"
	Win1 = "PWNED"
	Win2 = "BEAT"
	Win3 = "BARELY_BEAT"
	Win4 = "BARELY_BARELY_BEAT"
	Win5 = "WON_BY_DUBIOUS_REASONS"
)

type Direction string

const (
	DirectionNorth     = Direction("NORTH")
	DirectionNorthEast = Direction("NORTH_EAST")
	DirectionEast      = Direction("EAST")
	DirectionSouthEast = Direction("SOUTH_EAST")
	DirectionSouth     = Direction("SOUTH")
	DirectionSouthWest = Direction("SOUTH_WEST")
	DirectionWest      = Direction("WEST")
	DirectionNorthWest = Direction("NORTH_WEST")
	DirectionNone      = Direction("NONE")
	DirectionOmni      = Direction("OMNI")
)

type RobotId int

type Color [3]int

type RobotType string

const (
	RobotHQ                  = RobotType("HQ")
	RobotTower               = RobotType("TOWER")
	RobotSupplyDepot         = RobotType("SUPPLYDEPOT")
	RobotTechnologyInstitute = RobotType("TECHNOLOGYINSTITUTE")
	RobotBarracks            = RobotType("BARRACKS")
	RobotHelipad             = RobotType("HELIPAD")
	RobotTainingField        = RobotType("TRAININGFIELD")
	RobotTankFactory         = RobotType("TANKFACTORY")
	RobotMinerFactory        = RobotType("MINERFACTORY")
	RobotHandwashStation     = RobotType("HANDWASHSTATION")
	RobotAerospaceLab        = RobotType("AEROSPACELAB")
	RobotBeaver              = RobotType("BEAVER")
	RobotComputer            = RobotType("COMPUTER")
	RobotSoldier             = RobotType("SOLDIER")
	RobotBasher              = RobotType("BASHER")
	RobotMiner               = RobotType("MINER")
	RobotDrone               = RobotType("DRONE")
	RobotTank                = RobotType("TANK")
	RobotCommander           = RobotType("COMMANDER")
	RobotLauncher            = RobotType("LAUNCHER")
	RobotMissile             = RobotType("MISSILE")
)

type Replay struct {
	XMLName   xml.Name  `xml:"object-stream"`
	Header    Header    `xml:"ser.MatchHeader"`
	Metadata  Info      `xml:"ser.ExtensibleMetadata"`
	Round     []Round   `xml:",any"`
	GameStats GameStats `xml:"ser.GameStats"`
	Footer    Footer    `xml:"ser.MatchFooter"`
}

type Header struct {
	MatchNumber int         `xml:"matchNumber,attr"`
	MatchCount  int         `xml:"matchCount,attr"`
	Map         Map         `xml:"map"`
	State       []LongArray `xml:"state>long-array"`
}

type Info struct {
	Type  string      `xml:"type,attr"`
	TeamA string      `xml:"team-a,attr"`
	TeamB string      `xml:"team-b,attr"`
	Maps  StringArray `xml:"maps,attr"`
}

type GameStats struct {
	TimeToFirstKill       string           `xml:"timeToFirstKill,attr"`
	TimeToFirstArchonKill string           `xml:"timeToFirstArchonKill,attr"`
	TotalPoints           string           `xml:"totalPoints,attr"`
	NumArchons            string           `xml:"numArchons,attr"`
	TotalEnergon          string           `xml:"totalEnergon,attr"`
	DominationFactor      DominationFactor `xml:"dominationFactor,attr",`
	ExcitementFactor      float64          `xml:"excitementFactor,attr"`
	TimeToTallestTower    int              `xml:"timeToTallestTower,attr"`
	TallestTower          int              `xml:"tallestTower,attr"`
}

type Footer struct {
	Winner Team        `xml:"winner,attr"`
	State  []LongArray `xml:"state>long-array"`
}

type Map struct {
	Class         string     `xml:"class,attr"`
	Width         int        `xml:"mapWidth,attr"`
	Height        int        `xml:"mapHeight,attr"`
	OriginX       int        `xml:"mapOriginX,attr"`
	OriginY       int        `xml:"mapOriginY,attr"`
	Seed          int        `xml:"seed,attr"`
	MaxRounds     int        `xml:"maxRounds,attr"`
	Name          string     `xml:"mapName,attr"`
	MaxInitialOre int        `xml:"maxInitialOre,attr"`
	MapTiles      string     `xml:"mapTiles"`
	InitialOre    []IntArray `xml:"mapInitialOre>int-array"`
}

type MapTiles struct {
	TileData string `xml:",innerxml"`
}

type Round struct {
	XMLName xml.Name
	Signals []Signal   `xml:",any,omitempty"`
	Points  FloatArray `xml:"points,attr,omitempty"`
}

type Signal struct {
	XMLName   xml.Name
	MineLoc   *MapLoc    `xml:"mineLoc,attr" json:",omitempty"`
	RobotId   *RobotId   `xml:"robotID,attr" json:",omitempty"`
	TargetLoc *MapLoc    `xml:"targetLoc,attr" json:",omitempty"`
	RobotTeam *Team      `xml:"robotTeam,attr" json:",omitempty"`
	ParentId  *RobotId   `xml:"parentID,attr" json:",omitempty"`
	Loc       *MapLoc    `xml:"loc,attr" json:",omitempty"`
	Type      *RobotType `xml:"type,attr" json:",omitempty"`
	Team      *Team      `xml:"team,attr" json:",omitempty"`

	NewLoc          *MapLoc       `xml:"newLoc,attr" json:",omitempty"`
	IsMovingForward *bool         `xml:"isMovingForward,attr" json:",omitempty"`
	Delay           *int          `xml:"delay,attr" json:",omitempty"`
	RobotIds        *RobotIdArray `xml:"robotIDs,attr" json:",omitempty"`
	NumByteCodes    *IntArray     `xml:"numBytecodes,attr" json:",omitempty"`
	ControlBits     *int64        `xml:"controlBits,attr" json:",omitempty"`
	ObjectId        *RobotId      `xml:"objectID,attr" json:",omitempty"`
	Health          *string       `xml:"health,attr" json:",omitempty"`
	Location        *MapLoc       `xml:"location,attr" json:",omitempty"`
	Red             *int          `xml:"red,attr" json:",omitempty"`
	Green           *int          `xml:"green,attr" json:",omitempty"`
	Blue            *int          `xml:"blue,attr" json:",omitempty"`
	Loc1            *MapLoc       `xml:"loc1,attr" json:",omitempty"`
	Loc2            *MapLoc       `xml:"loc2,attr" json:",omitempty"`
	StringIndex     *int          `xml:"stringIndex,attr" json:",omitempty"`
	NewString       *string       `xml:"newString,attr" json:",omitempty"`
	Ore             *FloatArray   `xml:"ore,attr" json:",omitempty"`
	Observation     *string       `xml:"observation,attr" json:",omitempty"`
	MineTeam        *Team         `xml:"mineTeam,attr" json:",omitempty"`
	MinerType       *RobotType    `xml:"minerType,attr" json:",omitempty"`
	MissileCount    *int          `xml:"missileCount,attr" json:",omitempty"`

	CoreDelays   *FloatArray `xml:"coreDelays,attr" json:",omitempty"`
	WeaponDelays *FloatArray `xml:"weaponDelays,attr" json:",omitempty"`
	SupplyLevels *FloatArray `xml:"supplyLevels,attr" json:",omitempty"`
	FromId       *RobotId    `xml:"fromID,attr" json:",omitempty"`
	ToId         *RobotId    `xml:"toID,attr" json:",omitempty"`
	Amount       *float64    `xml:"amount,attr" json:",omitempty"`
	Xp           *int        `xml:"XP,attr" json:",omitempty"`
	DamageFactor *float64    `xml:"damageFactor,attr" json:",omitempty"`
}

type FloatArray string
type IntArray string
type LongArray string
type MapLog string
type RobotIdArray string
type StringArray string
type MapLoc string
