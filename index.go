package main

import (
	"net/http"

	prettyTime "github.com/andanhm/go-prettytime"
	"github.com/gin-gonic/gin"
)

// GetIndex renders the index page
func GetIndex(c *gin.Context) {
	var games []*GameInfo

	if err := dbase.Raw(`
		SELECT *
			FROM game_extended AS g
			WHERE g.start_time IS NOT NULL
				AND g.end_time IS NULL
			ORDER BY created_at DESC
			LIMIT 3;
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

	SendHTML(http.StatusOK, c, "index", gin.H{
		"games":          games,
		"formatTime":     prettyTime.Format,
		"hasActiveGames": len(games) > 0,
	})
}
