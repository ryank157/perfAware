package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/ryank157/perfAware/internal/generator"
	"github.com/ryank157/perfAware/internal/timing"
)

type GenerateTimers struct {
	Begin      uint64
	Read       uint64
	MiscSetup  uint64
	Generate   uint64
	Sum        uint64
	MiscOutput uint64
	End        uint64
}

func main() {
	// var timingEnabled bool
	// flag.BoolVar(&timingEnabled, "timing", false, "Enable timing measurements")
	flag.Parse()
	spread := flag.Arg(0)
	if spread != "uniform" && spread != "cluster" {
		log.Fatal("Invalid spread type.  Must be 'uniform' or 'cluster'.")
	}

	seed, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		log.Fatalf("Invalid seed.  Must be an integer: %v", err)
	}

	numPoints, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		log.Fatalf("Invalid numPoints. Must be an integer: %v", err)
	}

	// Generate data + answer file as bin
	timing.BeginProfile()
	avgDistance := generator.GenerateDataSetAndAnswerFiles(spread, seed, numPoints)

	fmt.Printf("Method: %s\n", spread)
	fmt.Printf("Random seed: %d\n", seed)
	fmt.Printf("Pair count: %d\n", numPoints)
	fmt.Printf("Average distance: %f\n", avgDistance)
	timing.EndAndPrintProfile()
}
