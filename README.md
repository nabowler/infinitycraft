# infinitecraft

A program to check and track pairs within https://neal.fun/infinite-craft/

## How to use

### Prerequisites

1. Install Go
2. Install [just](https://github.com/casey/just)
  - or just copy the commands from the file

#### Without CGO

3. `just run`

#### With CGO

3. Have a gcc compiler on your $PATH
  - required due to the mattn/sqlite3 package
4. `just run_cgo`

## Why

🤷