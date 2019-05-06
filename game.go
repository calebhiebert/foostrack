package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// GameEvent constants
const (
	GameEventStart              = "start"
	GameEventEnd                = "end"
	GameEventPlayerTakePosition = "ptp"
	GameEventGoal               = "goal"
)

// Game team constants
const (
	GameTeamRed  = "red"
	GameTeamBlue = "blue"
)

// Game positions
const (
	GamePositionForward = "forward"
	GamePositionGoalie  = "goalie"
)

// Game represents the game table
type Game struct {
	gorm.Model
	Events []GameEvent `gorm:"foreignkey:GameID"`
}

// GameEvent represents the game events table
type GameEvent struct {
	gorm.Model
	GameID    uint `gorm:"not null"`
	Game      Game `gorm:"association_foreignkey:GameID;"`
	UserID    string
	User      User `gorm:"association_foreignkey:UserID"`
	EventType string
	Team      string
	Position  string
}

// GetGame renders the game view page
func GetGame(c *gin.Context) {

	id := c.Param("id")

	var game Game

	if err := dbase.Find(&game, "id = ?", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.HTML(http.StatusBadRequest, "notfound", gin.H{
				"general": c.GetStringMapString("general"),
			})
			return
		} else {
			c.HTML(http.StatusBadRequest, "error", gin.H{
				"error":   err,
				"general": c.GetStringMapString("general"),
			})
			return
		}
	}

	c.HTML(http.StatusOK, "game", gin.H{
		"id":      id,
		"game":    game,
		"general": c.GetStringMapString("general"),
	})
}
