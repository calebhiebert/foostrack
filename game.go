package main

import (
	"fmt"
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
	UserID    *string
	User      User `gorm:"association_foreignkey:UserID"`
	EventType string
	Team      string
	Position  string
}

type CurrentGameState struct {
	Game        Game
	BlueGoalie  User
	BlueForward User
	RedGoalie   User
	RedForward  User
	Started     bool
	StartedAt   *time.Time
	EndedAt     *time.Time
	Ended       bool
	BlueGoals   int
	RedGoals    int
}

// GetGame renders the game view page
func GetGame(c *gin.Context) {

	id := c.Param("id")

	var game Game

	if err := dbase.Find(&game, "id = ?", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		} else {
			SendError(http.StatusBadRequest, c, err)
			return
		}
	}

	var gameState CurrentGameState

	// Select game start/stop status
	var startEvent GameEvent

	if err := dbase.First(&startEvent, GameEvent{GameID: game.ID, EventType: "start"}).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			gameState.StartedAt = nil
		} else {
			SendError(http.StatusInternalServerError, c, err)
			return
		}
	} else {
		gameState.StartedAt = &startEvent.CreatedAt
		gameState.Started = true
	}

	var endEvent GameEvent

	if err := dbase.First(&endEvent, GameEvent{GameID: game.ID, EventType: "end"}).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			gameState.StartedAt = nil
		} else {
			SendError(http.StatusInternalServerError, c, err)
			return
		}
	} else {
		gameState.EndedAt = &endEvent.CreatedAt
		gameState.Ended = true
	}

	// Select user positions
	var rows []User

	if err := dbase.Raw(`SELECT users.*, position, team
														FROM game_events
															JOIN users ON game_events.user_id = users.id
														WHERE game_events.id IN (SELECT MAX(id)
																						          FROM game_events
																			                WHERE game_id = ? AND event_type = 'ptp'
																                      GROUP BY team, position)
														ORDER BY team, position ASC;`, id).Scan(&rows).Error; err != nil {
		panic(err)
	}

	gameState.BlueForward = rows[0]
	gameState.BlueGoalie = rows[1]
	gameState.RedForward = rows[2]
	gameState.RedGoalie = rows[3]

	// Select goals
	if gameState.Started {
		var goals TeamGoals

		if err := dbase.Raw(`SELECT (SELECT COUNT(id)
																	FROM game_events
																	WHERE game_id = ? 
																		AND event_type = 'goal' 
																		AND team = 'blue') as bluegoals,
															  (SELECT COUNT(id)
																	FROM game_events
																	WHERE game_id = ? 
																		AND event_type = 'goal' 
																		AND team = 'red') as redgoals;`, game.ID, game.ID).Scan(&goals).Error; err != nil {
			panic(err)
		}

		fmt.Println(goals)

		gameState.BlueGoals = goals.BlueGoals
		gameState.RedGoals = goals.RedGoals
	}

	SendHTML(http.StatusOK, c, "game", gin.H{
		"id":        id,
		"game":      game,
		"gameState": gameState,
	})
}

func MarkGoal(c *gin.Context) {
	gameID := c.Param("id")
	team := c.Query("team")
	position := c.Query("position")

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

	var scoreUser User

	if err := dbase.Raw(`SELECT users.*
												FROM game_events
													JOIN users ON game_events.user_id = users.id
												WHERE game_id = 2
												AND game_events.id = (SELECT MAX(id) FROM game_events WHERE position = ? AND team = ?)`, position, team).
		Scan(&scoreUser).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
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

func MarkStarted(c *gin.Context) {
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
		EventType: GameEventStart,
	}

	if err := dbase.Create(&event).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/game/%d", game.ID))
}

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

func ListGames(c *gin.Context) {

	var games []GameInfo

	if err := dbase.Raw(`
		SELECT g.*, (SELECT COUNT(id) 
									FROM game_events 
										WHERE game_id = g.id 
											AND event_type = 'goal' 
											AND team = 'blue') AS bluegoals, 
							(SELECT COUNT(id) 
									FROM game_events 
									WHERE game_id = g.id 
											AND event_type = 'goal' 
											AND team = 'red') AS redgoals
		FROM games AS g
	`).Scan(&games).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	SendHTML(http.StatusOK, c, "games", gin.H{
		"games": games,
	})
}
