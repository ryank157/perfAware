package timing

import (
	"fmt"
	"runtime"
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
