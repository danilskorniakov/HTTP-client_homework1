package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	statsURL             = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval         = 1 * time.Second
	maxConsecutiveErrors = 3

	// Пороговые значения
	loadAverageThreshold  = 30
	memoryUsageThreshold  = 80 // процент
	diskUsageThreshold    = 90 // процент
	networkUsageThreshold = 90 // процент
)

type ServerStats struct {
	LoadAverage      int64
	TotalMemory      int64
	UsedMemory       int64
	TotalDisk        int64
	UsedDisk         int64
	NetworkBandwidth int64
	NetworkUsage     int64
}

func main() {
	consecutiveErrors := 0
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		stats, err := fetchStats(client)
		if err != nil {
			consecutiveErrors++
			if consecutiveErrors >= maxConsecutiveErrors {
				fmt.Println("Unable to fetch server statistic")
			}
		} else {
			consecutiveErrors = 0
			checkAndReport(stats)
		}

		time.Sleep(pollInterval)
	}
}

func fetchStats(client *http.Client) (*ServerStats, error) {
	resp, err := client.Get(statsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseStats(string(body))
}

func parseStats(data string) (*ServerStats, error) {
	// Убираем пробельные символы
	data = strings.TrimSpace(data)

	// Используем bufio.Scanner для разбора по запятым
	scanner := bufio.NewScanner(strings.NewReader(data))
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), ","); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	var values []int64
	for scanner.Scan() {
		val, err := strconv.ParseInt(strings.TrimSpace(scanner.Text()), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value: %v", err)
		}
		values = append(values, val)
	}

	if len(values) != 7 {
		return nil, fmt.Errorf("expected 7 values, got %d", len(values))
	}

	return &ServerStats{
		LoadAverage:      values[0],
		TotalMemory:      values[1],
		UsedMemory:       values[2],
		TotalDisk:        values[3],
		UsedDisk:         values[4],
		NetworkBandwidth: values[5],
		NetworkUsage:     values[6],
	}, nil
}

func checkAndReport(stats *ServerStats) {
	// Проверка Load Average
	if stats.LoadAverage > loadAverageThreshold {
		fmt.Printf("Load Average is too high: %d\n", stats.LoadAverage)
	}

	// Проверка использования памяти
	if stats.TotalMemory > 0 {
		memoryPercent := (stats.UsedMemory * 100) / stats.TotalMemory
		if memoryPercent > memoryUsageThreshold {
			fmt.Printf("Memory usage too high: %d%%\n", memoryPercent)
		}
	}

	// Проверка дискового пространства
	if stats.TotalDisk > 0 {
		diskPercent := (stats.UsedDisk * 100) / stats.TotalDisk
		if diskPercent > diskUsageThreshold {
			freeDiskBytes := stats.TotalDisk - stats.UsedDisk
			freeDiskMb := freeDiskBytes / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeDiskMb)
		}
	}

	// Проверка сетевой пропускной способности
	if stats.NetworkBandwidth > 0 {
		networkPercent := (stats.NetworkUsage * 100) / stats.NetworkBandwidth
		if networkPercent > networkUsageThreshold {
			freeNetworkBytes := stats.NetworkBandwidth - stats.NetworkUsage
			// Переводим байты в биты и затем в мегабиты
			freeNetworkMbit := (freeNetworkBytes * 8) / (1000 * 1000)
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeNetworkMbit)
		}
	}
}
