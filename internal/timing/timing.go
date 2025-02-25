package timing

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

const osTimerFreq = 1_000_000_000

func OsTimer() int64 {
	return time.Now().UnixNano()
}

func EstimateCPUFrequency() uint64 {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	const milliToWait = 100
	osFreq := uint64(osTimerFreq)

	cpuStart := CpuTimer()
	osStart := OsTimer()
	osEnd := int64(0)
	osElapsed := int64(0)
	osWaitTime := osFreq * milliToWait / 1000

	for osElapsed < int64(osWaitTime) {
		osEnd = OsTimer()
		osElapsed = osEnd - osStart
	}

	cpuEnd := CpuTimer()
	cpuElapsed := cpuEnd - cpuStart

	cpuFreq := uint64(0)
	if osElapsed > 0 {
		cpuFreq = osFreq * uint64(cpuElapsed) / uint64(osElapsed)
	}

	return cpuFreq
}

func PrintTimeElapsed(label string, totalTSCElapsed uint64, begin uint64, end uint64) {
	elapsed := end - begin
	percent := 100.0 * (float64(elapsed) / float64(totalTSCElapsed))
	fmt.Printf("   %-15s: %d (%.2f%%)\n", label, elapsed, percent)
}

// ProfileAnchor stores timing information for a code block.
type ProfileAnchor struct {
	TSCElapsed uint64
	HitCount   uint64
	Label      string
}

// Profiler manages the profiling data.
type Profiler struct {
	Anchors   [4096]ProfileAnchor
	AnchorMap map[string]int // Map label to anchor index
	StartTSC  uint64
	EndTSC    uint64
	Mutex     sync.Mutex
	Counter   int // Counter for next available anchor
}

// GlobalProfiler is the global instance of the profiler.
var GlobalProfiler = Profiler{
	AnchorMap: make(map[string]int),
}

// BeginProfile starts the profiling session.
func BeginProfile() {
	GlobalProfiler.Mutex.Lock()
	defer GlobalProfiler.Mutex.Unlock()
	GlobalProfiler.StartTSC = CpuTimer()
}

// EndAndPrintProfile ends the profiling session and prints the results.
func EndAndPrintProfile() {
	GlobalProfiler.Mutex.Lock()
	defer GlobalProfiler.Mutex.Unlock()
	GlobalProfiler.EndTSC = CpuTimer()

	cpuFreq := EstimateCPUFrequency()
	totalCPUElapsed := GlobalProfiler.EndTSC - GlobalProfiler.StartTSC

	if cpuFreq > 0 {
		fmt.Printf("\nTotal time: %.4fms (CPU freq %d)\n", 1000.0*float64(totalCPUElapsed)/float64(cpuFreq), cpuFreq)
	}

	for i := 0; i < len(GlobalProfiler.Anchors); i++ {
		anchor := &GlobalProfiler.Anchors[i]
		if anchor.TSCElapsed > 0 {
			printTimeElapsed(totalCPUElapsed, anchor)
		}
	}
}

func printTimeElapsed(totalTSCElapsed uint64, anchor *ProfileAnchor) {
	elapsed := anchor.TSCElapsed
	percent := 100.0 * (float64(elapsed) / float64(totalTSCElapsed))
	fmt.Printf("  %s[%d]: %d (%.2f%%)\n", anchor.Label, anchor.HitCount, elapsed, percent)
}

// TimeBlock is a function that returns a function to stop the timer
func TimeBlock(label string) func() {

	GlobalProfiler.Mutex.Lock()
	defer GlobalProfiler.Mutex.Unlock()

	anchorIndex, ok := GlobalProfiler.AnchorMap[label]
	if !ok {
		if GlobalProfiler.Counter >= len(GlobalProfiler.Anchors) {
			fmt.Println("Warning: Too many profile blocks, skipping:", label)
			return func() {} //No-op function
		}
		anchorIndex = GlobalProfiler.Counter
		GlobalProfiler.AnchorMap[label] = anchorIndex
		GlobalProfiler.Counter++
		GlobalProfiler.Anchors[anchorIndex].Label = label // Set label only once
	}

	startTime := CpuTimer()
	return func() {
		GlobalProfiler.Mutex.Lock()
		defer GlobalProfiler.Mutex.Unlock()
		elapsed := CpuTimer() - startTime
		anchor := &GlobalProfiler.Anchors[anchorIndex]
		anchor.TSCElapsed += elapsed
		anchor.HitCount++
	}
}

// TimeFunction is a helper function to time an entire function duration
func TimeFunction(funcName ...string) func() {
	var blockLabel string

	if len(funcName) > 0 {
		blockLabel = funcName[0]
	} else {
		// Get the name of the calling function
		pc, _, _, ok := runtime.Caller(1) // 1 level up the stack
		if !ok {
			blockLabel = "unknown"
		} else {
			funcName := runtime.FuncForPC(pc).Name()
			parts := strings.Split(funcName, "/")
			blockLabel = parts[len(parts)-1] // Last part of the path
		}
	}
	return TimeBlock(blockLabel)
}
