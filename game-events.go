package main

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

	var game Game

	if err := dbase.First(&game, "id = ?", gameID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		} else {
			SendError(http.StatusInternalServerError, c, err)
			return
		}
	}

	event := GameEvent{
		GameID:    game.ID,
		EventType: GameEventEnd,
	}

	if err := dbase.Create(&event).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
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
