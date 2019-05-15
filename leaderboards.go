package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLeaderboardAvgGoalsPerGame(c *gin.Context) {

	var users []UserWithStats

	if err := dbase.Raw(`
	SELECT id, username, avg_goals_per_game
		FROM user_stats
		ORDER BY avg_goals_per_game DESC
		LIMIT 10;`).Scan(&users).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	SendHTML(http.StatusOK, c, "l-avggoalspergame", gin.H{
		"users": users,
		"board": "avggoalspergame",
		"inc":   RankFromIndex,
		"fmt":   fmt.Sprintf,
	})
}

func GetLeaderboardWinrate(c *gin.Context) {

	var users []UserWithStats

	if err := dbase.Raw(`
	SELECT id, username, (CAST(games_won AS DECIMAL) / games_played) * 100 AS winrate
		FROM user_stats
		ORDER BY winrate DESC
		LIMIT 10;`).Scan(&users).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	SendHTML(http.StatusOK, c, "l-winrate", gin.H{
		"users": users,
		"board": "winrate",
		"inc":   RankFromIndex,
		"fmt":   fmt.Sprintf,
	})
}

func RankFromIndex(i int) int {
	return i + 1
}
