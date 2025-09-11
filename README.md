# Cadet revenue calculator

Allow to download a note with the differents shipments and calculate how much do you made.

I have my notes across the working day on Google Keep note taking app on my cellphone, this program will hit the Google API and download the notes onto the directory the command is run from.

## Format

The note has to be a text file, with the following format, before processing the format is check, if it is invalid, the user will be ask to correct it.


Name file: month<space>year.txt
First line: Canon<space>int
Subsecuent lines: Entry
Entry: 
Day<space>Date:Procedings
Or
Day<space>Date
M: Movements
Or
Day<space>Date
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
- check it for the correct format, if something wrong, open it with $EDITOR and tell the user where the error is
- after the user exit editor, execute previous step

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

- A main function from where it calls the main menu
