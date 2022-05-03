# RvBMoGo
RvBMo rewritten in Go. Various improvement have been made to the old version of RvBMo. 

# Features 
```
Commands/
├── ping
│   └── Test responsiveness
└── teams/
    ├── create/
    │   ├── by-name
    │   │   └── Create a team with a specific name
    │   └── batch
    │       └── Create a specific number of generic teams all at once
    └── delete/
        ├── by-role
        │   └── Delete a team based using their role as an identifier
        └── all
            └── Delete all non-protected teams (Default: Green Team, Red Team, RvBMo)
```

# Setup
1. Download `main.go` or copy its contents into a .go file
2. Put your bot's token in BotToken field in the file.
3. Run `go mod init rvbmo`, then `go mod tidy`
4. Run `go run ./main.go` or build an executable with `go build` and run that.
