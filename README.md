# RvBMoGo
RvBMo rewritten in Go. Various improvement have been made to the old version of RvBMo. 

# Features 
Commands/
├── ping/
│   └── Test responsiveness
└── teams/
    ├── create/
    │   ├── by-name/
    │   │   └── Create a team with a specific name
    │   └── batch/
    │       └── Create a specific number of generic teams all at once
    └── delete/
        ├── by-role/
        │   └── Delete a team based using their role as an identifier
        └── all/
            └── Delete all non-protected teams (Default: Green Team, Red Team, RvBMo)
