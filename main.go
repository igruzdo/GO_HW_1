package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL               = "http://srv.msk01.gigacorp.local/_stats"
	maxLoadAverage          = 30.0
	maxMemoryUsagePercent   = 80.0
	maxDiskUsagePercent     = 90.0
	maxNetworkUsagePercent  = 90.0
	checkInterval           = 10 * time.Second
	maxErrorCount           = 3
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get(serverURL)
		if err != nil {
			errorCount++
			handleFetchError(errorCount)
			time.Sleep(checkInterval)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			errorCount++
			handleFetchError(errorCount)
			resp.Body.Close()
			time.Sleep(checkInterval)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			handleFetchError(errorCount)
			time.Sleep(checkInterval)
			continue
		}

		values, parseErr := parseResponse(string(body))
		if parseErr != nil {
			errorCount++
			handleFetchError(errorCount)
			time.Sleep(checkInterval)
			continue
		}

		errorCount = 0 // сбрасываем счётчик ошибок при успешном получении данных
		checkStats(values)
		time.Sleep(checkInterval)
	}
}

// handleFetchError выводит сообщение об ошибке, если их было 3 или больше
func handleFetchError(errorCount int) {
	if errorCount >= maxErrorCount {
		fmt.Println("Unable to fetch server statistic.")
	}
}

// parseResponse парсит строку ответа сервера и возвращает массив чисел
func parseResponse(data string) ([]float64, error) {
	fields := strings.Split(strings.TrimSpace(data), ",")
	if len(fields) != 7 {
		return nil, fmt.Errorf("invalid response format")
	}

	values := make([]float64, len(fields))
	for i, field := range fields {
		val, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number format in response")
		}
		values[i] = val
	}
	return values, nil
}

// checkStats проверяет параметры и выводит предупреждения, если они превышают заданные пороги
func checkStats(values []float64) {
	loadAverage := values[0]
	totalMemory := values[1]
	usedMemory := values[2]
	totalDisk := values[3]
	usedDisk := values[4]
	totalBandwidth := values[5]
	usedBandwidth := values[6]

	// Проверка Load Average
	if loadAverage > maxLoadAverage {
		fmt.Printf("Load Average is too high: %.2f\n", loadAverage)
	}

	// Проверка использования памяти
	if usedMemory/totalMemory*100 > maxMemoryUsagePercent {
		memUsagePercent := usedMemory / totalMemory * 100
		fmt.Printf("Memory usage too high: %.2f%%\n", memUsagePercent)
	}

	// Проверка использования диска
	freeDiskMB := (totalDisk - usedDisk) / (1000 * 1000)
	if usedDisk/totalDisk*100 > maxDiskUsagePercent {
		fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeDiskMB)
	}

	// Проверка использования пропускной способности сети
	freeBandwidthMbit := (totalBandwidth - usedBandwidth) * 10 / (1000 * 1000)
	if usedBandwidth/totalBandwidth*100 > maxNetworkUsagePercent {
		fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", freeBandwidthMbit)
	}
}