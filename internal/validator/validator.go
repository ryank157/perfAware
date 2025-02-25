package validator

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"unsafe"

	"github.com/ryank157/perfAware/internal/shared"
)

type ValidateTimers struct {
	Begin      uint64
	Read       uint64
	MiscSetup  uint64
	Parse      uint64
	Sum        uint64
	MiscOutput uint64
	End        uint64
}

func ValidateData(inputFileName string, answersFileName string) {
	inputJSONBuffer, err := shared.ReadEntireFile(inputFileName)
	if err != nil {
		log.Fatalf("Error reading JSON file: %v", err)
	}
	defer shared.FreeBuffer(&inputJSONBuffer) // VERY IMPORTANT: Release memory

	minimumJSONPairEncoding := 6 * 4 // Minimal size based on C's u32

	maxPairCount := inputJSONBuffer.Count / int64(minimumJSONPairEncoding)
	if maxPairCount > 0 {
		parsedValuesBuffer := shared.AllocateBuffer(maxPairCount * int64(unsafe.Sizeof(shared.HaversinePair{}))) //Use unsafe.Sizeof!
		if parsedValuesBuffer.Count > 0 {
			defer shared.FreeBuffer(&parsedValuesBuffer) //VERY IMPORTANT: Release memory

			// Create a slice of HaversinePair using unsafe.Slice to interpret the raw bytes
			// as a slice of HaversinePair values directly.  This is similar to casting in C,
			// and it's essential for working with the byte-oriented buffer.
			pairs := unsafe.Slice((*shared.HaversinePair)(unsafe.Pointer(&parsedValuesBuffer.Data[0])), maxPairCount)

			pairCount := ParseHaversinePairs(inputJSONBuffer.Data, int(maxPairCount), pairs)
			sum := shared.SumHaversineDistances(pairCount, pairs[:pairCount]) // Slice only the populated part of the pair
			fmt.Printf("Input size: %d\n", inputJSONBuffer.Count)
			fmt.Printf("Pair count: %d\n", pairCount)
			fmt.Printf("Haversine sum: %.16f\n", sum)

			answersF64Buffer, err := shared.ReadEntireFile(answersFileName)
			if err != nil {
				log.Printf("Warning: Error reading answer file: %v", err) // Don't crash if validation fails
				return                                                    // Important to RETURN here
			}
			defer shared.FreeBuffer(&answersF64Buffer)

			if answersF64Buffer.Count != int64(unsafe.Sizeof(float64(0))) {
				log.Printf("Warning: Answer file size incorrect. Expected %d bytes,  got %d. Aborting  validation", int64(unsafe.Sizeof(float64(0))), answersF64Buffer.Count)
				return // Important to RETURN here
			}

			//was sizeof(double)  in C world

			//THIS IS WHAT C VERSION DOES.  We read directly *from the buffer.* No need
			//to play with the last element of the array and similar nonsense!
			bits := binary.LittleEndian.Uint64(answersF64Buffer.Data)
			refSumFloat := math.Float64frombits(bits)

			fmt.Printf("\nValidation:\n")
			fmt.Printf("Reference sum: %.16f\n", refSumFloat)
			fmt.Printf("Difference   :    %.16f\n", sum-refSumFloat)
			fmt.Printf("\n")

		} else {
			fmt.Fprintf(os.Stderr, "ERROR: Could not allocate memory for parsed values.\n")
			os.Exit(1) // Fatal error, so exit
		}
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: Malformed input JSON\n")
		os.Exit(1) // Fatal error, so exit
	}

}
