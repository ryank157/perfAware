package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ryank157/perfAware/internal/timing"
	"github.com/ryank157/perfAware/internal/validator"
)

func main() {

	// Parse the flags, and note must must be called before finding flags.
	// var timingEnabled bool
	// flag.BoolVar(&timingEnabled, "timing", false, "Enable timing measurements")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprint(os.Stderr, "Usage: <haversine_input.json> <answers.f64> \n")
		os.Exit(1)
	}
	inputFileName := flag.Arg(0)
	answersFileName := flag.Arg(1)

	timing.BeginProfile()

	validator.ValidateData(inputFileName, answersFileName)

	timing.EndAndPrintProfile()

}
