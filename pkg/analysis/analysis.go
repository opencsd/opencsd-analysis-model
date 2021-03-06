package analysis

import (
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mackerelio/go-osstat/memory"
)

type Analysis struct {
	Cpu    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Energy float64 `json:"energy"`
}

func cpuMeasure() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val
				if i == 4 {
					idle = val
				}
			}
			return
		}
	}
	return
}

var ssdCpu, csdCpu = 0, 0

func GetCPU(cpuChan chan float64) {
	idle0, total0 := cpuMeasure()
	time.Sleep(1 * time.Second)
	idle1, total1 := cpuMeasure()
	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

	cpuBenefit := csdCpu / ssdCpu
	log.Printf("CPU usage is %f%% [busy: %f, total: %f]\n", cpuUsage, totalTicks-idleTicks, totalTicks)
	cpuChan <- cpuUsage

	fmt.Println(cpuBenefit)
}

func GetMem(memChan chan float64) {
	mem, err := memory.Get()
	if err != nil {
		log.Println(err)
	}
	log.Println(mem.Total)
	log.Println(mem.Used)
	log.Println(mem.Cached)
	log.Println(mem.Free)
	memUsage := 100 * (float64(mem.Used) / float64(mem.Total))
	log.Println(memUsage)

	memChan <- memUsage
}

func GetMemory() {
	PrintMemUsage()

	for i := 0; i < 4; i++ {

		PrintMemUsage()
		time.Sleep(time.Second)
	}

	PrintMemUsage()

	runtime.GC()
	PrintMemUsage()
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	param := b
	return param / 1024 / 1024
}
