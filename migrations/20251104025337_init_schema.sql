-- +goose Up
-- +goose StatementBegin

-- Подключаем расширение для генерации UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- уникальный идентификатор пользователя
    username        VARCHAR(255) NOT NULL UNIQUE,               -- логин пользователя
    password_hash   VARCHAR(255) NOT NULL,                      -- хеш пароля
    role_id
    points          INT NOT NULL DEFAULT 0,                     -- количество очков/баллов пользователя
    referrer_id     UUID REFERENCES users(id) ON DELETE SET NULL, -- ID пользователя, который пригласил текущего (может быть NULL)
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()            -- дата создания пользователя
);

-- Индекс для топа по баллам
CREATE INDEX IF NOT EXISTS idx_users_points_desc ON users(points DESC); -- для быстрого получения топа пользователей по points

-- Задачи
CREATE TABLE IF NOT EXISTS tasks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- уникальный идентификатор задачи
    code            VARCHAR(100) NOT NULL UNIQUE,               -- системное имя задачи (например: "subscribe_telegram")
    description     TEXT,                                       -- описание задачи
    reward_points   INT NOT NULL,                                -- количество очков за выполнение задачи
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()            -- дата создания задачи
);

-- Выполненные задания
CREATE TABLE IF NOT EXISTS user_tasks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- уникальный идентификатор записи о выполнении
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- пользователь, который выполнил задачу
    task_id         UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE, -- выполненная задача
    completed_at    TIMESTAMP NOT NULL DEFAULT NOW()            -- дата и время выполнения задачи
);

-- Метаданные для выполненных заданий
CREATE TABLE IF NOT EXISTS user_task_metadata (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- уникальный идентификатор метаданных
    user_task_id    UUID NOT NULL REFERENCES user_tasks(id) ON DELETE CASCADE, -- ссылка на выполненное задание
    key             VARCHAR(100) NOT NULL,                        -- ключ метаданных (например: "referrer_id", "social")
    value           TEXT NOT NULL,                                -- значение метаданных
    CONSTRAINT uq_user_task_key_value UNIQUE(user_task_id, key, value) -- запрещаем дублирование одной и той же пары key/value для одного user_task
);

-- Вставка стандартных задач
INSERT INTO tasks (id, code, description, reward_points, created_at) VALUES
    (uuid_generate_v4(), 'enter_referral_code', 'Вводит реферальный код и получает награду', 100, NOW()),
    (uuid_generate_v4(), 'subscribe_telegram', 'Подписывается на Telegram-канал и получает награду', 50, NOW()),
    (uuid_generate_v4(), 'follow_twitter', 'Подписывается на Twitter и получает награду', 50, NOW());



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_task_metadata;
DROP TABLE IF EXISTS user_tasks;
DROP TABLE IF EXISTS tasks;
DROP INDEX IF EXISTS idx_users_points_desc;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
