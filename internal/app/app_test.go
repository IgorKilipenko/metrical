package app

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedAddr string
		expectedPort string
		expectError  bool
	}{
		{
			name:         "Default address",
			input:        "localhost:8080",
			expectedAddr: "localhost",
			expectedPort: "8080",
			expectError:  false,
		},
		{
			name:         "Custom address",
			input:        "localhost:9090",
			expectedAddr: "localhost",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "Only port",
			input:        "9090",
			expectedAddr: "localhost",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "Custom host and port",
			input:        "127.0.0.1:9090",
			expectedAddr: "127.0.0.1",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "Invalid address format",
			input:        "invalid:address:format",
			expectedAddr: "invalid",
			expectedPort: "address:format",
			expectError:  false,
		},
		{
			name:         "Empty string",
			input:        "",
			expectedAddr: "localhost",
			expectedPort: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if config.Addr != tt.expectedAddr {
					t.Errorf("Expected address %s, got %s", tt.expectedAddr, config.Addr)
				}
				if config.Port != tt.expectedPort {
					t.Errorf("Expected port %s, got %s", tt.expectedPort, config.Port)
				}
			}
		})
	}
}

func TestParseAddr(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedAddr string
		expectedPort string
		expectError  bool
	}{
		{
			name:         "Full address",
			input:        "localhost:8080",
			expectedAddr: "localhost",
			expectedPort: "8080",
			expectError:  false,
		},
		{
			name:         "Only port",
			input:        "9090",
			expectedAddr: "localhost",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "IP address",
			input:        "127.0.0.1:8080",
			expectedAddr: "127.0.0.1",
			expectedPort: "8080",
			expectError:  false,
		},
		{
			name:         "Empty string",
			input:        "",
			expectedAddr: "localhost",
			expectedPort: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, port, err := parseAddr(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if addr != tt.expectedAddr {
					t.Errorf("Expected address %s, got %s", tt.expectedAddr, addr)
				}
				if port != tt.expectedPort {
					t.Errorf("Expected port %s, got %s", tt.expectedPort, port)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	config := Config{Addr: "localhost", Port: "9090"}
	app := New(config)

	if app.GetPort() != "localhost:9090" {
		t.Errorf("New() addr = %s, want localhost:9090", app.GetPort())
	}

	if app.GetServer() != nil {
		t.Error("New() server should be nil before Run()")
	}
}

func TestApp_GetPort(t *testing.T) {
	config := Config{Addr: "localhost", Port: "8080"}
	app := New(config)

	addr := app.GetPort()
	if addr != "localhost:8080" {
		t.Errorf("GetPort() = %s, want localhost:8080", addr)
	}
}
