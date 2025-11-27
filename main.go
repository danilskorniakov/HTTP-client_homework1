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
	urlToPoll     = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval  = 1 * time.Second
	clientTimeout = 5 * time.Second
)

func main() {
	client := &http.Client{
		Timeout: clientTimeout,
	}

	consecutiveErrors := 0
	printedUnable := false

	for {
		err := pollOnce(client)
		if err != nil {
			consecutiveErrors++
			// When 3 or more consecutive errors - print the message once (until a successful fetch resets).
			if consecutiveErrors >= 3 && !printedUnable {
				fmt.Println("Unable to fetch server statistic.")
				printedUnable = true
			}
		} else {
			// success -> reset error counter and flag
			consecutiveErrors = 0
			printedUnable = false
		}

		time.Sleep(pollInterval)
	}
}

func pollOnce(client *http.Client) error {
	resp, err := client.Get(urlToPoll)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Expect HTTP 200
	if resp.StatusCode != http.StatusOK {
		// drain body to be polite
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	// Check Content-Type begins with text/plain (charset may be present)
	ct := resp.Header.Get("Content-Type")
	if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "text/plain") {
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("unexpected content-type: %s", ct)
	}

	// Read body (small)
	scanner := bufio.NewScanner(resp.Body)
	var body string
	if scanner.Scan() {
		body = scanner.Text()
		// If body spans multiple lines, join them (but expected single-line CSV)
		for scanner.Scan() {
			body += scanner.Text()
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return fmt.Errorf("empty body")
	}

	// Split by commas
	tokens := strings.Split(body, ",")
	if len(tokens) != 7 {
		return fmt.Errorf("unexpected token count: %d", len(tokens))
	}
	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}

	// Parse fields
	// 0: load average (may be float) - keep original string for printing
	loadStr := tokens[0]
	loadVal, err := strconv.ParseFloat(loadStr, 64)
	if err != nil {
		return fmt.Errorf("bad load value: %v", err)
	}

	// 1: total RAM bytes
	memTotal, err := strconv.ParseUint(tokens[1], 10, 64)
	if err != nil {
		return fmt.Errorf("bad mem total: %v", err)
	}
	// 2: used RAM bytes
	memUsed, err := strconv.ParseUint(tokens[2], 10, 64)
	if err != nil {
		return fmt.Errorf("bad mem used: %v", err)
	}
	// 3: disk total bytes
	diskTotal, err := strconv.ParseUint(tokens[3], 10, 64)
	if err != nil {
		return fmt.Errorf("bad disk total: %v", err)
	}
	// 4: disk used bytes
	diskUsed, err := strconv.ParseUint(tokens[4], 10, 64)
	if err != nil {
		return fmt.Errorf("bad disk used: %v", err)
	}
	// 5: network capacity bytes/sec
	netCap, err := strconv.ParseUint(tokens[5], 10, 64)
	if err != nil {
		return fmt.Errorf("bad net capacity: %v", err)
	}
	// 6: network usage bytes/sec
	netUsed, err := strconv.ParseUint(tokens[6], 10, 64)
	if err != nil {
		return fmt.Errorf("bad net used: %v", err)
	}

	// Threshold checks and messages

	// Load average > 30
	if loadVal > 30.0 {
		// print the original token so formatting remains identical to server output
		fmt.Printf("Load Average is too high: %s\n", loadStr)
	}

	// Memory usage > 80%
	if memTotal == 0 {
		return fmt.Errorf("mem total is zero")
	}
	memPercent := (float64(memUsed) / float64(memTotal)) * 100.0
	// round to nearest integer
	memPercentInt := int(memPercent + 0.5)
	if memPercent > 80.0 {
		fmt.Printf("Memory usage too high: %d%%\n", memPercentInt)
	}

	// Disk used > 90% -> print free MB left
	if diskTotal == 0 {
		return fmt.Errorf("disk total is zero")
	}
	diskUsageRatio := float64(diskUsed) / float64(diskTotal)
	if diskUsageRatio > 0.90 {
		var freeBytes uint64
		if diskTotal > diskUsed {
			freeBytes = diskTotal - diskUsed
		} else {
			freeBytes = 0
		}
		freeMB := freeBytes / (1024 * 1024)
		fmt.Printf("Free disk space is too low: %d Mb left\n", freeMB)
	}

	// Network usage > 90% of capacity -> print free Mbit/s available
	if netCap == 0 {
		return fmt.Errorf("net capacity is zero")
	}
	netUsageRatio := float64(netUsed) / float64(netCap)
	if netUsageRatio > 0.90 {
		var freeBytes uint64
		if netCap > netUsed {
			freeBytes = netCap - netUsed
		} else {
			freeBytes = 0
		}
		// convert bytes/sec to megabits/sec (Mb/s). 1 byte = 8 bits, 1 Mbit = 1_000_000 bits
		freeMbit := (float64(freeBytes) * 8.0) / 1000000.0
		// round to nearest integer
		freeMbitInt := int(freeMbit + 0.5)
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbitInt)
	}

	return nil
}
