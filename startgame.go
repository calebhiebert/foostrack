package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
	This file contains gin handlers related to creating a new game
	TODO add some error handling when creating a new game
*/

func renderStartGame(c *gin.Context, data gin.H) {
	if !EnsureLoggedIn(c) {
		return
	}

	var users []User

	if err := dbase.Find(&users).Error; err != nil {
		SendError(http.StatusBadRequest, c, err)
		return
	}

	data["users"] = users

	SendHTML(http.StatusOK, c, "startgame", data)
}

// GetStartGame renders the form that should be filled out to start a game
func GetStartGame(c *gin.Context) {
	renderStartGame(c, gin.H{})
}

// PostStartGame will create a new game
func PostStartGame(c *gin.Context) {
	if !EnsureLoggedIn(c) {
		return
	}

	blueGoalieID := c.PostForm("blue_goalie")
	blueForwardID := c.PostForm("blue_forward")
	redGoalieID := c.PostForm("red_goalie")
	redForwardID := c.PostForm("red_forward")

	if blueGoalieID == redGoalieID ||
		blueGoalieID == redForwardID ||
		blueForwardID == redGoalieID ||
		blueForwardID == redForwardID {
		renderStartGame(c, gin.H{
			"errors": []error{errors.New("One player cannot be on both teams")},
		})
		return
	}

	game := Game{
		WinGoals: 10,
	}

	// Create database events
	tx := dbase.Begin()

	if err := tx.Create(&game).Error; err != nil {
		tx.Rollback()
		SendError(http.StatusInternalServerError, c, err)
		return
	}

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
