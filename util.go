package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type TeamGoals struct {
	RedGoals  int `gorm:"column:red_goals"`
	BlueGoals int `gorm:"column:blue_goals"`
}

type GameInfo struct {
	gorm.Model
	TeamGoals
	StartTime   *time.Time `gorm:"column:start_time"`
	Started     bool
	EndTime     *time.Time `gorm:"column:end_time"`
	Ended       bool
	BlueGoalie  User
	BlueForward User
	RedGoalie   User
	RedForward  User
}

// CurrentGameState represents the current state of a single game
type CurrentGameState struct {
	BlueGoalie       User       `json:"blueGoalie"`
	BlueForward      User       `json:"redForward"`
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

type UserWithStats struct {
	User
	GamesPlayed         int     `gorm:"column:games_played"`
	GamesWon            int     `gorm:"column:games_won"`
	AverageGoalsPerGame float64 `gorm:"column:avg_goals_per_game"`
	GamesPlayedRed      int     `gorm:"column:games_played_red"`
	GamesPlayedBlue     int     `gorm:"column:games_played_blue"`
	AntiGoals           int     `gorm:"column:antigoals"`
}

type Count struct {
	Count int `gorm:"column:count"`
}

func SendHTML(statusCode int, c *gin.Context, page string, data gin.H) {

	if data == nil {
		data = gin.H{}
	}

	data["general"] = c.GetStringMapString("general")

	c.HTML(statusCode, page, data)
}

func SendError(code int, c *gin.Context, err error) {
	SendHTML(code, c, "error", gin.H{
		"error": err,
	})
}

func SendNotFound(c *gin.Context) {
	SendHTML(http.StatusNotFound, c, "notfound", nil)
}

func EnsureLoggedIn(c *gin.Context) bool {
	general := c.GetStringMapString("general")

	if general["isloggedin"] != "true" {
		SendHTML(http.StatusForbidden, c, "blocked", nil)
		return false
	}

	return true
}
