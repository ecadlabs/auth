ALTER TABLE users ADD CONSTRAINT users_service_no_email_pwd CHECK (account_type = 'regular' OR (email = '' AND password_hash = ''));
