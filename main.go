package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	// Пример данных от сервера (7 сообщений)
	responses := []string{
		"36,2147483648,1720000000,5497558138880,4940000000000,104857600,95000000",
		"22,2147483648,1600000000,5497558138880,4500000000000,104857600,90000000",
		"70,2147483648,1800000000,5497558138880,5200000000000,104857600,102000000",
		"15,2147483648,1500000000,5497558138880,4300000000000,104857600,85000000",
		"52,2147483648,1900000000,5497558138880,5000000000000,104857600,110000000",
		"3,2147483648,1400000000,5497558138880,4000000000000,104857600,80000000",
		"47,2147483648,1650000000,5497558138880,4800000000000,104857600,90000000",
	}

	errorCount := 0

	for _, line := range responses {
		fields := strings.Split(line, ",")
		if len(fields) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			continue
		}

		load, _ := strconv.Atoi(fields[0])
		memTotal, _ := strconv.ParseInt(fields[1], 10, 64)
		memUsed, _ := strconv.ParseInt(fields[2], 10, 64)
		diskTotal, _ := strconv.ParseInt(fields[3], 10, 64)
		diskUsed, _ := strconv.ParseInt(fields[4], 10, 64)
		netTotal, _ := strconv.ParseInt(fields[5], 10, 64)
		netUsed, _ := strconv.ParseInt(fields[6], 10, 64)

		// Load Average
		if load > 30 {
			fmt.Printf("Load Average is too high: %d\n", load)
		}

		// Memory usage
		memPercent := int(memUsed * 100 / memTotal) // целое число
		if memPercent > 80 {
			fmt.Printf("Memory usage too high: %d%%\n", memPercent)
		}

		// Disk space
		diskFreeMb := int((diskTotal - diskUsed) / 1024 / 1024)
		if diskUsed*100/diskTotal > 90 {
			fmt.Printf("Free disk space is too low: %d Mb left\n", diskFreeMb)
		}

		// Network usage
		netFreeMbit := int((netTotal - netUsed) * 8 / 1000000) // перевод в мегабиты
		if netUsed*100/netTotal > 90 {
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", netFreeMbit)
		}
	}
}
