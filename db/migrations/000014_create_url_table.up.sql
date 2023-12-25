CREATE TABLE nofronts.url(
    id SERIAL PRIMARY KEY,
    long_url VARCHAR(2048) NOT NULL,
    short_url VARCHAR(255) NOT NULL
);
