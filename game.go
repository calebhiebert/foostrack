package main

import (
	"fmt"
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
	GameEventAntiGoal           = "antigoal"
	GameEventDeadBall           = "dead"
	GameEventOutOfBounds        = "oob"
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
	Events   []GameEvent `gorm:"foreignkey:GameID"`
	WinGoals int         `gorm:"column:win_goals"`
}

// GameEvent represents the game events table
type GameEvent struct {
	gorm.Model
	GameID    uint `gorm:"not null"`
	Game      Game `gorm:"association_foreignkey:GameID;"`
	UserID    *string
	User      User `gorm:"association_foreignkey:UserID;foreignkey:ID"`
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
			SendNotFound(c)
			return
		}

		SendError(http.StatusBadRequest, c, err)
		return
	}

	var gameState CurrentGameState

	// Select Event List
	var events []GameEvent

	if err := dbase.Preload("User").Find(&events, "game_id = ?", game.ID).Order("created_at ASC").Error; err != nil {
		panic(err)
	}

	// Calculate current game state from events
	for _, evt := range events {
		switch evt.EventType {

		// Count goals
		case GameEventGoal:
			switch evt.Team {

			// Blue Team
			case GameTeamBlue:
				gameState.BlueGoals++

				// Red Team
			case GameTeamRed:
				gameState.RedGoals++
			}

		case GameEventAntiGoal:
			switch evt.Team {
			// Blue Team
			case GameTeamBlue:
				gameState.RedGoals++

				// Red Team
			case GameTeamRed:
				gameState.BlueGoals++
			}

			// Assign players to the correct positions on the team
		case GameEventPlayerTakePosition:
			switch evt.Team {

			// Assign blue team players
			case GameTeamBlue:
				switch evt.Position {

				// Forward
				case GamePositionForward:
					gameState.BlueForward = evt.User

					// Goalie
				case GamePositionGoalie:
					gameState.BlueGoalie = evt.User
				}

				// Assign red team players
			case GameTeamRed:
				switch evt.Position {

				// Forward
				case GamePositionForward:
					gameState.RedForward = evt.User

					// Goalie
				case GamePositionGoalie:
					gameState.RedGoalie = evt.User
				}
			}

			// Assign game started event
		case GameEventStart:
			gameState.StartedAt = &evt.CreatedAt
			gameState.Started = true

			// Assign game ended event
		case GameEventEnd:
			gameState.EndedAt = &evt.CreatedAt
			gameState.Ended = true
		}
	}

	gameState.IsMatchPoint = gameState.BlueGoals == game.WinGoals-1 || gameState.RedGoals == game.WinGoals-1
	gameState.GoalLimitReached = gameState.BlueGoals == game.WinGoals || gameState.RedGoals == game.WinGoals

	if gameState.BlueGoals == game.WinGoals {
		gameState.WinningTeam = GameTeamBlue
	} else if gameState.RedGoals == game.WinGoals {
		gameState.WinningTeam = GameTeamRed
	}

	SendHTML(http.StatusOK, c, "game", gin.H{
		"id":         id,
		"game":       game,
		"gameState":  gameState,
		"events":     events,
		"eventCount": len(events),
	})
}

func GetGameEventCount(c *gin.Context) {
	id := c.Param("id")

	var eventCount Count

	if err := dbase.Raw(`SELECT COUNT(id) FROM game_events WHERE game_id = ?`, id).Scan(&eventCount).Error; err != nil {
		SendError(http.StatusBadRequest, c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"eventCount": eventCount.Count,
	})
}

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
