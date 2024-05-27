-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE users ADD COLUMN api_key VARCHAR(64);

UPDATE users SET api_key = encode(sha256(random()::text::bytea), 'hex') WHERE api_key IS NULL;

ALTER TABLE users ALTER COLUMN api_key SET NOT NULL;

ALTER TABLE users ADD CONSTRAINT users_api_key_unique UNIQUE (api_key);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back
ALTER TABLE users DROP CONSTRAINT users_api_key_unique;

ALTER TABLE users DROP COLUMN api_key;

