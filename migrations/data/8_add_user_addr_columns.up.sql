ALTER TABLE users
	ADD COLUMN login_addr VARCHAR(64) NOT NULL DEFAULT '',
	ADD COLUMN login_ts TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT 'epoch',
	ADD COLUMN refresh_addr VARCHAR(64) NOT NULL DEFAULT '',
	ADD COLUMN refresh_ts TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT 'epoch';