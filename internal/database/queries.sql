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

-- return a year-month and the net profit
-- select e1.id, e1.strftime('%Y %m', e1.date), e1.canon, () as netRevenue
-- from entry as e1;

select *
from entry
where strftime('%Y', date) = '2025';

-- what years are available
select distinct strftime('%Y', date)
from entry
order by date;

-- all months of the given year
select distinct strftime('%m', date)
from entry
where strftime('%Y', date) = '?'
order by date;

-- all entries of a given month
select distinct *
from entry
where strftime('%m', date) = '01'
order by date;

