-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =========================
-- users
-- =========================
CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- unique + index สำหรับ email แบบ case-insensitive
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));

CREATE TRIGGER trigger_update_timestamp_users
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- =========================
-- todo_groups
--   one user -> many todo_groups
-- =========================
CREATE TABLE IF NOT EXISTS todo_groups (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index ไว้ query group ตาม user
CREATE INDEX IF NOT EXISTS idx_todo_groups_user_id ON todo_groups(user_id);

CREATE TRIGGER trigger_update_timestamp_todo_groups
BEFORE UPDATE ON todo_groups
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- =========================
-- todos
--   one todo_group -> many todos
--   one user       -> many todos (เก็บ user_id ซ้ำไว้ให้ filter ตรง ๆ ได้)
-- =========================
CREATE TABLE IF NOT EXISTS todos (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    date_start DATE,
    date_end DATE,
    is_success BOOLEAN NOT NULL DEFAULT FALSE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    todo_group_id BIGINT NOT NULL REFERENCES todo_groups(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index foreign key ต่าง ๆ
CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id);
CREATE INDEX IF NOT EXISTS idx_todos_todo_group_id ON todos(todo_group_id);

CREATE TRIGGER trigger_update_timestamp_todos
BEFORE UPDATE ON todos
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
-- ลบ trigger ก่อน (ตามลำดับ child -> parent)
DROP TRIGGER IF EXISTS trigger_update_timestamp_todos ON todos;
DROP TRIGGER IF EXISTS trigger_update_timestamp_todo_groups ON todo_groups;
DROP TRIGGER IF EXISTS trigger_update_timestamp_users ON users;

-- ลบ index ของ child ก่อน
DROP INDEX IF EXISTS idx_todos_todo_group_id;
DROP INDEX IF EXISTS idx_todos_user_id;
DROP INDEX IF EXISTS idx_todo_groups_user_id;
DROP INDEX IF EXISTS idx_users_email_lower;

-- ลบตาราง (child -> parent)
DROP TABLE IF EXISTS todos;
DROP TABLE IF EXISTS todo_groups;
DROP TABLE IF EXISTS users;

-- ลบ function ทิ้งท้าย
DROP FUNCTION IF EXISTS update_updated_at_column();
-- +goose StatementEnd
