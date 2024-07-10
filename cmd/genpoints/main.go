package main

import (
	"encoding/csv"
	"log"
	"math"
	"os"
	"strconv"
)

/// A -AZ- Z -ZB- B

// Z --- X --- X --- X --- Z
// |     |     |     |     |
// X --- Y --- Y --- Y --- X
// |     |     |     |     |
// X --- Y --- Y --- Y --- X
// |     |     |     |     |
// X --- Y --- Y --- Y --- X
// |     |     |     |     |
// Z --- X --- X --- X --- U

type R struct {
	rr [][]float64
}

func rr(lat1, lon1, lat2, lon2 float64, result *R) {
	midlat, midlon := midPoint(
		lat1,
		lon1,
		lat2,
		lon2,
	)

	d := distance(lat1, lon1, midlat, midlon, "K")

	if d <= 1 {
		return
	}

	result.rr = append(result.rr, []float64{midlat, midlon})

	rr(lat1, lon1, midlat, midlon, result)
	rr(midlat, midlon, lat2, lon2, result)
}

func main() {
	startlat, err := strconv.ParseFloat(os.Args[1], 64)

	if err != nil {
		panic(err)
	}

	startlng, err := strconv.ParseFloat(os.Args[2], 64)

	if err != nil {
		panic(err)
	}

	z1 := []float64{startlat, startlng}
	// z2 := []float64{13.981117, 100.798338}
	//
	// var result R
	//
	// rr(
	// 	z1[0],
	// 	z1[1],
	// 	z2[0],
	// 	z2[1],
	// 	&result,
	// )
	//
	// xpoints := append(result.rr, z1, z2)

	xpoints := [][]float64{
		z1,
	}

	for i := 1; i <= 50; i++ {
		xlat, xlon := movePoint(z1[0], z1[1], 1*1000*float64(i), 90)
		xpoints = append(xpoints, []float64{xlat, xlon})
	}

	var respoints [][]float64

	for _, xpoint := range xpoints {
		for i := 1; i <= 50; i++ {
			ylat, ylon := movePoint(xpoint[0], xpoint[1], 1*1000*float64(i), 180)
			respoints = append(respoints, []float64{ylat, ylon})
		}
	}

	allpoints := append(xpoints, respoints...)

	log.Println("all points", len(allpoints))

	cf, err := os.Create("points.csv")

	if err != nil {
		panic(err)
	}

	defer cf.Close()

	cw := csv.NewWriter(cf)

	defer cw.Flush()

	_ = cw.Write([]string{"Lat", "Long"})

	for _, p := range allpoints {
		_ = cw.Write([]string{
			strconv.FormatFloat(p[0], 'f', -1, 64),
			strconv.FormatFloat(p[1], 'f', -1, 64),
		})
	}
}

func midPoint(lat1, lon1, lat2, lon2 float64) (float64, float64) {
	dLon := math.Pi * (lon2 - lon1) / 180

	lat1 = math.Pi * lat1 / 180
	lat2 = math.Pi * lat2 / 180
	lon1 = math.Pi * lon1 / 180

	Bx := math.Cos(lat2) * math.Cos(dLon)
	By := math.Cos(lat2) * math.Sin(dLon)
	lat3 := math.Atan2(math.Sin(lat1)+math.Sin(lat2), math.Sqrt((math.Cos(lat1)+Bx)*(math.Cos(lat1)+Bx)+By*By))
	lon3 := lon1 + math.Atan2(By, math.Cos(lat1)+Bx)

	rlat := lat3 * 180 / math.Pi
	rlon := lon3 * 180 / math.Pi

	return rlat, rlon
}

func distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64, unit ...string) float64 {
	radlat1 := float64(math.Pi * lat1 / 180)
	radlat2 := float64(math.Pi * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(math.Pi * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515

	if len(unit) > 0 {
		if unit[0] == "K" {
			dist = dist * 1.609344
		} else if unit[0] == "N" {
			dist = dist * 0.8684
		}
	}

	return dist
}

func movePoint(lat, lon, distance, angle float64) (newLat, newLon float64) {
	const R = 6378137 // radius of the Earth (meters)

	// degrees to radians
	lat = lat * math.Pi / 180
	lon = lon * math.Pi / 180
	angle = angle * math.Pi / 180

	newLat = math.Asin(math.Sin(lat)*math.Cos(distance/R) + math.Cos(lat)*math.Sin(distance/R)*math.Cos(angle))
	newLon = lon + math.Atan2(math.Sin(angle)*math.Sin(distance/R)*math.Cos(lat), math.Cos(distance/R)-math.Sin(lat)*math.Sin(newLat))

	// radians to degrees
	newLat = newLat * 180 / math.Pi
	newLon = newLon * 180 / math.Pi

	return newLat, newLon
}
