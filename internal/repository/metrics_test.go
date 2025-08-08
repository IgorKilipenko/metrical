package repository

import (
	"testing"

	models "github.com/IgorKilipenko/metrical/internal/model"
)

func TestInMemoryMetricsRepository_UpdateGauge(t *testing.T) {
	storage := models.NewMemStorage()
	repo := NewInMemoryMetricsRepository(storage)

	err := repo.UpdateGauge("temperature", 23.5)
	if err != nil {
		t.Errorf("UpdateGauge failed: %v", err)
	}

	value, exists, err := repo.GetGauge("temperature")
	if err != nil {
		t.Errorf("GetGauge failed: %v", err)
	}
	if !exists {
		t.Error("Gauge should exist")
	}
	if value != 23.5 {
		t.Errorf("Expected 23.5, got %f", value)
	}
}

func TestInMemoryMetricsRepository_UpdateCounter(t *testing.T) {
	storage := models.NewMemStorage()
	repo := NewInMemoryMetricsRepository(storage)

	// Добавляем значение
	err := repo.UpdateCounter("requests", 100)
	if err != nil {
		t.Errorf("UpdateCounter failed: %v", err)
	}

	// Добавляем еще значение (должно накопиться)
	err = repo.UpdateCounter("requests", 50)
	if err != nil {
		t.Errorf("UpdateCounter failed: %v", err)
	}

	value, exists, err := repo.GetCounter("requests")
	if err != nil {
		t.Errorf("GetCounter failed: %v", err)
	}
	if !exists {
		t.Error("Counter should exist")
	}
	if value != 150 {
		t.Errorf("Expected 150, got %d", value)
	}
}

func TestInMemoryMetricsRepository_GetGauge_NotExists(t *testing.T) {
	storage := models.NewMemStorage()
	repo := NewInMemoryMetricsRepository(storage)

	value, exists, err := repo.GetGauge("nonexistent")
	if err != nil {
		t.Errorf("GetGauge failed: %v", err)
	}
	if exists {
		t.Error("Gauge should not exist")
	}
	if value != 0 {
		t.Errorf("Expected 0, got %f", value)
	}
}

func TestInMemoryMetricsRepository_GetCounter_NotExists(t *testing.T) {
	storage := models.NewMemStorage()
	repo := NewInMemoryMetricsRepository(storage)

	value, exists, err := repo.GetCounter("nonexistent")
	if err != nil {
		t.Errorf("GetCounter failed: %v", err)
	}
	if exists {
		t.Error("Counter should not exist")
	}
	if value != 0 {
		t.Errorf("Expected 0, got %d", value)
	}
}

func TestInMemoryMetricsRepository_GetAllGauges(t *testing.T) {
	storage := models.NewMemStorage()
	repo := NewInMemoryMetricsRepository(storage)

	// Добавляем несколько gauge метрик
	repo.UpdateGauge("temp1", 10.5)
	repo.UpdateGauge("temp2", 20.7)

	gauges, err := repo.GetAllGauges()
	if err != nil {
		t.Errorf("GetAllGauges failed: %v", err)
	}

	if len(gauges) != 2 {
		t.Errorf("Expected 2 gauges, got %d", len(gauges))
	}

	if gauges["temp1"] != 10.5 {
		t.Errorf("Expected temp1=10.5, got %f", gauges["temp1"])
	}
	if gauges["temp2"] != 20.7 {
		t.Errorf("Expected temp2=20.7, got %f", gauges["temp2"])
	}
}

func TestInMemoryMetricsRepository_GetAllCounters(t *testing.T) {
	storage := models.NewMemStorage()
	repo := NewInMemoryMetricsRepository(storage)

	// Добавляем несколько counter метрик
	repo.UpdateCounter("req1", 100)
	repo.UpdateCounter("req2", 200)

	counters, err := repo.GetAllCounters()
	if err != nil {
		t.Errorf("GetAllCounters failed: %v", err)
	}

	if len(counters) != 2 {
		t.Errorf("Expected 2 counters, got %d", len(counters))
	}

	if counters["req1"] != 100 {
		t.Errorf("Expected req1=100, got %d", counters["req1"])
	}
	if counters["req2"] != 200 {
		t.Errorf("Expected req2=200, got %d", counters["req2"])
	}
}
