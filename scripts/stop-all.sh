#!/bin/bash

# Скрипт для остановки всех сервисов
# Автор: Igor Kilipenko

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🛑 Остановка всех сервисов...${NC}"

# Остановка сервера
echo -e "${YELLOW}⏹️ Остановка сервера...${NC}"
SERVER_PID=$(pgrep -f "bin/server" 2>/dev/null || echo "")
if [ -n "$SERVER_PID" ]; then
    pkill -f "bin/server" 2>/dev/null
    sleep 1
    # Проверяем, что процесс действительно остановлен
    if ! pgrep -f "bin/server" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Сервер остановлен${NC}"
    else
        echo -e "${RED}❌ Не удалось остановить сервер${NC}"
    fi
else
    echo -e "${YELLOW}⚠️ Сервер не был запущен${NC}"
fi

# Остановка агента
echo -e "${YELLOW}⏹️ Остановка агента...${NC}"
AGENT_PID=$(pgrep -f "bin/agent" 2>/dev/null || echo "")
if [ -n "$AGENT_PID" ]; then
    pkill -f "bin/agent" 2>/dev/null
    sleep 1
    # Проверяем, что процесс действительно остановлен
    if ! pgrep -f "bin/agent" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Агент остановлен${NC}"
    else
        echo -e "${RED}❌ Не удалось остановить агент${NC}"
    fi
else
    echo -e "${YELLOW}⚠️ Агент не был запущен${NC}"
fi

# Остановка БД
echo -e "${YELLOW}⏹️ Остановка базы данных...${NC}"
if docker-compose ps postgres | grep -q "Up"; then
    docker-compose down
    echo -e "${GREEN}✅ База данных остановлена${NC}"
else
    echo -e "${YELLOW}⚠️ База данных не была запущена${NC}"
fi

echo -e "${GREEN}🎉 Все сервисы остановлены!${NC}"
