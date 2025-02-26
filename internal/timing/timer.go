//go:build amd64 && linux

package timing

/*
#cgo CFLAGS: -Wall -Werror
#include "cpu_timer.h"
*/
import "C"

// CpuTimer calls the C function to read the CPU timestamp counter.
func CpuTimer() uint64 {
    return uint64(C.cpu_timer())
}
