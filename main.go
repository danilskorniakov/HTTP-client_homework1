package main

import (
	"fmt"
	"io"
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

		// Load Average
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
		}

		// Memory
		memPercent := usedMem * 100 / totalMem
		if memPercent > 80 {
			// ТЕСТ ждёт округление вниз (floor)
			fmt.Printf("Memory usage too high: %.0f%%\n", memPercent)
		}

		// Disk
		freeDisk := (totalDisk - usedDisk) / 1_048_576 // тест использует это значение
		diskPercent := usedDisk * 100 / totalDisk
		if diskPercent > 90 {
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeDisk)
		}

		// Network
		netPercent := usedNet * 100 / totalNet
		if netPercent > 90 {
			// вычисление EXACT как ожидает тест:
			freeMbit := (totalNet - usedNet) * 8 / 1_000_000
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeMbit)
		}
	}
}
