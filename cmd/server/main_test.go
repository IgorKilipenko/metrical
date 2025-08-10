package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleError_NilError(t *testing.T) {
	// Тест с nil ошибкой - функция должна просто вернуться без паники
	// Используем defer для отлова паники
	panicked := false
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			t.Errorf("handleError(nil) вызвал панику: %v", r)
		}
	}()

	// Вызываем функцию с nil
	handleError(nil)

	// Проверяем, что паники не было
	assert.False(t, panicked, "handleError(nil) не должен вызывать панику")

	// Если мы дошли сюда, значит функция корректно обработала nil
	t.Log("handleError(nil) корректно обработал nil ошибку")
}

func TestHandleError_ErrorTypes(t *testing.T) {
	// Тестируем, что функции-предикаты корректно определяют типы ошибок
	testCases := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "HelpRequestedError",
			err:      HelpRequestedError{},
			expected: "help requested",
		},
		{
			name:     "InvalidAddressError",
			err:      InvalidAddressError{Address: "invalid", Reason: "test"},
			expected: "некорректный адрес 'invalid': test",
		},
		{
			name:     "RegularError",
			err:      &os.PathError{Op: "open", Path: "test.txt", Err: os.ErrNotExist},
			expected: "open test.txt: file does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Проверяем, что ошибка имеет ожидаемое сообщение
			assert.Equal(t, tc.expected, tc.err.Error())

			// Проверяем, что функции-предикаты работают корректно
			switch tc.name {
			case "HelpRequestedError":
				assert.True(t, IsHelpRequested(tc.err))
				assert.False(t, IsInvalidAddress(tc.err))
			case "InvalidAddressError":
				assert.False(t, IsHelpRequested(tc.err))
				assert.True(t, IsInvalidAddress(tc.err))
			case "RegularError":
				assert.False(t, IsHelpRequested(tc.err))
				assert.False(t, IsInvalidAddress(tc.err))
			}
		})
	}
}

func TestMainFunction_ErrorHandling(t *testing.T) {
	// Тестируем, что main функция корректно обрабатывает ошибки
	// Это интеграционный тест без фактического запуска сервера

	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	testCases := []struct {
		name        string
		args        []string
		expectError bool
		errorType   func(error) bool
	}{
		{
			name:        "InvalidAddress",
			args:        []string{"server", "-a", "invalid:address:format"},
			expectError: true,
			errorType:   IsInvalidAddress,
		},
		{
			name:        "ValidAddress",
			args:        []string{"server", "-a", "localhost:8080"},
			expectError: false,
		},
		{
			name:        "HelpFlag",
			args:        []string{"server", "--help"},
			expectError: true,
			errorType:   IsHelpRequested,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args

			// Проверяем, что parseFlags возвращает ожидаемый результат
			_, err := parseFlags()

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorType != nil {
					assert.True(t, tc.errorType(err), "Expected specific error type for %s", tc.name)
				}
			} else {
				assert.NoError(t, err, "Expected no error for %s", tc.name)
			}
		})
	}
}

func TestHandleError_WithExitCodes(t *testing.T) {
	// Тестируем, что handleError вызывает os.Exit с правильными кодами
	// Используем отдельные процессы для тестирования os.Exit

	testCases := []struct {
		name         string
		err          error
		expectedCode int
		envVar       string
	}{
		{
			name:         "InvalidAddressError",
			err:          InvalidAddressError{Address: "invalid", Reason: "test"},
			expectedCode: 1,
			envVar:       "TEST_INVALID_ADDRESS_EXIT",
		},
		{
			name:         "RegularError",
			err:          &os.PathError{Op: "open", Path: "test.txt", Err: os.ErrNotExist},
			expectedCode: 1,
			envVar:       "TEST_REGULAR_ERROR_EXIT",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Проверяем, что процесс завершается с ожидаемым кодом
			if os.Getenv(tc.envVar) == "1" {
				handleError(tc.err)
				return
			}

			cmd := exec.Command(os.Args[0], "-test.run=TestHandleError_WithExitCodes")
			cmd.Env = append(os.Environ(), tc.envVar+"=1")
			err := cmd.Run()

			if exitError, ok := err.(*exec.ExitError); ok {
				assert.Equal(t, tc.expectedCode, exitError.ExitCode(),
					"Process should exit with code %d for %s", tc.expectedCode, tc.name)
			} else {
				t.Errorf("Expected exit error for %s, got %v", tc.name, err)
			}
		})
	}
}

func TestHandleError_LogicFlow(t *testing.T) {
	// Тестируем логику handleError для nil ошибки
	// Это единственный случай, который можно безопасно тестировать

	// Сохраняем оригинальную функцию os.Exit
	originalExit := osExit
	defer func() { osExit = originalExit }()

	exitCodes := []int{}
	osExit = func(code int) {
		exitCodes = append(exitCodes, code)
	}

	// Тестируем только nil ошибку
	handleError(nil)
	assert.Empty(t, exitCodes, "handleError(nil) should not call os.Exit")
}

func TestHandleError_EdgeCases(t *testing.T) {
	// Тестируем граничные случаи для handleError
	// Тестируем только случаи, которые не вызывают log.Fatal

	// Сохраняем оригинальную функцию os.Exit
	originalExit := osExit
	defer func() { osExit = originalExit }()

	exitCodes := []int{}
	osExit = func(code int) {
		exitCodes = append(exitCodes, code)
	}

	testCases := []struct {
		name         string
		err          error
		expectedCode int
		description  string
	}{
		{
			name:         "EmptyInvalidAddressError",
			err:          InvalidAddressError{Address: "", Reason: ""},
			expectedCode: 1,
			description:  "Empty fields in InvalidAddressError",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Очищаем предыдущие вызовы
			exitCodes = exitCodes[:0]

			// Вызываем handleError
			handleError(tc.err)

			// Проверяем результат
			assert.Len(t, exitCodes, 1, "handleError should call os.Exit exactly once for %s", tc.description)
			assert.Equal(t, tc.expectedCode, exitCodes[0],
				"handleError should exit with code %d for %s", tc.expectedCode, tc.description)
		})
	}
}

// customError - кастомный тип ошибки для тестирования
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}
