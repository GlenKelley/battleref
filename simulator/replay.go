package simulator

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/GlenKelley/battleref/simulator/battlecode2015"
	"github.com/GlenKelley/battleref/simulator/battlecode2016"
	"io"
)

type Replay interface{}

func createReplay(category string) (Replay, error) {
	switch category {
	case "battlecode2015":
		return &bc2015.Replay{}, nil
	case "battlecode2016":
		return &bc2016.Replay{}, nil
	default:
		return nil, fmt.Errorf("Unknown category %v", category)
	}
}

func NewReplay(input io.Reader, category string) (Replay, error) {
	decoder := xml.NewDecoder(input)
	if replay, err := createReplay(category); err != nil {
		return nil, err
	} else if err := decoder.Decode(replay); err != nil {
		fmt.Println("ERROR DECODING", err)
		return nil, err
	} else {
		return replay, nil
	}
}

func NewReplayJson(input io.Reader, category string) (Replay, error) {
	decoder := json.NewDecoder(input)
	if replay, err := createReplay(category); err != nil {
		return nil, err
	} else if err := decoder.Decode(replay); err != nil {
		fmt.Println("ERROR DECODING", err)
		return nil, err
	} else {
		return replay, nil
	}
}
