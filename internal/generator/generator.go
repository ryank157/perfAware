package generator

import (
	"bufio"
	"encoding/binary"
	"log"
	"main/internal/shared"
	"os"
)

func GenerateDataSetAndAnswerFiles(spread string, seed int, numPoints int) float64 {
	// Create files
	binFile, err := os.Create("data.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer binFile.Close()

	outputFile, err := os.Create("data.json")
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	bufferedWriter := bufio.NewWriter(outputFile)
	defer bufferedWriter.Flush()

	_, err = bufferedWriter.WriteString("{\n  \"pairs\": [\n")
	if err != nil {
		log.Fatal(err)
	}

	avgDistance := shared.GeneratePoints(seed, numPoints, spread, bufferedWriter)

	_, err = bufferedWriter.WriteString("   ]\n}")
	if err != nil {
		log.Fatal(err)
	}

	err = binary.Write(binFile, binary.LittleEndian, avgDistance)
	if err != nil {
		log.Fatal(err)
	}

	// Now we Write `avgDistance` to the  answer file
	// Create a `answer.f64` File for New Validation
	answerFile, err := os.Create("answer.f64")
	if err != nil {
		log.Fatalf("Error creating answer.f64: %v", err)
	}
	defer answerFile.Close()

	// Write the average distance to the answer file with 16 decimals of precision.
	err = binary.Write(answerFile, binary.LittleEndian, avgDistance)
	//_, err = fmt.Fprintf(answerFile, "%.16f\n", avgDistance) //this is a STRING VERSION

	if err != nil {
		log.Fatal(err)
	}

	return avgDistance
}
