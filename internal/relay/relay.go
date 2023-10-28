package relay

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/stay-focused-dev/results-zone/internal/result"
)

type items struct {
	Items []item `json:"items"`
}

type item struct {
	Bib       string    `json:"bib"`
	Team      string    `json:"team"`
	Club      string    `json:"club"`
	Splits    []split   `json:"splits"`
	RelayTeam relayTeam `json:"relay_team"`
}

type split struct {
	Name     string          `json:"name"`
	Gun      decimal.Decimal `json:"gun"`
	Start    decimal.Decimal `json:"start"`
	Lap      uint32          `json:"lap"`
	Distance decimal.Decimal `json:"distance"`
}

type relayTeam struct {
	Name    string   `json:"name"`
	Gender  string   `json:"gender"`
	Members []string `json:"members"`
}

func Parse(done <-chan any, filename string) <-chan result.Result {
	ret := make(chan result.Result)
	go func() {
		defer close(ret)

		jsonFile, err := os.Open(filename)
		if err != nil {
			ret <- result.Result{Err: err}
			return
		}
		defer jsonFile.Close()

		bytes, err := io.ReadAll(jsonFile)
		if err != nil {
			ret <- result.Result{Err: err}
			return
		}

		var items items

		err = json.Unmarshal(bytes, &items)
		if err != nil {
			ret <- result.Result{Err: err}
			return
		}

		type splitDiff struct {
			s1 string
			s2 string
		}
		splits := []splitDiff{
			{s1: "T1", s2: "Этап 1"},
			{s1: "Т2", s2: "Этап 2"},
			{s1: "Т3", s2: "Этап 3"},
			{s1: "Т4", s2: "Этап 4"},
			{s1: "Т5", s2: "Этап 5"},
		}
		for _, v := range items.Items {
			for i, s := range splits {
				t, d, err := findTimeBetweenSplits(v.Splits, s.s1, s.s2)
				if err != nil {
					select {
					case ret <- result.Result{Err: err}:
						return
					}
				}

				select {
				case <-done:
					return
				case ret <- result.Result{
					Bib:      v.Bib,
					Team:     v.Team,
					TeamName: v.RelayTeam.Name,
					Club:     v.Club,
					Time:     secToHHMMSS(t),
					Pace:     secondsMetersToPaceMMSS(t, d),
					Members:  strings.Join(v.RelayTeam.Members, " / "),
					Member:   i + 1,
				}:
				}
			}
		}
	}()
	return ret
}

func findSplitByName(splits []split, splitName string) (split, error) {
	for _, v := range splits {
		if v.Name == splitName {
			return v, nil
		}
	}
	return split{}, fmt.Errorf("Unable to find split by name: %s", splitName)
}

func findTimeBetweenSplits(splits []split, from string, to string) (decimal.Decimal, decimal.Decimal, error) {
	s1, err := findSplitByName(splits, from)
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, err
	}

	s2, err := findSplitByName(splits, to)
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, err
	}

	return s2.Start.Sub(s1.Start), s2.Distance.Sub(s1.Distance), nil
}

func secToHHMMSS(s decimal.Decimal) string {
	sf, _ := s.Float64()

	hour := int(math.Floor(sf / 3600))
	sf = sf - float64(hour*3600)

	min := int(math.Floor(sf / 60))
	sf = sf - float64(min*60)

	sec := int(math.Floor(sf))
	sf = sf - float64(sec)

	ms := int(sf * 1000)

	return fmt.Sprintf("%d:%02d:%02d.%03d", hour, min, sec, ms)
}

func secondsMetersToPaceMMSS(s decimal.Decimal, m decimal.Decimal) string {
	sf, _ := s.Float64()
	mf, _ := m.Float64()

	pace := sf / (mf / 1000) // sec / km

	min := int(math.Floor(pace / 60))
	sec := int(pace - float64(min*60))

	return fmt.Sprintf("%02d:%02d", min, sec)
}
