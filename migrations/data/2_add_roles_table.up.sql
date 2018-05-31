CREATE TABLE roles(
	user_id UUID,
	role VARCHAR(1024),
	PRIMARY KEY (user_id, role)
);