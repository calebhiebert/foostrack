package main

import (
	"fmt"
	"net/http"
	"time"

	prettyTime "github.com/andanhm/go-prettytime"
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
	User      User `gorm:"association_foreignkey:UserID;foreignkey:ID"`
	EventType string
	Team      string
	Position  string
}

// CurrentGameState represents the current state of a single game
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
		}

		SendError(http.StatusBadRequest, c, err)
		return
	}

	var gameState CurrentGameState

	// Select Event List
	var events []GameEvent

	if err := dbase.Preload("User").Find(&events, "game_id = ?", game.ID).Error; err != nil {
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

	SendHTML(http.StatusOK, c, "game", gin.H{
		"id":        id,
		"game":      game,
		"gameState": gameState,
		"events":    events,
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

// ListGames lists out a page with all games
func ListGames(c *gin.Context) {

	var games []*GameInfo

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
											AND team = 'red') AS redgoals,
							(SELECT created_at FROM game_events ge WHERE ge.game_id = g.id AND ge.event_type = 'start') AS start_time,
							(SELECT created_at FROM game_events ge WHERE ge.game_id = g.id AND ge.event_type = 'end') AS end_time
		FROM games AS g
		ORDER BY created_at DESC
	`).Scan(&games).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	gameIds := make([]uint, 0)

	for _, game := range games {
		gameIds = append(gameIds, game.ID)
	}

	var currentPositions []GameEvent

	if err := dbase.Raw(`SELECT *
  										   FROM current_positions
												 WHERE game_id IN (?);`, gameIds).
		Scan(&currentPositions).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	userIds := make([]string, 0)

	for _, pos := range currentPositions {
		userIds = append(userIds, *pos.UserID)
	}

	var users []User

	if err := dbase.Raw(`SELECT * FROM users WHERE id IN (?)`, userIds).
		Scan(&users).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	userMap := make(map[string]User)

	for _, usr := range users {
		userMap[usr.ID] = usr
	}

	for _, gi := range games {
		gi.Started = gi.StartTime != nil
		gi.Ended = gi.EndTime != nil

		for _, pos := range currentPositions {
			if pos.GameID == gi.ID {
				switch pos.Team {

				// Blue Team
				case GameTeamBlue:
					switch pos.Position {
					// Forward
					case GamePositionForward:
						gi.BlueForward = userMap[*pos.UserID]

						// Goalie
					case GamePositionGoalie:
						gi.BlueGoalie = userMap[*pos.UserID]
					}
				case GameTeamRed:
					switch pos.Position {
					// Forward
					case GamePositionForward:
						gi.RedForward = userMap[*pos.UserID]

						// Goalie
					case GamePositionGoalie:
						gi.RedGoalie = userMap[*pos.UserID]
					}
				}
			}
		}
	}

	SendHTML(http.StatusOK, c, "games", gin.H{
		"games":      games,
		"formatTime": prettyTime.Format,
	})
}
