-- name: CreateEntry :one
INSERT INTO entry (
date, canon, profit
) VALUES (
?, ?, ?
)
RETURNING *;

-- name: CreateMovement :exec
INSERT INTO movement (
entry_id, shift, amount
) VALUES (
?, ?, ?
);

-- name: ListAllEntries :many
SELECT *
FROM entry;

-- name: ListAllMovements :many
SELECT *
FROM movement;
