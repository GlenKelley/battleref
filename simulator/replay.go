package simulator

import (
	"encoding/xml"
	"io"
)

func NewReplay(input io.Reader) (*Replay, error) {
	decoder := xml.NewDecoder(input)
	replay := Replay{}
	err := decoder.Decode(&replay)
	return &replay, err
}

/*
func NewReplay(input io.Reader) (*Replay, error) {
	decoder := xml.NewDecoder(input)
	var replay Replay
	for token, err := decoder.Token(); token != nil && err != io.EOF; token, err = decoder.Token() {
		if err != nil {
			return nil, err
		}
		switch t := token.(type) {
			case xml.StartElement:
				fmt.Println("Start element:", t.Name.Local)
				switch t.Name.Local {
					case "ser.MatchHeader":
						if err := decodeHeader(decoder, &replay.Header); err != nil {
							return nil, err
						}
						break
					case "ser.ExtensibleMetadata":
						if err := decodeMetadata(decoder, &replay.Info, t); err != nil {
							return nil, err
						}
						break
					case "ser.RoundDelta":
						var round Round
						if err := decodeRoundDelta(decoder, &round); err != nil {
							return nil, err
						}
						if _, err := decoder.Token(); err != nil
							return nil, err
						} else {
							if err := decodeRoundStats(decoder, &round); err != nil {
							return nil, err
						}
						break
				default:
				}
				break
			case xml.EndElement: fmt.Println("End element:", t.Name); break
			case xml.Comment: fmt.Println("Comment:", string(m)); break
			default:
		}
	}
	return &replay, nil
}

func decodeHeader(decoder *xml.Decoder, header *Header) error {
	fmt.Println("Parsing Header")
	for token, err := decoder.Token();; {
		if err != nil {
			return err
		}
		switch t := token.(type) {
			case xml.StartElement:
				fmt.Println("Start element:", t.Name.Local)
				switch t.Name.Local {
					case "map": decoder.Skip(); break
					case "state": decoder.Skip(); break
				default:
				}
				break;
			case xml.EndElement: return nil
			case xml.Comment: fmt.Println("Comment:", string(m)); break
			default:
		}
		token, err = decoder.Token()
	}
}

func decodeMetadata(decoder *xml.Decoder, info *Info, start *xml.StartElement) error {
	for k, v := range start.Attr {
		switch k {
		case "team-a": info.TeamA = v; break
		case "team-b": info.TeamB = v; break
		case "maps": info.maps = v; break
	}
	return decoder.Skip()
}

func decodeRoundDelta(decoder *xml.Decoder, round *Round) error {
	fmt.Println("Parsing Round Delta")
	for token, err := decoder.Token();; {
		if err != nil {
			return err
		}
		switch t := token.(type) {
			case xml.StartElement:
				fmt.Println("Start element:", t.Name.Local)
				m := attrMap(t.Attrs)
				var signal Signal
				var err error
				switch t.Name.Local {
					case "sig.AttackSignal": signal, err = decodeAttackSignal(m); decoder.Skip(); break
					case "sig.BashSignal": signal, err = decodeBashSignal(m); decoder.Skip(); break
					case "sig.BroadcastSignal": signal, err = decodeBroadcastSignal(m); decoder.Skip(); break
					case "sig.BuildSignal": signal, err = decodeBuildSignal(m); decoder.Skip(); break
					case "sig.BytecodeUsedSignal": signal, err = decodeBytecodeUsedSignal(m); decoder.Skip(); break
					case "sig.CastSignal": signal, err = decodeCastSignal(m); decoder.Skip(); break
					case "sig.ControlBitsSignal": signal, err = decodeControlBitsSignal(m); decoder.Skip(); break
					case "sig.DeathSignal": signal, err = decodeDeathSignal(m); decoder.Skip(); break
					case "sig.HealthChangeSignal": signal, err = decodeHealthChangeSignal(m); decoder.Skip(); break
					case "sig.IndicatorDotSignal": signal, err = decodeIndicatorDotSignal(m); decoder.Skip(); break
					case "sig.IndicatorLineSignal": signal, err = decodeIndicatorLineSignal(m); decoder.Skip(); break
					case "sig.IndicatorStringSignal": signal, err = decodeIndicatorStringSignal(m); decoder.Skip(); break
					case "sig.LocatorOreChangeSignal": signal, err = decodeLocatorOreChangeSignal(m); decoder.Skip(); break
					case "sig.MatchObservationSignal": signal, err = decodeMatchObservationSignal(m); decoder.Skip(); break
					case "sig.MineSignal": signal, err = decodeMineSignal(m); decoder.Skip(); break
					case "sig.MissileCountSignal": signal, err = decodeMissileCountSignal(m); decoder.Skip(); break
					case "sig.MovementOverrideSignal": signal, err = decodeMovementOverrideSignal(m); decoder.Skip(); break
					case "sig.MovementSignal": signal, err = decodeMovementSignal(m); decoder.Skip(); break
					case "sig.RobotInfoSignal": signal, err = decodeRobotInfoSignal(m); decoder.Skip(); break
					case "sig.SelfDestructSignal": signal, err = decodeSelfDestructSignal(m); decoder.Skip(); break
					case "sig.SpawnSignal": signal, err = decodeSpawnSignal(m); decoder.Skip(); break
					case "sig.TeamOreSignal": signal, err = decodeTeamOreSignal(m); decoder.Skip(); break
					case "sig.TransferSupplySignal": signal, err = decodeTransferSupplySignal(m); decoder.Skip(); break
					case "sig.XpSignal": signal, err = decodeXpSignal(m); decoder.Skip(); break
					default:
				}
				break;
			case xml.EndElement: return nil
			case xml.Comment: fmt.Println("Comment:", string(m)); break
			default:
		}
		token, err = decoder.Token()
	}
}

func intArray(ss string) ([]int, error) {
	is := []int{}
	for s := range strings.split(ss, ",") {
		if n, err := strconv.Atoi(s); err != nil {
			return is, err
		} else {
			is = append(is, n)
		}
	}
	return is, nil
}

func decodeLoc(s string) (MapLoc, error) {
	if is, err := intArray(s); err != nil {
		return MapLoc{}, err
	} else {
		return MapLoc{is[0], is[1]}, nil
	}
}

func decodeAttackSignal(attrs map[string]string) (Signal, error) {
	signal := Signal{}
	if id, err := attrs["RobotId"]; err != nil {
		return signal, err
	} else if loc, err := decodeLoc(atrs["targetLoc"]) {
		return signal, err
	} else {
		signal.AttackSignal = AttackSignal{id, loc}
		return signal, nil
	}
}

func decodeBashSignal(attrs map[string]string) (Signal, error) {
	return Signal{}, nil
}

func decodeBroadcastSignal(attrs map[string]string) (Signal, error) {
	return Signal{}, nil
}

func decodeBuildSignal(attrs map[string]string) (Signal, error) {
	return Signal{}, nil
}

func decodeBytecodeUsedSignal(attrs map[string]string) (Signal, error) {
	return Signal{}, nil
}

func decodeCastSignal(attrs map[string]string) (Signal, error) {
	return Signal{}, nil
}

func decodeControlBitsSignal(attrs map[string]string) (Signal, error) {
	return Signal{}, nil
}

func decodeDeathSignal(attrs map[string]string) (Signal, error) {
	return Signal{}, nil
}

func decodeHealthChangeSignal(attrs map[string]string) (Signal, error) {
	return Signal{}, nil
}

func decodeIndicatorDotSignal(attrs map[string]string) (Signal, error) {
	return Signal{}, nil
}

func decodeIndicatorLineSignal(attrs map[string]string) (Signal, error) {
	return Signal{}, nil
}

func decodeIndicatorStringSignal(attrs map[string]string) (Signal, error) {
func decodeLocatorOreChangeSignal(attrs map[string]string) (Signal, error) {
func decodeMatchObservationSignal(attrs map[string]string) (Signal, error) {
func decodeMineSignal(attrs map[string]string) (Signal, error) {
func decodeMissileCountSignal(attrs map[string]string) (Signal, error) {
func decodeMovementOverrideSignal(attrs map[string]string) (Signal, error) {
func decodeMovementSignal(attrs map[string]string) (Signal, error) {
func decodeRobotInfoSignal(attrs map[string]string) (Signal, error) {
func decodeSelfDestructSignal(attrs map[string]string) (Signal, error) {
func decodeSpawnSignal(attrs map[string]string) (Signal, error) {
func decodeTeamOreSignal(attrs map[string]string) (Signal, error) {
func decodeTransferSupplySignal(attrs map[string]string) (Signal, error) {
func decodeXpSignal(attrs map[string]string) (Signal, error) { {
}

func decodeRoundStats(decoder *xml.Decoder, round *Round, start *xml.StartElement) error {
	for k, v := range start.Attr {
		switch k {
		case "points":
			for s := range strings.split(v, ",") {
				if n, err := strconv.Atoi(s); err != nil {
					return err
				} else {
					round.Stats = append(round.Stats, n)
				}
			}
			break
		}
	}
	return decoder.Skip()
}

*/
