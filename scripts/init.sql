DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS users;
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    isActive BOOLEAN NOT NULL
);
CREATE TABLE IF NOT EXISTS items (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    count INT NOT NULL,
    userId VARCHAR(255) NOT NULL REFERENCES users(id)
);
INSERT INTO users (id, username, password, isActive)
VALUES ('1', 'admin', 'admin', true);