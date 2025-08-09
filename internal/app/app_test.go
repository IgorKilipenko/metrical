package app

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedAddr string
		expectedPort string
		expectError  bool
	}{
		{
			name:         "Default address",
			args:         []string{},
			expectedAddr: "localhost",
			expectedPort: "8080",
			expectError:  false,
		},
		{
			name:         "Custom address",
			args:         []string{"--address=localhost:9090"},
			expectedAddr: "localhost",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "Only port",
			args:         []string{"--address=9090"},
			expectedAddr: "localhost",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "Custom host and port",
			args:         []string{"--address=127.0.0.1:9090"},
			expectedAddr: "127.0.0.1",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "Unknown argument",
			args:         []string{"unknown"},
			expectedAddr: "",
			expectedPort: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый Cobra команду для каждого теста
			cmd := &cobra.Command{
				Use:   "test",
				Short: "Test command",
				RunE: func(cmd *cobra.Command, args []string) error {
					if len(args) > 0 {
						return fmt.Errorf("неизвестные аргументы: %v", args)
					}
					return nil
				},
			}

			var addr string
			cmd.Flags().StringVarP(&addr, "address", "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")

			// Устанавливаем аргументы для теста
			cmd.SetArgs(tt.args)

			// Выполняем команду
			err := cmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for unknown arguments, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Парсим адрес и порт
				serverAddr, serverPort, err := parseAddr(addr)
				if err != nil {
					t.Errorf("Failed to parse address: %v", err)
				}

				if serverAddr != tt.expectedAddr {
					t.Errorf("Expected address %s, got %s", tt.expectedAddr, serverAddr)
				}
				if serverPort != tt.expectedPort {
					t.Errorf("Expected port %s, got %s", tt.expectedPort, serverPort)
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
