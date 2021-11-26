package tools

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"math"
	"runtime"
	"time"
)

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	msg := fmt.Sprintf("HeapAlloc = %v MiB  TotalAlloc = %v MiB  Sys = %v MiB  tNumGC = %v  PauseTotal %vms HeapObjects = %v  Goroutine = %v",
		bToMb(m.HeapAlloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC, m.PauseTotalNs/100000, m.HeapObjects, runtime.NumGoroutine())
	println(msg)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

type HostInfo struct {
	Total     int32 // 总内存MB
	Used      int32 // 已用内存MB
	Avaliable int32 // 可用内存MB
	CpuUsage  int32 // 使用百分比
	CpuCores  int32 // CPU线程数
}

func GetHostInfo() (HostInfo, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return HostInfo{}, err
	}
	percents, err := cpu.Percent(time.Second, false)
	if err != nil {
		return HostInfo{}, err
	}
	cores, err := cpu.Counts(true)
	if err != nil {
		return HostInfo{}, err
	}
	return HostInfo{
		Total:     int32(bToMb(v.Total)),
		Used:      int32(bToMb(v.Used)),
		Avaliable: int32(bToMb(v.Available)),
		CpuUsage:  int32(math.Round(percents[0])),
		CpuCores:  int32(cores),
	}, nil
}
