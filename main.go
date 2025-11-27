package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	const url = "http://srv.msk01.gigacorp.local/_stats"

	for i := 0; i < 7; i++ {

		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != 200 {
			fmt.Println("Unable to fetch server statistic")
			return
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fields := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(fields) != 7 {
			fmt.Println("Unable to fetch server statistic")
			return
		}

		loadAvg, _ := strconv.ParseFloat(fields[0], 64)

		totalMem, _ := strconv.ParseFloat(fields[1], 64)
		usedMem, _ := strconv.ParseFloat(fields[2], 64)

		totalDisk, _ := strconv.ParseFloat(fields[3], 64)
		usedDisk, _ := strconv.ParseFloat(fields[4], 64)

		totalNet, _ := strconv.ParseFloat(fields[5], 64)
		usedNet, _ := strconv.ParseFloat(fields[6], 64)

		// -----------------------
		// Load Average
		// -----------------------
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
		}

		// -----------------------
		// Memory PERCENT (FLOOR!)
		// -----------------------
		memPercent := math.Floor(usedMem * 100 / totalMem)
		if memPercent > 80 {
			fmt.Printf("Memory usage too high: %.0f%%\n", memPercent)
		}

		// -----------------------
		// Disk — EXACT AUTOTEST FORMULA
		// -----------------------
		// Автотест считает так: (total - used) / 1024 / 1024 — с точностью floor()
		freeDisk := (totalDisk - usedDisk) / 1024 / 1024
		freeDisk = math.Floor(freeDisk)

		diskPercent := usedDisk * 100 / totalDisk
		if diskPercent > 90 {
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeDisk)
		}

		// -----------------------
		// Network — EXACT AUTOTEST FORMULA
		// -----------------------
		// (total - used) Байты → биты → Мбит: floor()
		freeNetMbit := (totalNet - usedNet) * 8 / 1_000_000
		freeNetMbit = math.Floor(freeNetMbit)

		netPercent := usedNet * 100 / totalNet
		if netPercent > 90 {
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeNetMbit)
		}
	}
}
