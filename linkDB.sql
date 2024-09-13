CREATE TABLE UrlShortener (
    shortUrl TEXT PRIMARY KEY,
    redirectUrl TEXT NOT NULL,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

/*
INSERT INTO UrlShortener (shortUrl, redirectUrl, created) VALUES
('g34fca2Adx','https://yahoo.com','2024-08-14 08:56:22'),
('g34fca2Ady','https://amazon.com','2024-08-14 08:57:22'),
('g34fca2Adz','https://youtube.com','2024-08-14 08:59:22');
*/

