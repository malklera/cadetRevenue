# Cadet revenue calculator

Allow to download a note with the different shipments and calculate how much do you made.

I have my notes across the working day on Google Keep note taking app on my cellphone, you have to manually create a text file with the correct name, then copy into it the content of the note.

I saw something about hitting the google API, but is too complicate for me at this moment.

## Format

The note has to be a text file, with the following format, before processing the format is check, if it is invalid, the user will be ask to correct it.


Name file: month-\[number of notes of this month\]-year.txt
Where year is the four digit numerical representation, month is the word on English

First line: Canon<space>int

After first line: Entry
    Entry: 
        Day<space>Date:Procedings
        Or
        Day<space>Date
After Day<space>Date:Procedings: Entry
After Day<space>Date: Turn
    Turn:
        M: Movements
        Or
        M: Movements
        T: Movements
Followed by either another entry or end of file

Where Day can be:
Lunes, Martes, Miércoles, Miercoles, Jueves, Viernes, Sábado, Sabado

Date can be:
1/1, 01/01, 01/1, 1/01
Each has to be a valid date of the corresponding month-year

Procedings can be: ignore spaces
0, -int

Movements can be: ignore spaces
0, int, -int, int+int..., int...-int

After a ":" there may be a <space>, strip it
How to do it
Menu that ask
1) process files
2) show files
3) exit

1) do
- list all valid files on the directory the command is run and show it
- pick the first file fron the list
- call checkFormat on it
- with the correct format, call processFile on it
- move file to "originals" directory
- select next file, repeat previous three steps

checkFormat: 
- check it for the correct format, if something wrong, present a input field prefilled
with the wrong content and let the user change it.
- Check the corrected input from the user.

processFile:
- read whole file into a list where each item is a line of the file
- for loop the whole list
- case for this options
+ Case "canon", get the int and put it on canon, jump to next line
+ Case ":", take the date out, put 0 as income of the day, if " - ", put the number after as expense of the day, jump to next line
+ Case "Day" (check if it enter the case ":" it would or nlt enter this) take the date out, read the next line, add its positive movements and put then as income, if there is negative put it as expense, read next line, take out the values, jump 2 lines


Database
Table entries
Row: id, year int, month int, day int, canon int, income int, expenses int
(year, month, day) is unique

## Structure of the program

It is a CLI with the following flags available.

`-format` `-f`
    Ensure the files in `originals` are correctly formatted, if wrong, point out
    the error and prompt the user into correcting it.

`-process` `-p`
    Retrieve all data from files in `originals` and save them to the db, move the
    successfully processed files into `processed`, skip and return error if any
    file is not the correct format, continue with the rest of files.

`-`


- A main function from where it calls the main menu

### menu()

- Ask the user the following options
    - Process Notes
        - checkFileNames() check all .txt files on current directory so it have
        the correct format, ask the user to correct when it do not, return error
        - If error != nil and error != errNoFiles, call listFiles(), return a list of txt files names
        - loop throught the listNotes and on each call checkFormatNote("fileName") return error
    - Show Notes
    - Exit

## TO DO:

[x] Change checkFileNames() to checkFileName() the actual implementation, i have listFiles() so i should call it and loop through the return and call checkFileName() on each

[x] Test checkFileName()

[x] Make sure a note has content, otherwise the program panic(found out while testing checkFileName())

[x] On checkFormatNote() if I encounter an empty line, erase it without asking the user

[x] Change my regex's, will make all strings lower case, simpler to dealth with

[x] Test all my test cases for checkFormatNote()

[x] NEED to test checkFileName() again, change the structure of the project, now files live either on originals/ or formated/ or processed/

[x] FIX error on formating, canon gets included twice at the beginning of a file

[x] Change the formating to add a T:0 to sabado, otherwise it will complicate the data extraction

[x] Change the formating to add 0 padding to dates

[x] Create a function to extract data from correctly formated notes, and put it on the DB

[x] Create the DB schema and a initial selection of functions to interface with it, create table, input data, get data out?? think about this one more

[x] Move processed notes to the processed/ directory

[x] Do some calculations for my notes, like getting a net profict daily

[x] Make something for the Show menu, I have to think about what I care about

[x] Check all TODO and NOTE on the project and dealt with them

[x] For now the used directories have to be created manually

[x] For now, this will be the state of it, it do what i want, need to study other things for now


# TODO v2

[x] Change the shape of the cli, setup, show, format, process, will be subcommands.

[x] Ensure data for a specific day is unique, I copied the content of abril-4-2024.txt to abril-5-2024.txt to test

[x] Fix the help messages.

[x] Think if the subcommands should be flagsets or not.

[x] Check how the db works, need to rework it, for now, leave it as is.

[x] Think which flags are global(target) and belong to a subcommand. show(year, month, day)

[x] Update all errors to a uniform format.

[x] Ensure data for each day is unique in the database.

[x] Change switch to use ',' for all that have a equal action

[x] Change the name of the originals files, to year-month-n.txt, try to see how
they get sorted. Has to update `processNote()`.

[x] fix `validFirstLine()`

[x] Test `validFirstLine`

[x] When `formatNote()` if there is a space anywhere, strip it.

[x] All line check should check if there is an entry for the next line.

[x] I think when i modify a line in the last line, like 'sabado ... m:' i do not
add the needed 't:0' below.

[x] `Lunes 29/9:` is this an error or consider it like `Lunes 29/9:0`?

[x] When testing `formatNote()` ensure it output the first line canon

[x] from the file name take the year

[x] take canon, day-month, morning movements, afternoon movements, expenses

[x] calculate the net profit of the day

[x] Refactor moving related functions to their own file.

[x] Remake the database.

[x] Change the storage of income



[ ] fix the logic with dayNoWorkRe, check the ones in testFormat to see the error

[ ] fix logic of canonRe, same as above.

[ ] `error saving note '2024-abril-5.txt': UNIQUE constraint failed: entry.date`
update this error to show which date is the repeated one

[ ] Work on process.go

[ ] add a `version` sub-command

[ ] list all files in formatedDir

[ ] save the day entry into the db

[ ] Do i check the file name each time a list all files in a directory?

[ ] Decide if when showing current line(`nextLineInvalid`) i will show the original
line or the formated one, currently i show the original, so original: T:0, formated:
t:0

[ ] At some point use the `validDate()` to ensure i have possible dates,
not actually correct ones yet.

[ ] Rework the user input interaction, need a way of going back, always, try to
make each interaction its own function, just pass the needed checks.

[ ] Calculate the revenue of a given month.

[ ] Change setup for init and make the goose migration here.

[ ] Test `createEnv`?


## Test cases

Files to copy from backup/ to test
cp backup/febrero-*-2026.txt test-13-05/originals

For format with errors
cp backup/agosto-1-2-2024.txt test-13-05/originals
cp backup/diciembre-2-2024.txt test-13-05/originals
cp backup/enero-2-2026.txt test-13-05/originals
cp backup/enero-4-2024.txt test-13-05/originals
cp backup/noviembre-4-2024.txt test-13-05/originals
cp backup/septiembre-4-2025.txt test-13-05/originals
cp backup/abril-1-2024.txt test-13-05/originals

cp backup/ test-13-05/originals
cp backup/ test-13-05/originals

## Valid movement

m:0
t:0
m: 10
m:10
t:-10
m:10+19
t:10-18
Only one negative number.
Any number of sums
