ALTER TABLE api_keys
ALTER COLUMN user_id TYPE UUID USING user_id::uuid;