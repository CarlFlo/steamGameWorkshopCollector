# Steam Game Workshop Collector

This project will scrape a steam games workshop page and create a list with all workshop ID's

## Usage
The program accepts several flags, where only the gameID is mandatory. The other flags are optional and have default values if not provided.

### Flags
**gameID** (string): ID of the game. This is a required flag.
**startPage** (int): Which workshop page to start at. Default is 1.
**endPage** (int): Set which workshop page to end at. Default is 0. (meaing it will go forever)
**delay** (int): Set the delay between each request (in milliseconds). Default is 25 ms.
**randomDelay** (int): Add an extra randomized duration to wait added to delay before creating a new request (in milliseconds). Default is 0 ms.

```bash
program.exe -gameID <GameID> [-startPage <StartPage>] [-endPage <EndPage>] [-delay <Delay>] [-randomDelay <RandomDelay>]

```

Start from page 2 and end at page 5.
```bash
program.exe -gameID 108600 -startPage 2 -endPage 5

```