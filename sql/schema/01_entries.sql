-- +goose Up
CREATE TABLE entry (
    id TEXT PRIMARY KEY,
    date DATE UNIQUE NOT NULL,
    canon INTEGER NOT NULL,
    profit REAL NOT NULL
);

CREATE TABLE movement (
    id TEXT PRIMARY KEY,
    entry_id TEXT NOT NULL,
    shift TEXT NOT NULL,
    amount INTEGER NOT NULL,
    FOREIGN KEY (entry_id) REFERENCES entry (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE entry;
DROP TABLE movement;
