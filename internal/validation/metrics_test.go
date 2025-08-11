package validation

import (
	"strconv"
	"testing"

	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestValidateMetricRequest(t *testing.T) {
	tests := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue string
		wantErr     bool
		errType     string
	}{
		{
			name:        "Valid gauge metric",
			metricType:  "gauge",
			metricName:  "memory_usage",
			metricValue: "85.7",
			wantErr:     false,
		},
		{
			name:        "Valid counter metric",
			metricType:  "counter",
			metricName:  "request_count",
			metricValue: "123",
			wantErr:     false,
		},
		{
			name:        "Invalid metric type",
			metricType:  "unknown",
			metricName:  "test",
			metricValue: "123",
			wantErr:     true,
			errType:     "validation",
		},
		{
			name:        "Empty metric name",
			metricType:  "gauge",
			metricName:  "",
			metricValue: "123.45",
			wantErr:     true,
			errType:     "validation",
		},
		{
			name:        "Invalid gauge value",
			metricType:  "gauge",
			metricName:  "test",
			metricValue: "abc",
			wantErr:     true,
			errType:     "validation",
		},
		{
			name:        "Invalid counter value",
			metricType:  "counter",
			metricName:  "test",
			metricValue: "123.45",
			wantErr:     true,
			errType:     "validation",
		},
		{
			name:        "Empty value",
			metricType:  "gauge",
			metricName:  "test",
			metricValue: "",
			wantErr:     true,
			errType:     "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ValidateMetricRequest(tt.metricType, tt.metricName, tt.metricValue)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType == "validation" {
					assert.True(t, models.IsValidationError(err))

					// Проверяем, что это именно ValidationError
					var validationErr models.ValidationError
					assert.ErrorAs(t, err, &validationErr)

					// Проверяем, что сообщение об ошибке информативное
					assert.Contains(t, err.Error(), "validation error")
				}
				assert.Nil(t, req)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, tt.metricType, req.Type)
				assert.Equal(t, tt.metricName, req.Name)

				// Проверяем, что значение правильно распарсено
				switch tt.metricType {
				case "gauge":
					val, ok := req.Value.(float64)
					assert.True(t, ok, "Value should be float64 for gauge")
					expectedVal, _ := strconv.ParseFloat(tt.metricValue, 64)
					assert.Equal(t, expectedVal, val)
				case "counter":
					val, ok := req.Value.(int64)
					assert.True(t, ok, "Value should be int64 for counter")
					expectedVal, _ := strconv.ParseInt(tt.metricValue, 10, 64)
					assert.Equal(t, expectedVal, val)
				}
			}
		})
	}
}

func TestValidateMetricName(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		wantErr    bool
	}{
		{
			name:       "Valid metric name",
			metricName: "memory_usage",
			wantErr:    false,
		},
		{
			name:       "Empty metric name",
			metricName: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetricName(tt.metricName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, models.IsValidationError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMetricType(t *testing.T) {
	tests := []struct {
		name       string
		metricType string
		wantErr    bool
	}{
		{
			name:       "Valid gauge type",
			metricType: "gauge",
			wantErr:    false,
		},
		{
			name:       "Valid counter type",
			metricType: "counter",
			wantErr:    false,
		},
		{
			name:       "Invalid metric type",
			metricType: "unknown",
			wantErr:    true,
		},
		{
			name:       "Empty metric type",
			metricType: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetricType(tt.metricType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, models.IsValidationError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
