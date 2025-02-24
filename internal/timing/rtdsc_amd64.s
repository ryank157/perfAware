//go:build amd64
//go:build linux

#include "textflag.h"

// Exported function
TEXT ·CpuTimer(SB),NOSPLIT,$0-8
    // RDTSC
    BYTE $0x0F; BYTE $0x31

    SHLQ $32, DX
    ORQ AX, DX
    MOVQ DX, ret+0(FP)
    RET
