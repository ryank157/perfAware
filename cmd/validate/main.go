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
	var timingEnabled bool
	flag.BoolVar(&timingEnabled, "timing", false, "Enable timing measurements")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprint(os.Stderr, "Usage: --timing <haversine_input.json> <answers.f64> \n")
		os.Exit(1)
	}
	timers := validator.ValidateTimers{}
	inputFileName := flag.Arg(0)
	answersFileName := flag.Arg(1)

	if timingEnabled {
		timers.Begin = uint64(timing.CpuTimer())
	}
	validator.ValidateData(inputFileName, answersFileName, &timers)

	if timingEnabled {

		TotalCPUElapsed := timers.End - timers.Begin

		CPUFreq := timing.EstimateCPUFrequency()
		fmt.Printf("\nTotal time: %0.4fms (CPU freq %d)\n", 1000.0*float64(TotalCPUElapsed)/float64(CPUFreq), CPUFreq)

		timing.PrintTimeElapsed("Startup", TotalCPUElapsed, timers.Begin, timers.Read)
		timing.PrintTimeElapsed("Read", TotalCPUElapsed, timers.Read, timers.MiscSetup)
		timing.PrintTimeElapsed("MiscSetup", TotalCPUElapsed, timers.MiscSetup, timers.Parse)
		timing.PrintTimeElapsed("Parse", TotalCPUElapsed, timers.Parse, timers.Sum)
		timing.PrintTimeElapsed("Sum", TotalCPUElapsed, timers.Sum, timers.MiscOutput)
		timing.PrintTimeElapsed("MiscOutput", TotalCPUElapsed, timers.MiscOutput, timers.End)
	}

}
