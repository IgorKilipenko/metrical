# Makefile для проекта go-metrics
# Автор: Igor Kilipenko
# Описание: Автоматизация сборки, тестирования и запуска автотестов

# Переменные
AUTO_TEST_BINARY := ../auto-tests/metricstest_v2
SERVER_BINARY := bin/server
AGENT_BINARY := bin/agent
SERVER_PORT := 9091
DATABASE_DSN := 'postgres://metricaldb:Secret@localhost:5432/metricaldb?sslmode=disable'
FILE_STORAGE_PATH := /tmp/iteration9-metrics.json

# Цвета для вывода
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Флаги для автотестов
AUTO_TEST_FLAGS := -test.v -binary-path=$(SERVER_BINARY) -agent-binary-path=$(AGENT_BINARY) -source-path=. -server-port=$(SERVER_PORT)

.PHONY: help build clean test auto-test auto-test-iteration auto-test-all check-deps

# Помощь
help: ## Показать справку по командам
	@echo "$(BLUE)Доступные команды:$(NC)"
	@echo ""
	@echo "$(GREEN)Сборка:$(NC)"
	@echo "  build          - Собрать агент и сервер"
	@echo "  build-agent    - Собрать только агент"
	@echo "  build-server   - Собрать только сервер"
	@echo "  clean          - Очистить кэш сборки и бинарники"
	@echo ""
	@echo "$(GREEN)Тестирование:$(NC)"
	@echo "  test           - Запустить все unit тесты"
	@echo "  test-coverage  - Запустить тесты с покрытием"
	@echo "  auto-test-1    - Запустить автотесты итерации 1"
	@echo "  auto-test-2    - Запустить автотесты итерации 2"
	@echo "  auto-test-3    - Запустить автотесты итерации 3"
	@echo "  auto-test-4    - Запустить автотесты итерации 4"
	@echo "  auto-test-5    - Запустить автотесты итерации 5"
	@echo "  auto-test-6    - Запустить автотесты итерации 6"
	@echo "  auto-test-7    - Запустить автотесты итерации 7"
	@echo "  auto-test-8    - Запустить автотесты итерации 8"
	@echo "  auto-test-9    - Запустить автотесты итерации 9"
	@echo "  auto-test-10   - Запустить автотесты итерации 10"
	@echo "  auto-test-11   - Запустить автотесты итерации 11"
	@echo "  auto-test-all  - Запустить все автотесты (1-11)"
	@echo "  full-test      - Запустить unit тесты + автотесты"
	@echo ""
	@echo "$(GREEN)Запуск:$(NC)"
	@echo "  run-server     - Запустить сервер"
	@echo "  run-agent      - Запустить агент"
	@echo ""
	@echo "$(GREEN)Утилиты:$(NC)"
	@echo "  check-deps     - Проверить зависимости"
	@echo "  lint           - Запустить линтеры"

# Сборка
build: build-agent build-server ## Собрать агент и сервер

build-agent: ## Собрать агент
	@echo "$(BLUE)Сборка агента...$(NC)"
	@go build -o $(AGENT_BINARY) ./cmd/agent
	@echo "$(GREEN)Агент собран: $(AGENT_BINARY)$(NC)"

build-server: ## Собрать сервер
	@echo "$(BLUE)Сборка сервера...$(NC)"
	@go build -o $(SERVER_BINARY) ./cmd/server
	@echo "$(GREEN)Сервер собран: $(SERVER_BINARY)$(NC)"

# Очистка
clean: ## Очистить кэш сборки и бинарники
	@echo "$(BLUE)Очистка...$(NC)"
	@go clean
	@rm -f $(SERVER_BINARY) $(AGENT_BINARY)
	@echo "$(GREEN)Очистка завершена$(NC)"

# Unit тесты
test: ## Запустить все unit тесты
	@echo "$(BLUE)Запуск unit тестов...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)Unit тесты завершены$(NC)"

test-coverage: ## Запустить тесты с покрытием
	@echo "$(BLUE)Запуск тестов с покрытием...$(NC)"
	@go test -v -cover ./...
	@echo "$(GREEN)Тесты с покрытием завершены$(NC)"

# Тесты с базой данных
test-db: test-db-up test-db-run test-db-down ## Запустить тесты с PostgreSQL

test-db-up: ## Запустить тестовую БД
	@echo "$(BLUE)Запуск тестовой PostgreSQL...$(NC)"
	@docker-compose -f docker-compose.test.yml up -d
	@echo "$(YELLOW)Ожидание готовности БД...$(NC)"
	@for i in $$(seq 1 30); do \
		if docker-compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U test -d testdb >/dev/null 2>&1; then \
			echo "$(GREEN)Тестовая БД готова$(NC)"; \
			exit 0; \
		fi; \
		echo "Ожидание... ($$i/30)"; \
		sleep 1; \
	done; \
	echo "$(RED)БД не готова через 30 секунд$(NC)"; \
	exit 1

test-db-run: ## Запустить тесты с БД
	@echo "$(BLUE)Запуск тестов с PostgreSQL...$(NC)"
	@TEST_DATABASE_URL="postgres://test:test@localhost:5433/testdb?sslmode=disable" go test -v ./internal/repository/ -run TestPostgreSQL
	@echo "$(GREEN)Тесты с БД завершены$(NC)"

test-db-down: ## Остановить тестовую БД
	@echo "$(BLUE)Остановка тестовой PostgreSQL...$(NC)"
	@docker-compose -f docker-compose.test.yml down -v
	@echo "$(GREEN)Тестовая БД остановлена$(NC)"

test-integration: test-db ## Алиас для интеграционных тестов

# Проверка зависимостей
check-deps: ## Проверить зависимости
	@echo "$(BLUE)Проверка зависимостей...$(NC)"
	@if [ ! -f "$(AUTO_TEST_BINARY)" ]; then \
		echo "$(RED)ОШИБКА: Автотесты не найдены в $(AUTO_TEST_BINARY)$(NC)"; \
		echo "$(YELLOW)Убедитесь, что автотесты установлены в правильной директории$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Зависимости проверены$(NC)"

# Линтеры
lint: ## Запустить линтеры
	@echo "$(BLUE)Запуск линтеров...$(NC)"
	@go vet ./...
	@echo "$(GREEN)Линтеры завершены$(NC)"

# Автотесты - отдельные итерации
auto-test-1: build check-deps ## Запустить автотесты итерации 1
	@echo "$(BLUE)Запуск автотестов итерации 1...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration1$$ || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 1 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 1 прошли успешно!$(NC)"

auto-test-2: build check-deps ## Запустить автотесты итерации 2
	@echo "$(BLUE)Запуск автотестов итерации 2...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration2[AB]$$ || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 2 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 2 прошли успешно!$(NC)"

auto-test-3: build check-deps ## Запустить автотесты итерации 3
	@echo "$(BLUE)Запуск автотестов итерации 3...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration3[AB]$$ || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 3 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 3 прошли успешно!$(NC)"

auto-test-4: build check-deps ## Запустить автотесты итерации 4
	@echo "$(BLUE)Запуск автотестов итерации 4...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration4$$ || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 4 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 4 прошли успешно!$(NC)"

auto-test-5: build check-deps ## Запустить автотесты итерации 5
	@echo "$(BLUE)Запуск автотестов итерации 5...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration5$$ || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 5 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 5 прошли успешно!$(NC)"

auto-test-6: build check-deps ## Запустить автотесты итерации 6
	@echo "$(BLUE)Запуск автотестов итерации 6...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration6$$ || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 6 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 6 прошли успешно!$(NC)"

auto-test-7: build check-deps ## Запустить автотесты итерации 7
	@echo "$(BLUE)Запуск автотестов итерации 7...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration7$$ || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 7 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 7 прошли успешно!$(NC)"

auto-test-8: build check-deps ## Запустить автотесты итерации 8
	@echo "$(BLUE)Запуск автотестов итерации 8...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration8$$ || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 8 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 8 прошли успешно!$(NC)"

auto-test-9: build check-deps ## Запустить автотесты итерации 9
	@echo "$(BLUE)Запуск автотестов итерации 9...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration9$$ -file-storage-path=$(FILE_STORAGE_PATH) || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 9 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 9 прошли успешно!$(NC)"

auto-test-10: build check-deps ## Запустить автотесты итерации 10
	@echo "$(BLUE)Запуск автотестов итерации 10...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration10[AB]$$ -database-dsn=$(DATABASE_DSN) || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 10 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 10 прошли успешно!$(NC)"

auto-test-11: build check-deps ## Запустить автотесты итерации 11
	@echo "$(BLUE)Запуск автотестов итерации 11...$(NC)"
	@$(AUTO_TEST_BINARY) $(AUTO_TEST_FLAGS) -test.run=^TestIteration11$$ -database-dsn=$(DATABASE_DSN) || (echo "$(RED)АВТОТЕСТЫ ИТЕРАЦИИ 11 НЕ ПРОШЛИ!$(NC)" && exit 1)
	@echo "$(GREEN)Автотесты итерации 11 прошли успешно!$(NC)"

# Все автотесты
auto-test-all: build check-deps ## Запустить все автотесты (1-11)
	@echo "$(BLUE)Запуск всех автотестов (итерации 1-11)...$(NC)"
	@echo "$(YELLOW)Это может занять несколько минут...$(NC)"
	@echo ""
	@$(MAKE) auto-test-1 || exit 1
	@$(MAKE) auto-test-2 || exit 1
	@$(MAKE) auto-test-3 || exit 1
	@$(MAKE) auto-test-4 || exit 1
	@$(MAKE) auto-test-5 || exit 1
	@$(MAKE) auto-test-6 || exit 1
	@$(MAKE) auto-test-7 || exit 1
	@$(MAKE) auto-test-8 || exit 1
	@$(MAKE) auto-test-9 || exit 1
	@$(MAKE) auto-test-10 || exit 1
	@$(MAKE) auto-test-11 || exit 1
	@echo ""
	@echo "$(GREEN)🎉 ВСЕ АВТОТЕСТЫ ПРОШЛИ УСПЕШНО! 🎉$(NC)"

# Полный набор тестов
full-test: test auto-test-all ## Запустить unit тесты + автотесты
	@echo "$(GREEN)🎉 ВСЕ ТЕСТЫ ПРОШЛИ УСПЕШНО! 🎉$(NC)"

# Запуск приложений
run-server: build-server ## Запустить сервер
	@echo "$(BLUE)Запуск сервера на localhost:9090...$(NC)"
	@$(SERVER_BINARY) -a=localhost:9090

run-agent: build-agent ## Запустить агент
	@echo "$(BLUE)Запуск агента на localhost:8080...$(NC)"
	@$(AGENT_BINARY) -a=localhost:8080 -r=2s

# Быстрые команды для разработки
quick-test: ## Быстрый тест (только unit тесты)
	@$(MAKE) test

quick-build: ## Быстрая сборка
	@$(MAKE) build

# Проверка конкретной итерации (для отладки)
check-iteration: ## Проверить конкретную итерацию (использование: make check-iteration ITERATION=2)
	@if [ -z "$(ITERATION)" ]; then \
		echo "$(RED)ОШИБКА: Укажите итерацию: make check-iteration ITERATION=2$(NC)"; \
		exit 1; \
	fi
	@$(MAKE) auto-test-$(ITERATION)

# Статистика
stats: ## Показать статистику проекта
	@echo "$(BLUE)Статистика проекта:$(NC)"
	@echo "  Go файлов: $$(find . -name '*.go' | wc -l)"
	@echo "  Тестов: $$(find . -name '*_test.go' | wc -l)"
	@echo "  Строк кода: $$(find . -name '*.go' -exec wc -l {} + | tail -1 | awk '{print $$1}')"
	@echo "  Размер бинарников:"
	@if [ -f "$(SERVER_BINARY)" ]; then echo "    Сервер: $$(du -h $(SERVER_BINARY) | cut -f1)"; fi
	@if [ -f "$(AGENT_BINARY)" ]; then echo "    Агент: $$(du -h $(AGENT_BINARY) | cut -f1)"; fi
