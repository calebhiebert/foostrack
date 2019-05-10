package main

import (
	"net/http"
	"time"

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
	Elapsed   time.Duration `gorm:"-"`
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
	var events []*GameEvent

	if err := dbase.Preload("User").Order("created_at").Find(&events, "game_id = ?", game.ID).Error; err != nil {
		panic(err)
	}

	// Calculate current game state from events
	for idx, evt := range events {

		// Calculate duration since last event
		if idx == 0 {
			evt.Elapsed = time.Duration(0)
		} else {
			evt.Elapsed = evt.CreatedAt.Sub(events[idx-1].CreatedAt)
		}

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
		"fmtdur":     PrettyDuration,
		"exfname":    ExtractFirstName,
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
