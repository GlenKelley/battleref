package simulator

type SpawnSignal struct {
	Id     RobotId
	Parent RobotId
	Loc    MapLoc
	Type   RobotType
	Team   Team
	Delay  int
}

type AttackSignal struct {
	Id        RobotId
	TargetLoc MapLoc
}

type BashSignal struct {
	Id        RobotId
	TargetLoc MapLoc
}

type Broadcast struct {
	Id      RobotId
	Team    Team
	Message map[int]int
}

type Build struct {
	Id        RobotId
	Parent    RobotId
	TargetLoc MapLoc
	Type      RobotType
	Team      Team
	Delay     int
	Dir       Direction
}

type BytecodesUsedSignal struct {
	BytecodesUsed map[RobotId]int
}

type CastSignal struct {
	Id        RobotId
	TargetLoc MapLoc
}

type ControlBitsSignal struct {
	Id          RobotId
	ControlBits int64
}

type DeathSignal struct {
	Id RobotId
}

type HeathChangeSignal struct {
	HealthChange map[RobotId]float64
}

type IndicatorDotSignal struct {
	Id    RobotId
	Team  Team
	Loc   MapLoc
	Color Color
}

type IndicatorLineSignal struct {
	Id    RobotId
	Team  Team
	LocA  MapLoc
	LocB  MapLoc
	Color Color
}

type IndicatorStringSignal struct {
	Id    RobotId
	Index int
	Value string
}

type LocationOreChanged struct {
	Loc MapLoc
	Ore float64
}

type MatchObservationSignal struct {
	Id          RobotId
	Observation string
}

type MineSignal struct {
	Id   RobotId
	Loc  MapLoc
	Team Team
	Type RobotType
}

type MissileCountSignal struct {
	Id           RobotId
	MissingCount int
}

type MovementOverrideSignal struct {
	Id  RobotId
	Loc MapLoc
}

type MovementSignal struct {
	Id              RobotId
	Loc             MapLoc
	IsMovingForward bool
	Delay           int
}

type RobotInfoSignal struct {
	Infos map[RobotId]RobotInfo
}

type RobotInfo struct {
	CoreDelay   float64
	WeaponDelay float64
	SupplyLevel float64
}

type SelfDestructSignal struct {
	Id           RobotId
	Loc          MapLoc
	DamageFactor float64
}

type TeamOreSignal struct {
	Ore []float64
}

type TransferSupplySignal struct {
	From   RobotId
	To     RobotId
	Amount float64
}

type XpSignal struct {
	Id RobotId
	Xp int
}
