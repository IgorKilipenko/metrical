package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	serverURL := "http://localhost:8080"

	// Тестируем gauge метрики
	gaugeMetrics := map[string]float64{
		"temperature": 23.5,
		"humidity":    65.2,
		"pressure":    1013.25,
	}

	for name, value := range gaugeMetrics {
		url := fmt.Sprintf("%s/update/gauge/%s/%f", serverURL, name, value)
		resp, err := http.Post(url, "text/plain", nil)
		if err != nil {
			log.Printf("Error sending gauge metric %s: %v", name, err)
			continue
		}
		resp.Body.Close()
		log.Printf("Sent gauge metric %s = %f, status: %d", name, value, resp.StatusCode)
	}

	// Тестируем counter метрики
	counterMetrics := map[string]int64{
		"requests":    100,
		"errors":      5,
		"connections": 25,
	}

	for name, value := range counterMetrics {
		url := fmt.Sprintf("%s/update/counter/%s/%d", serverURL, name, value)
		resp, err := http.Post(url, "text/plain", nil)
		if err != nil {
			log.Printf("Error sending counter metric %s: %v", name, err)
			continue
		}
		resp.Body.Close()
		log.Printf("Sent counter metric %s = %d, status: %d", name, value, resp.StatusCode)
	}

	// Тестируем добавление к counter метрике
	url := fmt.Sprintf("%s/update/counter/requests/50", serverURL)
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		log.Printf("Error sending additional counter metric: %v", err)
	} else {
		resp.Body.Close()
		log.Printf("Sent additional counter metric requests = 50, status: %d", resp.StatusCode)
	}

	log.Println("Agent finished sending metrics")
}
