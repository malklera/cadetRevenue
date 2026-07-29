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
FROM entry
ORDER BY date;

-- name: ListAllMovements :many
SELECT *
FROM movement;

-- name: ListAvailableDates :many
SELECT date
FROM entry
ORDER BY date;

-- name: ProfitDay :one
SELECT profit
FROM entry
WHERE date = ?;

-- name: ProfitMonth :one
SELECT SUM(profit)
FROM entry
WHERE date >= ? AND date < ?;
