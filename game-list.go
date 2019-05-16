package main

/*
	This file contains the gin handler for rendering the games list page
*/

import (
	"math"
	"net/http"
	"strconv"

	prettyTime "github.com/andanhm/go-prettytime"
	"github.com/gin-gonic/gin"
)

// ListGames lists out a page with all games
func ListGames(c *gin.Context) {

	limit := 20
	offset := 0

	pageQuery := c.Query("page")

	parsedPageQuery, err := strconv.Atoi(pageQuery)
	if err != nil || parsedPageQuery < 1 {
		// page query param is not valid
	} else {
		offset = (parsedPageQuery - 1) * limit
	}

	currentPage := offset / limit

	var games []*GameExtended

	if err := dbase.Raw(`
		SELECT *
		FROM game_extended AS g
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset).Scan(&games).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	var count Count

	if err := dbase.Raw(`SELECT COUNT(id) AS count FROM games`).Scan(&count).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	totalPages := int(math.Ceil(float64(count.Count) / float64(limit)))

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

	teamIds := make([]uint, 0)

	for _, game := range games {
		if game.RedTeamID != nil {
			teamIds = append(teamIds, *game.RedTeamID)
		}

		if game.BlueTeamID != nil {
			teamIds = append(teamIds, *game.BlueTeamID)
		}
	}

	var teams []Team

	if err := dbase.Where(teamIds).Find(&teams).Error; err != nil {
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

		for _, team := range teams {
			if *gi.BlueTeamID == team.ID {
				gi.BlueTeam = team
			}

			if *gi.RedTeamID == team.ID {
				gi.RedTeam = team
			}
		}

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

	pageArr := make([]int, totalPages)

	for i := 0; i < totalPages; i++ {
		pageArr[i] = i + 1
	}

	SendHTML(http.StatusOK, c, "games", gin.H{
		"games":      games,
		"formatTime": prettyTime.Format,
		"page":       currentPage + 1,
		"totalPages": pageArr,
		"exfname":    ExtractFirstName,
	})
}
