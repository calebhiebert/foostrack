package main

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

/*
	This file contains models that do not directly map to database tables
*/

// TeamGoals struct is used in a few other structs, it
type TeamGoals struct {
	RedGoals  int `gorm:"column:red_goals"`
	BlueGoals int `gorm:"column:blue_goals"`
}

// GameExtended maps to the game_extended database view
type GameExtended struct {
	gorm.Model
	RedGoals    int        `gorm:"column:red_goals"`
	BlueGoals   int        `gorm:"column:blue_goals"`
	StartTime   *time.Time `gorm:"column:start_time"`
	Started     bool
	EndTime     *time.Time     `gorm:"column:end_time"`
	BlueMembers pq.StringArray `gorm:"column:blue_members;type:VARCHAR(40)[]"`
	RedMembers  pq.StringArray `gorm:"column:red_members;type:VARCHAR(40)[]"`
	WinGoals    int            `gorm:"column:win_goals"`
	Ended       bool
	BlueGoalie  User
	BlueForward User
	RedGoalie   User
	RedForward  User
}

// CurrentGameState represents the current state of a single game
// This object should be calculatable from a game's event stream
type CurrentGameState struct {
	BlueGoalie       User       `json:"blueGoalie"`
	BlueForward      User       `json:"blueForward"`
	RedGoalie        User       `json:"redGoalie"`
	RedForward       User       `json:"redForward"`
	Started          bool       `json:"started"`
	StartedAt        *time.Time `json:"startedAt"`
	EndedAt          *time.Time `json:"endedAt"`
	Ended            bool       `json:"ended"`
	BlueGoals        int        `json:"blueGoals"`
	RedGoals         int        `json:"redGoals"`
	IsMatchPoint     bool       `json:"isMatchPoint"`
	GoalLimitReached bool       `json:"goalLimitReached"`
	WinningTeam      string     `json:"winningTeam"`
}

// UserWithStats corresponds to the user_stats database view
type UserWithStats struct {
	User
	GamesPlayed         int     `gorm:"column:games_played"`
	GamesWon            int     `gorm:"column:games_won"`
	AverageGoalsPerGame float64 `gorm:"column:avg_goals_per_game"`
	GamesPlayedRed      int     `gorm:"column:games_played_red"`
	GamesPlayedBlue     int     `gorm:"column:games_played_blue"`
	AntiGoals           int     `gorm:"column:antigoals"`
	Goals               int     `gorm:"column:goals"`
	NonSaves            int     `gorm:"column:non_saves"`
}

// Count is a helper struct for database queries where the only result is a single count column
type Count struct {
	Count int `gorm:"column:count"`
}

// UserGoals is a helper struct to keep track of the number of goals a user has
type UserGoals struct {
	User      User `json:"user"`
	Goals     int  `json:"goals"`
	AntiGoals int  `json:"antigoals"`

	// A random color
	Color string `json:"color"`
}
