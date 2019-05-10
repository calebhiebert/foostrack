package main

import (
	"fmt"
	"net/http"
	"strings"
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
	Goals               int     `gorm:"column:goals"`
}

type Count struct {
	Count int `gorm:"column:count"`
}

// SendHTML is a wrapper around gin context HTML function, it includes the "general"
// object to be sent to the template. This object contains info such as whether or not
// the user is logged in, their username, etc...
func SendHTML(statusCode int, c *gin.Context, page string, data gin.H) {

	if data == nil {
		data = gin.H{}
	}

	data["general"] = c.GetStringMapString("general")

	c.HTML(statusCode, page, data)
}

// SendError is a wrapper around SendHTML that sends the error.html template
func SendError(code int, c *gin.Context, err error) {
	SendHTML(code, c, "error", gin.H{
		"error": err,
	})
}

// SendNotFound is a wrapper around SendHTML that sends a 404 page
func SendNotFound(c *gin.Context) {
	SendHTML(http.StatusNotFound, c, "notfound", nil)
}

// EnsureLoggedIn will send a not logged in page/message and return false if the
// current user is not logged in, does nothing and returns true otherwise
func EnsureLoggedIn(c *gin.Context) bool {
	general := c.GetStringMapString("general")

	if general["isloggedin"] != "true" {
		SendHTML(http.StatusForbidden, c, "blocked", nil)
		return false
	}

	return true
}

// PrettyDuration formats a time.Duration into a string, (only goes up to minutes)
func PrettyDuration(duration time.Duration) string {
	seconds := duration.Seconds()

	remainingSeconds := int64(seconds) % 60
	remainingMinutes := (int64(seconds) - remainingSeconds) / 60

	if remainingMinutes > 0 {
		return fmt.Sprintf("%d min %d sec", remainingMinutes, remainingSeconds)
	} else {
		return fmt.Sprintf("%d sec", remainingSeconds)
	}
}

// ExtractFirstName returns the first name from a full name
func ExtractFirstName(name string) string {
	nameParts := strings.Split(name, " ")

	if len(nameParts) > 0 {
		return nameParts[0]
	}

	return ""
}
