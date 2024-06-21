CREATE TABLE IF NOT EXISTS subscriptions (
    id INTEGER PRIMARY KEY,
    subscriber INTEGER NOT NULL,
    subscribe_to INTEGER NOT NULL,
    FOREIGN KEY (subscriber) REFERENCES users(id),
    FOREIGN KEY (subscribe_to) REFERENCES users(id),
    UNIQUE (subscriber, subscribe_to)
);