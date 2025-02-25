package shared

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/ryank157/perfAware/internal/timing"
)

const (
	uniform = "uniform"
	cluster = "cluster"
)

type Point struct {
	X, Y float64
}

type Cluster struct {
	Xmin, Xmax float64
	Ymin, Ymax float64
}

type HaversinePair struct {
	X0 float64
	Y0 float64
	X1 float64
	Y1 float64
}

type HaversineData struct {
	pairs []HaversinePair
}

func GeneratePoints(seed int, numPoints int, spreadType string, writer *bufio.Writer) float64 {
	defer timing.TimeFunction()()
	r := rand.New(rand.NewSource(int64(seed)))
	sum := 0.0
	isFirst := true

	var clusters []Cluster
	var ptsPerCluster int

	if spreadType == uniform {
		clusters = []Cluster{{Xmin: -180, Xmax: 180, Ymin: -90, Ymax: 90}}
		ptsPerCluster = numPoints
	} else if spreadType == cluster {
		const (
			NumClusters int     = 64
			ClusterSize float64 = 16
		)
		ptsPerCluster = numPoints / NumClusters
		clusters = make([]Cluster, NumClusters)
		for i := range NumClusters {
			centerX := r.Float64()*360 - 180
			centerY := r.Float64()*180 - 90

			clusters[i] = Cluster{
				Xmin: math.Max(centerX-ClusterSize, -180),
				Xmax: math.Min(centerX+ClusterSize, 180),
				Ymin: math.Max(centerY-ClusterSize, -90),
				Ymax: math.Min(centerY+ClusterSize, 90),
			}
		}
	} else {
		log.Fatalf("Invalid spread type: %s", spreadType)
		return 0 // Unreachable
	}

	clusterIndex := 0
	pointsInCluster := 0

	for range numPoints {
		if pointsInCluster >= ptsPerCluster {
			clusterIndex++
			pointsInCluster = 0
			if clusterIndex >= len(clusters) {
				clusterIndex = 0
			}
		}

		c := clusters[clusterIndex]
		p0 := Point{r.Float64()*(c.Xmax-c.Xmin) + c.Xmin, r.Float64()*(c.Ymax-c.Ymin) + c.Ymin}
		p1 := Point{r.Float64()*(c.Xmax-c.Xmin) + c.Xmin, r.Float64()*(c.Ymax-c.Ymin) + c.Ymin}
		dist := Haversine(HaversinePair{p0.X, p0.Y, p1.X, p1.Y})
		sum += dist

		pair := HaversinePair{
			X0: p0.X,
			Y0: p0.Y,
			X1: p1.X,
			Y1: p1.Y,
		}

		pairJSON := fmt.Sprintf(`{"X0":%.15f,"Y0":%.15f,"X1":%.15f,"Y1":%.15f}`, pair.X0, pair.Y0, pair.X1, pair.Y1)

		if !isFirst {
			_, err := writer.WriteString(",\n")
			if err != nil {
				log.Fatal(err)
			}
		}
		_, err := writer.WriteString("    ")
		if err != nil {
			log.Fatal(err)
		}
		_, err = writer.WriteString(pairJSON)
		if err != nil {
			log.Fatal(err)
		}
		isFirst = false
		pointsInCluster++
	}
	return sum / float64(numPoints)

}

func Radians(d float64) float64 {
	return d * math.Pi / 180
}

func Square(x float64) float64 {
	return math.Pow(x, 2)
}

const EarthRadius = 6372.8 // km

func Haversine(pair HaversinePair) float64 {
	defer timing.TimeFunction()()
	dX := Radians(pair.X1 - pair.X0)
	dY := Radians(pair.Y1 - pair.Y0)
	y0 := Radians(pair.Y0)
	y1 := Radians(pair.Y1)

	a := Square(math.Sin(dY/2)) + math.Cos(y0)*math.Cos(y1)*Square(math.Sin(dX/2))

	c := 2 * EarthRadius * math.Asin(math.Sqrt(a))

	return c
}

func SumHaversineDistances(pairCount int, pairs []HaversinePair) float64 {
	defer timing.TimeFunction()()
	sum := 0.0
	sumCoef := 1.0 / float64(pairCount)
	for pairIndex := range pairCount {
		pair := pairs[pairIndex]
		dist := Haversine(pair)
		sum += sumCoef * dist
	}
	return sum
}
