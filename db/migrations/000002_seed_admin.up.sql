INSERT INTO users (email, password_hash)
VALUES ('admin@waxp.com', '$2a$10$ysLGEY/eUOniH2eVzRGpQ.SmVS7PfQZOLaQ4QGgfcpF0E8uO98Tz6')
ON CONFLICT (email) DO NOTHING;
