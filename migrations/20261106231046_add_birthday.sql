-- +goose Up
ALTER TABLE users ADD COLUMN birthday DATE;
ALTER TABLE settings ADD COLUMN track_birthdays BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE settings ADD COLUMN minimum_age INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE settings DROP COLUMN minimum_age;
ALTER TABLE settings DROP COLUMN track_birthdays;
ALTER TABLE users DROP COLUMN birthday;
