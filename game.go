package main

/*
	This page contains gin handlers for single-game related things
	For specific game event handlers, see game-events.go
*/

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/lucasb-eyer/go-colorful"
)

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

	if err := dbase.Preload("User").Order("id").Find(&events, "game_id = ?", game.ID).Error; err != nil {
		panic(err)
	}

	userGoals := make(map[string]UserGoals)

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
			if val, ok := userGoals[*evt.UserID]; ok {
				userGoals[*evt.UserID] = UserGoals{
					User:      evt.User,
					Goals:     val.Goals + 1,
					AntiGoals: val.AntiGoals,
				}
			}
			switch evt.Team {

			// Blue Team
			case GameTeamBlue:
				gameState.BlueGoals++

				// Red Team
			case GameTeamRed:
				gameState.RedGoals++
			}

		case GameEventAntiGoal:
			if val, ok := userGoals[*evt.UserID]; ok {
				userGoals[*evt.UserID] = UserGoals{
					User:      evt.User,
					AntiGoals: val.AntiGoals + 1,
					Goals:     val.Goals,
				}
			}
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
			if _, ok := userGoals[*evt.UserID]; !ok {
				userGoals[*evt.UserID] = UserGoals{
					User: evt.User,
				}
			}
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

	userGoalsArr := make([]UserGoals, 0)

	for _, ug := range userGoals {
		userGoalsArr = append(userGoalsArr, UserGoals{
			User:      ug.User,
			Color:     colorful.FastHappyColor().Hex(),
			Goals:     ug.Goals,
			AntiGoals: ug.AntiGoals,
		})
	}

	SendHTML(http.StatusOK, c, "game", gin.H{
		"id":         id,
		"game":       game,
		"gameState":  gameState,
		"events":     events,
		"eventCount": len(events),
		"fmtdur":     PrettyDuration,
		"exfname":    ExtractFirstName,
		"userGoals":  userGoalsArr,
	})
}

// GetGameEventCount returns a json object containing the number of events in a given game
func GetGameEventCount(c *gin.Context) {
	id := c.Param("id")

	var eventCount Count

	if err := dbase.Raw(`SELECT COUNT(id) FROM game_events WHERE game_id = ? AND deleted_at IS NULL`, id).Scan(&eventCount).Error; err != nil {
		SendError(http.StatusBadRequest, c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"eventCount": eventCount.Count,
	})
}
