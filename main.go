package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	const url = "http://srv.msk01.gigacorp.local/_stats"

	errorCount := 0

	for {
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != 200 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(time.Second)
			continue
		}

		fields := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(fields) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(time.Second)
			continue
		}

		loadAvg, _ := strconv.ParseFloat(fields[0], 64)

		totalMem, _ := strconv.ParseFloat(fields[1], 64)
		usedMem, _ := strconv.ParseFloat(fields[2], 64)

		totalDisk, _ := strconv.ParseFloat(fields[3], 64)
		usedDisk, _ := strconv.ParseFloat(fields[4], 64)

		totalNet, _ := strconv.ParseFloat(fields[5], 64)
		usedNet, _ := strconv.ParseFloat(fields[6], 64)

		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
		}

		memUsagePercent := usedMem / totalMem * 100
		if memUsagePercent > 80 {
			fmt.Printf("Memory usage too high: %.0f%%\n", memUsagePercent)
		}

		freeDisk := (totalDisk - usedDisk) / 1_000_000
		if usedDisk/totalDisk*100 > 90 {
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeDisk)
		}

		freeMbit := (totalNet - usedNet) * 8 / 1_000_000
		if usedNet/totalNet > 0.9 {
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeMbit)
		}

		time.Sleep(1 * time.Second)
	}
}
