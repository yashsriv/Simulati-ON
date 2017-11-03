package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"time"

	gnuplot "github.com/sbinet/go-gnuplot"
)

const timeStep = 0.01
const g = 1

// Vector3d represents a vector in 3 dimensional space.
type Vector3d struct {
	X float64
	Y float64
	Z float64
}

func (v1 Vector3d) multiplyVal(val float64) Vector3d {
	var v Vector3d
	v.X = v1.X * val
	v.Y = v1.Y * val
	v.Z = v1.Z * val
	return v
}

func (v1 Vector3d) addVec(v2 Vector3d) Vector3d {
	var v Vector3d
	v.X = v1.X + v2.X
	v.Y = v1.Y + v2.Y
	v.Z = v1.Z + v2.Z
	return v
}

func (v1 Vector3d) subVec(v2 Vector3d) Vector3d {
	var v Vector3d
	v.X = v1.X - v2.X
	v.Y = v1.Y - v2.Y
	v.Z = v1.Z - v2.Z
	return v
}

// Planet represents one of our planets.
type Planet struct {
	Mass float64
	Pos  Vector3d
	Vel  Vector3d
}

func calculateNewPos(planets []Planet, planetid int) Planet {
	var acc, pos Vector3d
	var t, factor, m float64
	var curPlanet = planets[planetid]
	for i, planet := range planets {
		if i != planetid {
			a := planet.Pos
			b := curPlanet.Pos
			del := a.subVec(b)
			m = planet.Mass

			t = math.Sqrt(del.X*del.X + del.Y*del.Y + del.Z*del.Z)
			factor = (g * m) / (t * t * t)
			acc = acc.addVec(del.multiplyVal(factor))
		}
	}
	pos = curPlanet.Pos
	// ds = udt + 0.5 a(dt)^2
	pos = pos.addVec(curPlanet.Vel.multiplyVal(timeStep)) // Add u*t
	delvel := acc.multiplyVal(timeStep)                   // dv = a * dt
	t = 0.5 * timeStep * timeStep                         // 0.5 (dt)*2
	pos = pos.addVec(acc.multiplyVal(t))                  // Add 0.5 a (dt)*2

	return Planet{
		Pos:  pos,
		Vel:  planets[planetid].Vel.addVec(delvel),
		Mass: planets[planetid].Mass,
	}

}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}

// Job represents a job for our worker pool
type Job struct {
	Current []Planet
	ID      int
}

func worker(jobs <-chan Job, results chan<- Planet) {
	for job := range jobs {
		results <- calculateNewPos(job.Current, job.ID)
	}
}

func loopOnce(planets []Planet, jobs chan<- Job, results <-chan Planet) []Planet {
	defer elapsed("loopOnce")()
	var newplanets = make([]Planet, len(planets))
	for j := range planets {
		// Parallelize stuff
		// jobs <- Job{
		// 	Current: planets,
		// 	ID:      j,
		// }
		newplanets[j] = calculateNewPos(planets, j)
	}
	// for j := range planets {
	// 	newplanets[j] = <-results
	// }
	return newplanets
}

func main() {
	raw, err := ioutil.ReadFile("./planets.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var planets []Planet
	if err = json.Unmarshal(raw, &planets); err != nil {
		panic(err)
	}

	// Setup worker pool
	jobs := make(chan Job, len(planets))
	results := make(chan Planet, len(planets))

	for w := 1; w <= runtime.NumCPU(); w++ {
		go worker(jobs, results)
	}

	p, err := gnuplot.NewPlotter("", true, false)
	if err != nil {
		errString := fmt.Sprintf("** err: %v\n", err)
		panic(errString)
	}
	defer p.Close()

	for {
		newplanets := loopOnce(planets, jobs, results)
		// Display Here!!

		p.CheckedCmd("set term %s size %d, %d", "x11", 1500, 1200)
		p.CheckedCmd("set xrange [%d:%d]", -500, 500)
		p.CheckedCmd("set yrange [%d:%d]", -500, 500)
		p.CheckedCmd("set zrange [%d:%d]", -5, 5)
		xarr := make([]float64, len(planets))
		yarr := make([]float64, len(planets))
		zarr := make([]float64, len(planets))
		for j := 0; j < len(planets); j++ {
			xarr[j] = planets[j].Pos.X
			yarr[j] = planets[j].Pos.Y
			zarr[j] = planets[j].Pos.Z
		}
		p.PlotXYZ(xarr, yarr, zarr, "N-Body Simulation")

		planets = newplanets
		p.ResetPlot()
	}
}
