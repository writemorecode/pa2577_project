DROP DATABASE IF EXISTS images;

CREATE DATABASE IF NOT EXISTS images;

USE images;

CREATE TABLE IF NOT EXISTS images (
    id INT AUTO_INCREMENT PRIMARY KEY,
	user_id INT NOT NULL,
	filename VARCHAR(100) UNIQUE NOT NULL,
	upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
