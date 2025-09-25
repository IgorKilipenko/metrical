#!/bin/bash

# Скрипт для запуска тестов с PostgreSQL
# Автор: Igor Kilipenko

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🐳 Запуск тестов с PostgreSQL...${NC}"

# Проверяем, установлен ли Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker не установлен!${NC}"
    echo -e "${YELLOW}Установите Docker: https://docs.docker.com/get-docker/${NC}"
    exit 1
fi

# Проверяем, установлен ли docker-compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ docker-compose не установлен!${NC}"
    echo -e "${YELLOW}Установите docker-compose: https://docs.docker.com/compose/install/${NC}"
    exit 1
fi

# Функция для очистки при выходе
cleanup() {
    echo -e "${YELLOW}🧹 Очистка...${NC}"
    docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
}

# Устанавливаем trap для очистки при выходе
trap cleanup EXIT

echo -e "${BLUE}📦 Запуск тестовой PostgreSQL...${NC}"
docker-compose -f docker-compose.test.yml up -d

echo -e "${YELLOW}⏳ Ожидание готовности БД...${NC}"
# Ждем, пока БД станет доступной
for i in {1..30}; do
    if docker-compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U test -d testdb >/dev/null 2>&1; then
        echo -e "${GREEN}✅ БД готова!${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ БД не готова через 30 секунд${NC}"
        exit 1
    fi
    sleep 1
done

echo -e "${BLUE}🧪 Запуск тестов...${NC}"
export TEST_DATABASE_URL="postgres://test:test@localhost:5433/testdb?sslmode=disable"

# Запускаем тесты
if go test -v ./internal/repository/ -run TestPostgreSQL; then
    echo -e "${GREEN}🎉 Все тесты прошли успешно!${NC}"
else
    echo -e "${RED}❌ Тесты не прошли${NC}"
    exit 1
fi

echo -e "${GREEN}✨ Тестирование завершено!${NC}"
