package template

import (
	"strings"
	"testing"

	models "github.com/IgorKilipenko/metrical/internal/model"
)

func TestNewMetricsTemplate(t *testing.T) {
	mt, err := NewMetricsTemplate()
	if err != nil {
		t.Fatalf("Failed to create metrics template: %v", err)
	}
	if mt == nil {
		t.Error("Template should not be nil")
	}
}

func TestMetricsTemplate_Execute(t *testing.T) {
	mt, err := NewMetricsTemplate()
	if err != nil {
		t.Fatalf("Failed to create metrics template: %v", err)
	}

	data := MetricsData{
		Gauges: models.GaugeMetrics{
			"temperature": 23.5,
			"memory":      1024.0,
		},
		Counters: models.CounterMetrics{
			"requests": 100,
			"errors":   5,
		},
		GaugeCount:   2,
		CounterCount: 2,
	}

	result, err := mt.Execute(data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := string(result)

	// Проверяем наличие ключевых элементов
	expectedElements := []string{
		"<title>Metrics Dashboard</title>",
		"<h1>Metrics Dashboard</h1>",
		"Gauge Metrics (2)",
		"Counter Metrics (2)",
		"temperature",
		"23.5",
		"requests",
		"100",
	}

	for _, element := range expectedElements {
		if !strings.Contains(html, element) {
			t.Errorf("Expected HTML to contain '%s', but it doesn't", element)
		}
	}
}

func TestMetricsTemplate_Execute_EmptyData(t *testing.T) {
	mt, err := NewMetricsTemplate()
	if err != nil {
		t.Fatalf("Failed to create metrics template: %v", err)
	}

	data := MetricsData{
		Gauges:       models.GaugeMetrics{},
		Counters:     models.CounterMetrics{},
		GaugeCount:   0,
		CounterCount: 0,
	}

	result, err := mt.Execute(data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := string(result)

	// Проверяем наличие сообщений об отсутствии метрик
	expectedElements := []string{
		"Gauge Metrics (0)",
		"Counter Metrics (0)",
		"No gauge metrics available",
		"No counter metrics available",
	}

	for _, element := range expectedElements {
		if !strings.Contains(html, element) {
			t.Errorf("Expected HTML to contain '%s', but it doesn't", element)
		}
	}
}
