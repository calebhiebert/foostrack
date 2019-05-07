package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetStartGame(c *gin.Context) {
	if !EnsureLoggedIn(c) {
		return
	}

	var users []User

	if err := dbase.Find(&users).Error; err != nil {
		SendError(http.StatusBadRequest, c, err)
		return
	}

	SendHTML(http.StatusOK, c, "startgame", gin.H{
		"users": users,
	})
}

func PostStartGame(c *gin.Context) {
	if !EnsureLoggedIn(c) {
		return
	}

	game := Game{}

	if err := dbase.Create(&game).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	blueGoalieID := c.PostForm("blue_goalie")
	blueForwardID := c.PostForm("blue_forward")
	redGoalieID := c.PostForm("red_goalie")
	redForwardID := c.PostForm("red_forward")

	blueGoalieEvent := GameEvent{
		GameID:    game.ID,
		EventType: GameEventPlayerTakePosition,
		UserID:    &blueGoalieID,
		Team:      GameTeamBlue,
		Position:  GamePositionGoalie,
	}

	blueForwardEvent := GameEvent{
		GameID:    game.ID,
		EventType: GameEventPlayerTakePosition,
		UserID:    &blueForwardID,
		Team:      GameTeamBlue,
		Position:  GamePositionForward,
	}

	redGoalieEvent := GameEvent{
		GameID:    game.ID,
		EventType: GameEventPlayerTakePosition,
		UserID:    &redGoalieID,
		Team:      GameTeamRed,
		Position:  GamePositionGoalie,
	}

	redForwardEvent := GameEvent{
		GameID:    game.ID,
		EventType: GameEventPlayerTakePosition,
		UserID:    &redForwardID,
		Team:      GameTeamRed,
		Position:  GamePositionForward,
	}

	// Create database events
	tx := dbase.Begin()

	if err := tx.Create(&blueGoalieEvent).Error; err != nil {
		tx.Rollback()
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	if err := tx.Create(&blueForwardEvent).Error; err != nil {
		tx.Rollback()
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	if err := tx.Create(&redGoalieEvent).Error; err != nil {
		tx.Rollback()
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	if err := tx.Create(&redForwardEvent).Error; err != nil {
		tx.Rollback()
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", game.ID))
}
