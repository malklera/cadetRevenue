-- name: CreateEntry :exec
INSERT INTO entry (
id, date, canon, profit
) VALUES (
?, ?, ?, ?
);

-- name: CreateMovement :exec
INSERT INTO movement (
id, entry_id, shift, amount
) VALUES (
?, ?, ?, ?
);

-- name: ListAllEntries :many
SELECT *
FROM entry;

-- name: ListAllMovements :many
SELECT *
FROM movement;
