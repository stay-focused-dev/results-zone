package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/stay-focused-dev/results-zone-reports/internal/exitcode"
	"github.com/stay-focused-dev/results-zone-reports/internal/relay"
)

var opts struct {
	InputFile string `long:"input-file" description:"input file with results"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		parser.WriteHelp(os.Stderr)
		os.Exit(exitcode.UnableToParseArgs)
	}

	if opts.InputFile == "" {
		parser.WriteHelp(os.Stderr)
		os.Exit(exitcode.UndefinedInputFile)
	}

	done := make(chan any)
	defer close(done)

	for v := range relay.Parse(done, opts.InputFile) {
		if v.Err != nil {
			fmt.Fprintf(os.Stderr, "some error occured: %s", v.Err.Error())
		} else {
			fmt.Printf("team: %s, club: %s, bib: %s, name: %s, member: %d, members: %v, time: %s, pace: %s\n",
				v.Team,
				v.Club,
				v.Bib,
				v.TeamName,
				v.Member,
				v.Members,
				v.Time,
				v.Pace,
			)
		}
	}
	os.Exit(exitcode.Ok)
}
