# Cadet revenue calculator

Allow to download a note with the differents shipments and calculate how much do you made.

I have my notes across the working day on Google Keep note taking app on my cellphone, you have to manually create a text file with the correct name, then copy into it the content of the note.

I saw something about hiting the google api, but is too complicate for me at this moment.

## Format

The note has to be a text file, with the following format, before processing the format is check, if it is invalid, the user will be ask to correct it.


Name file: month-\[number of notes of this month\]-year.txt
Where year is the four digit numerical representation, month is the word on english

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

### menu()

- Ask the user the following options
    - Process Notes
        - checkFileNames() check all .txt files on current directory so it have
        the correct format, ask the user to correct when it do not, return error
        - If error != nil and error != errNoFiles, call listFiles(), return a list of txt files names
        - loop throught the listNotes and on each call checkFormatNote("fileName") return error
    - Show Notes
    - Exit

## Test cases

This should be all valid
```

Canon 7000
Lunes 29/9:-4000
Martes 30/9: 0
Miércoles 1/10
M:2000
T:2000+2000
Jueves 2/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T: 2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M: 2000+2500+4500+4500+4000-2000

```

Error first line has to be Canon int
```

Lunes 29/9:-4000
Martes 30/9: 0
Miércoles 1/10
M:2000
T:2000+2000
Jueves 2/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000+2500+4500+4500+4000-2000

```

Error on canon
```

Canon
Lunes 29/9:-4000
Martes 30/9: 0
Miércoles 1/10
M:2000
T:2000+2000
Jueves 2/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000+2500+4500+4500+4000-2000

```

Error on Lunes
```

Canon 7000
Lunes 29/9:
Martes 30/9: 0
Miércoles 1/10
M:2000
T:2000+2000
Jueves 2/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000+2500+4500+4500+4000-2000

```

Error on Domingo
```

Canon 7000
Domingo 29/9:-4000
Martes 30/9: 0
Miércoles 1/10
M:2000
T:2000+2000
Jueves 2/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000+2500+4500+4500+4000-2000

```

Error on Jueves
```

Canon 7000
Lunes 29/9:-4000
Martes 30/9: 0
Miércoles 1/10
M:2000
T:2000+2000
Jueves 40/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000+2500+4500+4500+4000-2000

```

Error on Viernes
```

Canon 7000
Lunes 29/9:-4000
Martes 30/9: 0
Miércoles 1/10
M:2000
T:2000+2000
Jueves 2/10
M:-4500
T:-4000
Viernes 3/13
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000+2500+4500+4500+4000-2000

```

Error on M of Sabado
```

Canon 7000
Lunes 29/9:-4000
Martes 30/9: 0
Miércoles 1/10
M:2000
T:2000+2000
Jueves 2/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000++2500+4500+4500+4000-2000

```

Error below Lunes
```

Canon 7000
Lunes 29/9:-4000
M:2000
Martes 30/9: 0
Miércoles 1/10
M:2000
T:2000+2000
Jueves 2/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000+2500+4500+4500+4000-2000

```

Error below Martes
```

Canon 7000
Lunes 29/9:-4000
M:2000
Martes 30/9: 0
comprar aceite
Miércoles 1/10
M:2000
T:2000+2000
Jueves 2/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000+2500+4500+4500+4000-2000

```

No really treated as an Error below Miercoles, but it just has to add procedings 0
```

Canon 7000
Lunes 29/9:-4000
Martes 30/9: 0
Miércoles 1/10
Jueves 2/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000+2500+4500+4500+4000-2000

```


```

Canon 7000
Lunes 29/9:-4000
Martes 30/9: 0
Miércoles 1/10
M:2000
T:2000+2000
Jueves 2/10
M:-4500
T:-4000
Viernes 3/10
M:2500+2200+2500+6000+2000+4000+4000+2000-14800
T:2000+5000+2000+2000+2000+3000+2500-3300
Sábado 4/10
M:2000+2500+4500+4500+4000-2000

```
