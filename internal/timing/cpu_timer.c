#include <stdint.h>
#ifdef _MSC_VER // MSVC compiler
#include <intrin.h>  // For __rdtsc intrinsic
#pragma intrinsic(__rdtsc) // Enable the intrinsic
#else
#include <x86intrin.h> // For __rdtsc
#endif
uint64_t cpu_timer() {
#ifdef _MSC_VER
return __rdtsc();
#else
return __rdtsc();
#endif
}
