-- +goose Up
CREATE TABLE entry (
    id INTEGER PRIMARY KEY,
    date TEXT UNIQUE,
    canon INTEGER NOT NULL,
    profit REAL NOT NULL
);

CREATE TABLE movement (
    id INTEGER PRIMARY KEY,
    entry_id INTEGER NOT NULL,
    shift TEXT NOT NULL,
    ammount INTEGER NOT NULL,
    FOREIGN KEY (entry_id) REFERENCES entry (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE entry;
DROP TABLE movement;
