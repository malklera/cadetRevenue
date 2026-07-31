# Cadet revenue calculator

I have my notes across the working day on Google Keep note taking app on my cellphone,
you have to manually create a text file with the correct name, then copy into it the content of the note.

## Format

The note has to be a text file, with the following format.


### Name file

`<year>-<month>-<number>.txt`

Where year is the four digit numerical representation.

Month is the one or two digit numerical representation.

Number is one digit.

### Note

First line

`Canon <int>`

Afterwards
```
<Entry>
M: <movements>
T: <movements>
[Canon]
...
```

Entry is

`Day date`

Or

`Day date:<movements>`

Day is the day of the week in spanish

`Lunes, Martes, Miércoles, Miercoles, Jueves, Viernes, Sábado, Sabado`

Date can be

`<day>/<month>`

Day is from 1 to 31, may or may not be zero padded.

Month is from 1 to 12, may or may not be zero padded.

Movements can be

`0, int, -int, int+int..., int...-int`

A valid note is like this.

`2026-julio-1.txt`

```
Canon 8500
Lunes 20/7:0
Martes 21/7:-1000
Miércoles 22/7
M:2500
T:-3000
Jueves 23/7
M: 2500+3000+2500+8300+5000+3000+7500-6000
T: 2500+5000+4000+2500
Canon 9000
Viernes 24/7
M: 2500+2500+5000+2500+6000+3000-12000
T: 3000+3000+4500+3500+3000-6000

Sábado 25/7
M: 8000+2500+13300+5000+2500+5000
```

White lines are ignored.

The lats entry of the note may end in `M:...`

# Install

`go install `

# Use

`cadetRevenue init`

In the `originals` directory create the notes files.

`cadetRevenue format`

If a note has the wrong format the user will be prompted to edit it.

`cadetRevenue process`

Show all entries available.

`cadetRevenue show`

To show the profit of a day, month or year.

`cadetRevenue profit -year <year> [-month <month>] [-day <day>]`

# TO DO

[ ] Change setup for init and make the goose migration here, change setup command
for init.

[ ] add a `version` command.

[ ] Add `help` command.

[ ] See about erasing the last "," when printing the days in showAll()

[ ] Add target flag logic to show

[ ] Add target flag logic to profit

[ ] `error saving note '2024-abril-5.txt': UNIQUE constraint failed: entry.date`
update this error to show which date is the repeated one?

[ ] list all files in formatedDir

[ ] Do i check the file name each time a list all files in a directory?

[ ] Decide if when showing current line(`nextLineInvalid`) i will show the original
line or the formated one, currently i show the original, so original: T:0, formated:
t:0

[ ] At some point use the `validDate()` to ensure i have possible dates,
not actually correct ones yet.

[ ] Should I save the day? Like "Lunes"

[ ] Rework the user input interaction, need a way of going back, always, try to
make each interaction its own function, just pass the needed checks.

[ ] Test `createEnv`?

[ ] See about just showing year-month with numbers in showProfitMonth

[ ] show help for profit when error of year

[ ] if there are no profits for the given date, show an error indicating that,
not the stack trace

[ ] `m: 100` is valid to format, my function strip the space, but it is not valid
for processing, meaning when i have this `m: 100- 300` and by accident modify to
`m: 100-300` it is considered valid, but the space do not get striped

[ ] improve the `show` output, the years need to be more apparent

# Check this

```
Processing: 2024-abril-5.txt
error saving note '2024-abril-5.txt': UNIQUE constraint failed: entry.date
```

```
Processing: 2025-noviembre-4.txt
error processing note '2025-noviembre-4.txt': dayWorkRe.MatchString(martes 04/11): processProcedings(t: 7300+2000+2500+2300+3000+5500-3300): strings.Contains( 7300+2000+2500+2300+3000+5500-3300, "-"): strconv.ParseInt( 7300, 10, 64): strconv.ParseInt: parsing " 7300": invalid syntax
```

```
Processing: 2026-abril-1.txt
error saving note '2026-abril-1.txt': UNIQUE constraint failed: entry.date
```
