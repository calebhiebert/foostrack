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

	game, err := createGame(blueGoalieID, blueForwardID, redGoalieID, redForwardID, 10)
	if err != nil {
		renderStartGame(c, gin.H{
			"errors": []error{err},
		})
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", game.ID))
}

func createGame(bgID, bfID, rgID, rfID string, winGoals int) (*Game, error) {
	if bgID == rgID ||
		bgID == rfID ||
		bfID == rgID ||
		bfID == rfID {
		return nil, errors.New("One player cannot be on both teams")
	}

	game := Game{
		WinGoals: winGoals,
	}

	// Create database events
	tx := dbase.Begin()

	if err := tx.Create(&game).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	blueGoalieEvent := GameEvent{
		GameID:    game.ID,
		EventType: GameEventPlayerTakePosition,
		UserID:    &bgID,
		Team:      GameTeamBlue,
		Position:  GamePositionGoalie,
	}

	blueForwardEvent := GameEvent{
		GameID:    game.ID,
		EventType: GameEventPlayerTakePosition,
		UserID:    &bfID,
		Team:      GameTeamBlue,
		Position:  GamePositionForward,
	}

	redGoalieEvent := GameEvent{
		GameID:    game.ID,
		EventType: GameEventPlayerTakePosition,
		UserID:    &rgID,
		Team:      GameTeamRed,
		Position:  GamePositionGoalie,
	}

	redForwardEvent := GameEvent{
		GameID:    game.ID,
		EventType: GameEventPlayerTakePosition,
		UserID:    &rfID,
		Team:      GameTeamRed,
		Position:  GamePositionForward,
	}

	if err := tx.Create(&blueGoalieEvent).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Create(&blueForwardEvent).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Create(&redGoalieEvent).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Create(&redForwardEvent).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &game, nil
}
