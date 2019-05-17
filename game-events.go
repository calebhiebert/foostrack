package main

/*
	This file contains the gin handlers for events that occur during a game
	- Goal
	- AntiGoal
	- Game Started
	- Game Ended
	- Dead Ball
	- Out of Bounds
	- Position Swap
*/

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// MarkGoal records a single goal for a given team
func MarkGoal(c *gin.Context) {
	gameID := c.Param("id")
	team := c.Query("team")
	position := c.Query("position")

	if c.Request.Method == "POST" {
		team = c.PostForm("team")
		position = c.PostForm("position")
	}

	var game Game

	if err := dbase.First(&game, "id = ?", gameID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	var scoreUser User

	if err := dbase.Raw(`SELECT u.*
						FROM current_positions cp
							JOIN users u ON cp.user_id = u.id
						WHERE cp.game_id = ? AND position = ? AND team = ?`, game.ID, position, team).
		Scan(&scoreUser).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	event := GameEvent{
		GameID:    game.ID,
		EventType: GameEventGoal,
		UserID:    &scoreUser.ID,
		Team:      team,
		Position:  position,
	}

	if err := dbase.Create(&event).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", game.ID))
}

// MarkAntiGoal records a single anti goal for a given team
func MarkAntiGoal(c *gin.Context) {
	gameID := c.Param("id")
	team := c.Query("team")
	position := c.Query("position")

	if c.Request.Method == "POST" {
		team = c.PostForm("team")
		position = c.PostForm("position")
	}

	var game Game

	if err := dbase.Preload("Events").First(&game, "id = ?", gameID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	var scoreUser User

	if err := dbase.Raw(`SELECT users.*
						FROM game_events
							JOIN users ON game_events.user_id = users.id
						WHERE game_id = ?
						AND game_events.id = (SELECT MAX(id) FROM game_events WHERE position = ? AND team = ?)`, game.ID, position, team).
		Scan(&scoreUser).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	event := GameEvent{
		GameID:    game.ID,
		EventType: GameEventAntiGoal,
		UserID:    &scoreUser.ID,
		Team:      team,
		Position:  position,
	}

	if err := dbase.Create(&event).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", game.ID))
}

// MarkStarted marks a game as begun
func MarkStarted(c *gin.Context) {
	gameID := c.Param("id")

	var game Game

	if err := dbase.First(&game, "id = ?", gameID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	event := GameEvent{
		GameID:    game.ID,
		EventType: GameEventStart,
	}

	if err := dbase.Create(&event).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", game.ID))
}

// MarkEnded marks a game as complete
func MarkEnded(c *gin.Context) {
	gameID := c.Param("id")

	var game GameExtended

	if err := dbase.Raw(`SELECT * FROM game_extended g WHERE g.id = ? LIMIT 1;`, gameID).Scan(&game).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	event := GameEvent{
		GameID:    game.ID,
		EventType: GameEventEnd,
	}

	if err := dbase.Create(&event).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	// Check if this is a tournament game
	var bracketPosition []BracketPosition

	if err := dbase.Find(&bracketPosition, "game_id = ?", game.ID).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	// If this is indeed a tournament game, bump the winning team to the next bracket
	if len(bracketPosition) > 0 && game.BlueTeamID != nil && game.RedTeamID != nil {
		var winningTeamID uint

		if game.BlueGoals == game.WinGoals {
			winningTeamID = *game.BlueTeamID
		} else if game.RedGoals == game.WinGoals {
			winningTeamID = *game.RedTeamID
		}

		// Find the next highest bracket position
		var count Count

		if err := dbase.Raw(`SELECT MAX(bracket_level) AS count
													FROM bracket_positions
													WHERE tournament_id = ?
														AND bracket_level = ?`, bracketPosition[0].TournamentID, bracketPosition[0].BracketLevel+1).Scan(&count).Error; err != nil {
			SendError(http.StatusInternalServerError, c, err)
			return
		}

		nextBracket := BracketPosition{
			TournamentID:    bracketPosition[0].TournamentID,
			TeamID:          winningTeamID,
			BracketLevel:    bracketPosition[0].BracketLevel + 1,
			BracketPosition: count.Count + 1,
		}

		if err := dbase.Create(&nextBracket).Error; err != nil {
			SendError(http.StatusInternalServerError, c, err)
			return
		}

		if err := CheckBracket(bracketPosition[0].TournamentID); err != nil {
			SendError(http.StatusInternalServerError, c, err)
			return
		}
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", game.ID))
}

// MarkDeadBall records a dead ball event
func MarkDeadBall(c *gin.Context) {
	gameID := c.Param("id")

	var game Game

	if err := dbase.Preload("Events").First(&game, "id = ?", gameID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	event := GameEvent{
		GameID:    game.ID,
		EventType: GameEventDeadBall,
	}

	if err := dbase.Create(&event).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", game.ID))
}

// MarkOutOfBounds records an out of bounds event
func MarkOutOfBounds(c *gin.Context) {
	gameID := c.Param("id")

	var game Game

	if err := dbase.Preload("Events").First(&game, "id = ?", gameID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	event := GameEvent{
		GameID:    game.ID,
		EventType: GameEventOutOfBounds,
	}

	if err := dbase.Create(&event).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", game.ID))
}

// MarkSwap swaps the positions of two players on a single team
func MarkSwap(c *gin.Context) {
	gameID := c.Param("id")
	team := c.PostForm("team")

	var currentPositions []GameEvent

	if err := dbase.Raw(`SELECT * FROM current_positions WHERE game_id = ? AND team = ? ORDER BY position ASC`, gameID, team).
		Scan(&currentPositions).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	swapEvents := make([]GameEvent, 0)

	switch team {
	case GameTeamBlue:
		swapEvents = append(swapEvents, GameEvent{
			GameID:    currentPositions[0].GameID,
			EventType: GameEventPlayerTakePosition,
			Team:      GameTeamBlue,
			Position:  GamePositionForward,
			UserID:    currentPositions[1].UserID,
		})

		swapEvents = append(swapEvents, GameEvent{
			GameID:    currentPositions[0].GameID,
			EventType: GameEventPlayerTakePosition,
			Team:      GameTeamBlue,
			Position:  GamePositionGoalie,
			UserID:    currentPositions[0].UserID,
		})
	case GameTeamRed:
		swapEvents = append(swapEvents, GameEvent{
			GameID:    currentPositions[0].GameID,
			EventType: GameEventPlayerTakePosition,
			Team:      GameTeamRed,
			Position:  GamePositionForward,
			UserID:    currentPositions[1].UserID,
		})

		swapEvents = append(swapEvents, GameEvent{
			GameID:    currentPositions[0].GameID,
			EventType: GameEventPlayerTakePosition,
			Team:      GameTeamRed,
			Position:  GamePositionGoalie,
			UserID:    currentPositions[0].UserID,
		})
	}

	tx := dbase.Begin()

	for _, evt := range swapEvents {
		if err := tx.Create(&evt).Error; err != nil {
			tx.Rollback()
			SendError(http.StatusInternalServerError, c, err)
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", currentPositions[0].GameID))
}
