-- CreateSchema()
CREATE TABLE IF NOT EXISTS entry (
		id INTEGER PRIMARY KEY,
		date TEXT NOT NULL,
		canon INTEGER NOT NULL,
		incomeM INTEGER NOT NULL,
		incomeT INTEGER NOT NULL,
		expenses INTEGER NOT NULL
);

-- AddEntry()
INSERT INTO entry (
	date, canon, incomeM, incomeT, expenses) VALUES (?, ?, ?, ?, ?);

-- ShowAll()
SELECT id, date, canon, incomeM, incomeT, expenses
FROM entry;
