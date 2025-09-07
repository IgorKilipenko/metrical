package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"testing"

	models "github.com/IgorKilipenko/metrical/internal/model"
)

func TestAgent_CompressData(t *testing.T) {
	agent := &Agent{}

	// Тест 1: Сжатие JSON данных
	t.Run("compress json data", func(t *testing.T) {
		testData := []byte(`{"test": "data", "number": 42}`)

		compressed, err := agent.compressData(testData)
		if err != nil {
			t.Fatalf("Failed to compress data: %v", err)
		}

		// Проверяем, что сжатые данные можно распаковать
		gzReader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()

		uncompressed, err := io.ReadAll(gzReader)
		if err != nil {
			t.Fatalf("Failed to decompress data: %v", err)
		}

		if string(uncompressed) != string(testData) {
			t.Errorf("Expected %s, got %s", string(testData), string(uncompressed))
		}
	})

	// Тест 2: Сжатие пустых данных
	t.Run("compress empty data", func(t *testing.T) {
		testData := []byte("")

		compressed, err := agent.compressData(testData)
		if err != nil {
			t.Fatalf("Failed to compress empty data: %v", err)
		}

		// Пустые данные должны быть сжаты (хотя размер может не уменьшиться)
		if len(compressed) == 0 {
			t.Error("Compressed data should not be empty")
		}
	})

	// Тест 3: Сжатие больших данных
	t.Run("compress large data", func(t *testing.T) {
		// Создаем большие тестовые данные
		testData := bytes.Repeat([]byte("This is a test string that will be repeated many times. "), 1000)

		compressed, err := agent.compressData(testData)
		if err != nil {
			t.Fatalf("Failed to compress large data: %v", err)
		}

		// Для повторяющихся данных сжатие должно быть эффективным
		compressionRatio := float64(len(compressed)) / float64(len(testData))
		if compressionRatio > 0.8 {
			t.Errorf("Expected better compression ratio, got %.2f", compressionRatio)
		}
	})
}

func TestAgent_PrepareMetricJSON(t *testing.T) {
	agent := &Agent{}

	t.Run("prepare gauge metric", func(t *testing.T) {
		name := "test_gauge"
		value := float64(42.5)

		metric, err := agent.prepareMetricJSON(name, value)
		if err != nil {
			t.Fatalf("Failed to prepare gauge metric: %v", err)
		}

		if metric.ID != name {
			t.Errorf("Expected metric ID %s, got %s", name, metric.ID)
		}

		if metric.MType != "gauge" {
			t.Errorf("Expected metric type 'gauge', got %s", metric.MType)
		}

		if metric.Value == nil || *metric.Value != value {
			t.Errorf("Expected metric value %f, got %v", value, metric.Value)
		}

		if metric.Delta != nil {
			t.Errorf("Expected delta to be nil for gauge, got %v", metric.Delta)
		}
	})

	t.Run("prepare counter metric", func(t *testing.T) {
		name := "test_counter"
		value := int64(100)

		metric, err := agent.prepareMetricJSON(name, value)
		if err != nil {
			t.Fatalf("Failed to prepare counter metric: %v", err)
		}

		if metric.ID != name {
			t.Errorf("Expected metric ID %s, got %s", name, metric.ID)
		}

		if metric.MType != "counter" {
			t.Errorf("Expected metric type 'counter', got %s", metric.MType)
		}

		if metric.Delta == nil || *metric.Delta != value {
			t.Errorf("Expected metric delta %d, got %v", value, metric.Delta)
		}

		if metric.Value != nil {
			t.Errorf("Expected value to be nil for counter, got %v", metric.Value)
		}
	})

	t.Run("prepare metric with unknown type", func(t *testing.T) {
		name := "test_unknown"
		value := "string_value"

		_, err := agent.prepareMetricJSON(name, value)
		if err == nil {
			t.Error("Expected error for unknown metric type")
		}

		expectedError := "unknown metric type for test_unknown: string"
		if err.Error() != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestAgent_CompressDataIntegration(t *testing.T) {
	agent := &Agent{}

	// Тест интеграции: подготовка метрики -> JSON -> сжатие
	t.Run("metric to json to compression", func(t *testing.T) {
		// Подготавливаем метрику
		metric, err := agent.prepareMetricJSON("test_metric", float64(123.45))
		if err != nil {
			t.Fatalf("Failed to prepare metric: %v", err)
		}

		// Кодируем в JSON
		jsonData, err := json.Marshal(metric)
		if err != nil {
			t.Fatalf("Failed to marshal metric: %v", err)
		}

		// Сжимаем JSON
		compressed, err := agent.compressData(jsonData)
		if err != nil {
			t.Fatalf("Failed to compress JSON: %v", err)
		}

		// Проверяем, что сжатие работает (для маленьких данных размер может не уменьшиться)
		// Главное - что данные можно корректно распаковать

		// Проверяем, что можно распаковать и получить исходные данные
		gzReader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()

		uncompressed, err := io.ReadAll(gzReader)
		if err != nil {
			t.Fatalf("Failed to decompress: %v", err)
		}

		if string(uncompressed) != string(jsonData) {
			t.Errorf("Expected %s, got %s", string(jsonData), string(uncompressed))
		}

		// Проверяем, что распакованные данные можно распарсить обратно в метрику
		var restoredMetric models.Metrics
		if err := json.Unmarshal(uncompressed, &restoredMetric); err != nil {
			t.Fatalf("Failed to unmarshal restored metric: %v", err)
		}

		if restoredMetric.ID != metric.ID {
			t.Errorf("Expected ID %s, got %s", metric.ID, restoredMetric.ID)
		}

		if restoredMetric.MType != metric.MType {
			t.Errorf("Expected type %s, got %s", metric.MType, restoredMetric.MType)
		}
	})
}
