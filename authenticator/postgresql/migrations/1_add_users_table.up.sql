CREATE TABLE users(
	id UUID PRIMARY KEY,
	email TEXT UNIQUE,
	password_hash TEXT NOT NULL,
	first_name TEXT,
	last_name TEXT,
	added TIMESTAMP DEFAULT NOW(),
	modified TIMESTAMP DEFAULT NOW()
);