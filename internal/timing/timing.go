//go:build !timing

package timing

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
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
	TSCElapsedExclusive atomic.Uint64
	TSCElapsedInclusive atomic.Uint64
	HitCount            atomic.Uint64
	Label               string
}

// Profiler manages the profiling data.
type Profiler struct {
	Anchors       [4096]ProfileAnchor
	AnchorMap     sync.Map
	StartTSC      atomic.Uint64
	EndTSC        atomic.Uint64
	Counter       atomic.Int32 // Use atomic for concurrent access
	currentParent atomic.Int32 // Use atomic for currentParent
}

// GlobalProfiler is the global instance of the profiler.
var GlobalProfiler = Profiler{}

// BeginProfile starts the profiling session.
func BeginProfile() {
	GlobalProfiler.StartTSC.Store(CpuTimer())

	// Initialize the root anchor
	GlobalProfiler.Anchors[0].Label = "Root"
	GlobalProfiler.Anchors[0].TSCElapsedExclusive.Store(0)
	GlobalProfiler.Anchors[0].TSCElapsedInclusive.Store(0)
	GlobalProfiler.Anchors[0].HitCount.Store(0)
	GlobalProfiler.AnchorMap.Store("Root", 0)
	GlobalProfiler.Counter.Store(1)
	GlobalProfiler.currentParent.Store(0)

}

// EndAndPrintProfile ends the profiling session and prints the results.
func EndAndPrintProfile() {
	GlobalProfiler.EndTSC.Store(CpuTimer())

	cpuFreq := EstimateCPUFrequency()
	totalCPUElapsed := GlobalProfiler.EndTSC.Load() - GlobalProfiler.StartTSC.Load()

	if cpuFreq > 0 {
		fmt.Printf("\nTotal time: %.4fms (CPU freq %d)\n", 1000.0*float64(totalCPUElapsed)/float64(cpuFreq), cpuFreq)
	}

	for i := 0; i < int(GlobalProfiler.Counter.Load()); i++ {
		anchor := &GlobalProfiler.Anchors[i]
		if anchor.TSCElapsedExclusive.Load() > 0 {
			printTimeElapsed(totalCPUElapsed, anchor)
		}
	}
}

func printTimeElapsed(totalTSCElapsed uint64, anchor *ProfileAnchor) {
	exclusive := anchor.TSCElapsedExclusive.Load()

	if exclusive > totalTSCElapsed {
		fmt.Printf("WARNING: Invalid timing for %s - elapsed time exceeds total time\n", anchor.Label)
		exclusive = totalTSCElapsed
	}

	percent := 100.0 * (float64(exclusive) / float64(totalTSCElapsed))
	fmt.Printf("  %s[%d]: %d (%.2f%%", anchor.Label, anchor.HitCount.Load(), exclusive, percent)

	inclusive := anchor.TSCElapsedInclusive.Load()
	if inclusive != exclusive {
		percentWithChildren := 100.0 * (float64(inclusive) / float64(totalTSCElapsed))
		fmt.Printf(", %.2f%% w/children", percentWithChildren)
	}
	fmt.Printf(")\n")
}

var enableTimingStr = "false"

func IsTimingEnabled() bool {
	return enableTimingStr == "true"
}

// TimeBlock is a function that returns a function to stop the timer
func TimeBlock(label string) func() {
	if !IsTimingEnabled() {
		return func() {}
	}

	parentIndex := GlobalProfiler.currentParent.Load()
	anchorIndex := getOrAddAnchor(label)

	oldInclusive := GlobalProfiler.Anchors[anchorIndex].TSCElapsedInclusive.Load()
	GlobalProfiler.currentParent.Store(anchorIndex)
	startTime := CpuTimer()

	return func() {

		elapsed := CpuTimer() - startTime
		GlobalProfiler.currentParent.Store(parentIndex)

		anchor := &GlobalProfiler.Anchors[anchorIndex]
		parent := &GlobalProfiler.Anchors[parentIndex]

		//1. Subtract elapsed time from parent's exlusive time
		if parentIndex > 0 {
			parent.TSCElapsedExclusive.Add(^(elapsed - 1))
		}

		//2. Add elapsed time to current anchor's exclusive time
		anchor.TSCElapsedExclusive.Add(elapsed)

		//3. Set inclusive time (total time including children)
		anchor.TSCElapsedInclusive.Store(oldInclusive + elapsed)

		//4. Increment hit count
		anchor.HitCount.Add(1)

	}
}

func getOrAddAnchor(label string) int32 {
	val, ok := GlobalProfiler.AnchorMap.Load(label)
	if ok {
		return val.(int32)
	}
	newIndex := GlobalProfiler.Counter.Add(1) - 1
	if int(newIndex) >= len(GlobalProfiler.Anchors) {
		fmt.Println("Warning: Too many profile blocks, skipping:", label)
		return 0 // Return a dummy anchor index
	}

	GlobalProfiler.AnchorMap.Store(label, newIndex)
	anchor := &GlobalProfiler.Anchors[newIndex]
	anchor.Label = label
	return newIndex
}

// TimeFunction is a helper function to time an entire function duration
func TimeFunction(funcName ...string) func() {
	if !IsTimingEnabled() {
		return func() {}
	}

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
