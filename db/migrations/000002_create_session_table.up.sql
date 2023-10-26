CREATE TABLE nofronts.session(
	id VARCHAR(32) PRIMARY KEY,
	created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
	user_id BIGINT UNIQUE REFERENCES nofronts.user(id)
);
