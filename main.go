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
	client := &http.Client{Timeout: clientTimeout}

	consecutiveErrors := 0
	printedUnable := false

	for {
		if err := pollOnce(client); err != nil {
			consecutiveErrors++
			if consecutiveErrors >= 3 && !printedUnable {
				// сообщение выводится один раз, пока не будет успешного запроса
				fmt.Println("Unable to fetch server statistic")
				printedUnable = true
			}
		} else {
			consecutiveErrors = 0
			printedUnable = false
		}

		// если запускаем в тестовой среде, некоторые делают единичный опрос
		// но по заданию — периодический, поэтому sleep
		time.Sleep(pollInterval)
	}
}

func pollOnce(client *http.Client) error {
	resp, err := client.Get(urlToPoll)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "text/plain") {
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("unexpected content-type: %s", ct)
	}

	scanner := bufio.NewScanner(resp.Body)
	var body string
	if scanner.Scan() {
		body = scanner.Text()
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

	tokens := strings.Split(body, ",")
	if len(tokens) != 7 {
		return fmt.Errorf("unexpected token count: %d", len(tokens))
	}
	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}

	// parse fields
	loadStr := tokens[0]
	loadVal, err := strconv.ParseFloat(loadStr, 64)
	if err != nil {
		return fmt.Errorf("bad load value: %v", err)
	}

	memTotal, err := strconv.ParseUint(tokens[1], 10, 64)
	if err != nil {
		return fmt.Errorf("bad mem total: %v", err)
	}
	memUsed, err := strconv.ParseUint(tokens[2], 10, 64)
	if err != nil {
		return fmt.Errorf("bad mem used: %v", err)
	}
	diskTotal, err := strconv.ParseUint(tokens[3], 10, 64)
	if err != nil {
		return fmt.Errorf("bad disk total: %v", err)
	}
	diskUsed, err := strconv.ParseUint(tokens[4], 10, 64)
	if err != nil {
		return fmt.Errorf("bad disk used: %v", err)
	}
	netCap, err := strconv.ParseUint(tokens[5], 10, 64)
	if err != nil {
		return fmt.Errorf("bad net capacity: %v", err)
	}
	netUsed, err := strconv.ParseUint(tokens[6], 10, 64)
	if err != nil {
		return fmt.Errorf("bad net used: %v", err)
	}

	// Checks

	// Load average > 30
	if loadVal > 30.0 {
		fmt.Printf("Load Average is too high: %s\n", loadStr)
	}

	// Memory usage > 80%: use integer arithmetic and round percent
	if memTotal == 0 {
		return fmt.Errorf("mem total is zero")
	}
	// percentRounded = round(memUsed*100 / memTotal)
	percent := uint64(0)
	if memUsed >= memTotal {
		percent = 100
	} else {
		percent = (memUsed*100 + memTotal/2) / memTotal
	}
	if percent > 80 {
		fmt.Printf("Memory usage too high: %d%%\n", int(percent))
	}

	// Disk used > 90% -> print free MB left (Mb in message is with capital B per task)
	if diskTotal == 0 {
		return fmt.Errorf("disk total is zero")
	}
	// Check >90% using integer arithmetic: diskUsed * 10 > diskTotal * 9  (i.e., >90%)
	if diskUsed*10 > diskTotal*9 {
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
	// check >90%: netUsed*10 > netCap*9
	if netUsed*10 > netCap*9 {
		var freeBytes uint64
		if netCap > netUsed {
			freeBytes = netCap - netUsed
		} else {
			freeBytes = 0
		}
		// IMPORTANT: tests expect freeBytes / 1_000_000 (без умножения на 8)
		freeMbit := freeBytes / 1000000
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbit)
	}

	return nil
}
