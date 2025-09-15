-- Миграция 001: Создание таблиц для метрик
-- Автор: Igor Kilipenko
-- Описание: Создание базовых таблиц для хранения gauge и counter метрик

-- Настройка PostgreSQL
SET standard_conforming_strings = on;

-- Создание таблицы для gauge метрик
CREATE TABLE IF NOT EXISTS gauge_metrics (
    id VARCHAR(255) PRIMARY KEY,
    value DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для counter метрик
CREATE TABLE IF NOT EXISTS counter_metrics (
    id VARCHAR(255) PRIMARY KEY,
    value BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_gauge_metrics_id ON gauge_metrics(id);
CREATE INDEX IF NOT EXISTS idx_counter_metrics_id ON counter_metrics(id);

-- Создание индексов для временных меток
CREATE INDEX IF NOT EXISTS idx_gauge_metrics_updated_at ON gauge_metrics(updated_at);
CREATE INDEX IF NOT EXISTS idx_counter_metrics_updated_at ON counter_metrics(updated_at);

-- Комментарии к таблицам
COMMENT ON TABLE gauge_metrics IS 'Таблица для хранения gauge метрик (значения типа float64)';
COMMENT ON TABLE counter_metrics IS 'Таблица для хранения counter метрик (значения типа int64)';

COMMENT ON COLUMN gauge_metrics.id IS 'Уникальный идентификатор метрики';
COMMENT ON COLUMN gauge_metrics.value IS 'Значение gauge метрики (double precision)';
COMMENT ON COLUMN gauge_metrics.created_at IS 'Время создания записи';
COMMENT ON COLUMN gauge_metrics.updated_at IS 'Время последнего обновления';

COMMENT ON COLUMN counter_metrics.id IS 'Уникальный идентификатор метрики';
COMMENT ON COLUMN counter_metrics.value IS 'Значение counter метрики (bigint)';
COMMENT ON COLUMN counter_metrics.created_at IS 'Время создания записи';
COMMENT ON COLUMN counter_metrics.updated_at IS 'Время последнего обновления';
