package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type TeamGoals struct {
	RedGoals  int `gorm:"column:redgoals"`
	BlueGoals int `gorm:"column:bluegoals"`
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
	Game             Game
	BlueGoalie       User
	BlueForward      User
	RedGoalie        User
	RedForward       User
	Started          bool
	StartedAt        *time.Time
	EndedAt          *time.Time
	Ended            bool
	BlueGoals        int
	RedGoals         int
	IsMatchPoint     bool
	GoalLimitReached bool
	WinningTeam      string
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
