package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/stay-focused-dev/results-zone-reports/internal/exitcode"
	"github.com/stay-focused-dev/results-zone-reports/internal/relay"
	"github.com/xuri/excelize/v2"
)

var opts struct {
	InputFile  string `long:"input-file" description:"input file with results"`
	OutputFile string `long:"output-file" description:"output file with results"`
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

	if opts.OutputFile == "" {
		parser.WriteHelp(os.Stderr)
		os.Exit(exitcode.UndefinedOutputFile)
	}

	done := make(chan any)
	defer close(done)

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheet := "Sheet1"
	f.SetCellValue(sheet, "A1", "Название команды")
	f.SetCellValue(sheet, "B1", "Участник")
	f.SetCellValue(sheet, "C1", "Номер этапа")
	f.SetCellValue(sheet, "D1", "Время (без транзитки)")
	f.SetCellValue(sheet, "E1", "Темп (без транзитки)")
	f.SetCellValue(sheet, "F1", "Участники команды")

	i := 1
	for v := range relay.Parse(done, opts.InputFile) {
		if v.Err != nil {
			fmt.Fprintf(os.Stderr, "some error occured: %s", v.Err.Error())
		} else {
			i++

			f.SetCellValue(sheet, fmt.Sprintf("A%d", i), v.TeamName)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", i), "")
			f.SetCellValue(sheet, fmt.Sprintf("C%d", i), v.Member)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", i), v.Time)
			f.SetCellValue(sheet, fmt.Sprintf("E%d", i), v.Pace)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", i), v.Members)
		}
	}

	if err := f.SaveAs(opts.OutputFile); err != nil {
		fmt.Println(err)
	}

	os.Exit(exitcode.Ok)
}
