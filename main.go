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
	errorCount := 0

	for i := 0; i < 7; i++ {
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != 200 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
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
			continue
		}

		fields := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(fields) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			continue
		}

		// Сбрасываем счетчик ошибок при успешном запросе
		errorCount = 0

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
			fmt.Printf("Memory usage too high: %.0f%%\n", math.Floor(memPercent))
		}

		// Disk
		freeDiskBytes := totalDisk - usedDisk
		freeDiskMB := freeDiskBytes / (1024 * 1024)
		diskPercent := usedDisk * 100 / totalDisk
		if diskPercent > 90 {
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", math.Floor(freeDiskMB))
		}

		// Network
		netPercent := usedNet * 100 / totalNet
		if netPercent > 90 {
			// Вариант 1: свободная полоса в мегабайтах в секунду
			freeMBperSec := (totalNet - usedNet) / (1024 * 1024)
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", math.Floor(freeMBperSec))
		}
	}
}
