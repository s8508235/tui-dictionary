## v0.1.0 (2022-08-14)

### Fix

- add missing cmd
- add missing break line
- fix output file as io.Writer no assigned
- add pagination with single line word definition
- add build tags for cloudflare scraper
- bypass color code problem with blink by putting warn message after it
- use alt screen
- add back validate and padding check
- correct exit position and add nosec to Exit function
- add client for single search
- trim space for input word and collins crawler selector update
- trim suffix for windows change line
- replace with cross platform methods and add build method for them

### Refactor

- update go mod and remove function scope cmd
- **all**: refactor app to quizlet output using bubbletea
- add migration to text file and update urban query selector
- fix golang lint errcheck
- enhance display function
- **dictionary-refactor-with-general-web-crawler-struct**: replacing dict.org with learner's dictionary in my prefer dictionary
- let terminal cursor move using other lib for input

### Feat

- add cancel for searching state
- add interruptable search and dev null as destination
- add single line scroll for dictionarySelectDef
- add collins client by using python3 cloudflarescaper on linux only
- add support for suggestion of existing target
- **dictionary**: add webster dictionary
- **dictionary**: remove cache for searching web dictionaries
- add spinner
- **tools**: add cross platform Clear/Lines/Cols
- add clear terminal util and print function improvement
- add new dictionarys as urban/dict.org and combination of them
- add simple dictionary client for top 3 definition using collins
