#!/bin/bash

# Скрипт для инициализации основной БД
# Автор: Igor Kilipenko

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🐳 Инициализация основной PostgreSQL БД...${NC}"

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
    docker-compose down 2>/dev/null || true
}

# Устанавливаем trap для очистки при выходе
trap cleanup EXIT

echo -e "${BLUE}📦 Запуск основной PostgreSQL...${NC}"
docker-compose up -d postgres

echo -e "${YELLOW}⏳ Ожидание готовности БД...${NC}"
# Ждем, пока БД станет доступной
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U metricaldb -d metricaldb >/dev/null 2>&1; then
        echo -e "${GREEN}✅ БД готова!${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ БД не готова через 30 секунд${NC}"
        exit 1
    fi
    sleep 1
done

echo -e "${BLUE}🔧 Выполнение миграций...${NC}"
# Запускаем миграции через Go приложение
if go run ./cmd/migrate/main.go; then
    echo -e "${GREEN}✅ Миграции выполнены!${NC}"
else
    echo -e "${YELLOW}⚠️ Миграции не выполнены (возможно, уже выполнены)${NC}"
fi

echo -e "${GREEN}🎉 Основная БД инициализирована!${NC}"
echo -e "${YELLOW}📊 Подключение: postgres://metricaldb:Secret@localhost:5432/metricaldb${NC}"
echo -e "${YELLOW}🔗 Для подключения к БД: make db-shell${NC}"
echo -e "${YELLOW}📝 Для просмотра логов: make db-logs${NC}"

# Не очищаем БД при выходе
trap - EXIT
