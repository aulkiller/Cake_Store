CREATE TABLE IF NOT EXISTS `cakes` (
	id INTEGER PRIMARY KEY AUTO_INCREMENT,
	title VARCHAR(100) NOT NULL,
    description VARCHAR(250),
    rating FLOAT,
    image VARCHAR(255),
    created_at DATETIME,
    updated_at DATETIME
);