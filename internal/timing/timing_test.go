package timing

import (
	"fmt"
	"runtime"
	"testing"
)

func TestOSTimer(t *testing.T) {
	runtime.LockOSThread() // Important: do not let scheduler switch us out
	defer runtime.UnlockOSThread()

	osStart := OsTimer()
	var osEnd int64
	var osElapsed int64

	for osElapsed < osTimerFreq {
		osEnd = OsTimer()
		osElapsed = osEnd - osStart
	}

	fmt.Printf("OS timer: %d -> %d = %d elapsed\n", osStart, osEnd, osElapsed)
	fmt.Printf("OS seconds: %.4f\n", float64(osElapsed)/float64(osTimerFreq))
}

func TestCPUTimer(t *testing.T) {
	runtime.LockOSThread() // Important: do not let scheduler switch us out
	defer runtime.UnlockOSThread()

	cpuStart := CpuTimer()
	osStart := OsTimer()
	var osEnd int64
	var osElapsed int64

	for osElapsed < osTimerFreq {
		osEnd = OsTimer()
		osElapsed = osEnd - osStart
	}

	cpuEnd := CpuTimer()
	cpuElapsed := cpuEnd - cpuStart

	fmt.Printf("OS timer: %d -> %d = %d elapsed\n", osStart, osEnd, osElapsed)
	fmt.Printf("OS seconds: %.4f\n", float64(osElapsed)/float64(osTimerFreq))
	fmt.Printf("CPU timer: %d -> %d = %d elapsed\n", cpuStart, cpuEnd, cpuElapsed)
}

func TestCPUFreq(t *testing.T) {
	cpuFreq := EstimateCPUFrequency()
	fmt.Printf("Estimated CPU Frequency: %d\n", cpuFreq)
}
