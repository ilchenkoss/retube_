CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY,
        username TEXT NOT NULL UNIQUE,
        telegram_id INTEGER NOT NULL UNIQUE,
        birthday DATE NOT NULL,
        notify_birthday BOOLEAN DEFAULT FALSE
);