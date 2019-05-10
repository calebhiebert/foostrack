package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// GetUser will render a single user's page
func GetUser(c *gin.Context) {
	id := c.Param("id")

	var user UserWithStats

	if err := dbase.Raw(`SELECT * FROM user_stats WHERE id = ?`, id).Scan(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	winrate := (float64(user.GamesWon) / float64(user.GamesPlayed)) * 100
	redPercent := (float64(user.GamesPlayedRed) / float64(user.GamesPlayed)) * 100
	bluePercent := (float64(user.GamesPlayedBlue) / float64(user.GamesPlayed)) * 100

	SendHTML(http.StatusOK, c, "user", gin.H{
		"user":               user,
		"winpercent":         fmt.Sprintf("%.0f", winrate),
		"games_played":       user.GamesPlayed,
		"red_games_percent":  fmt.Sprintf("%.0f", redPercent),
		"blue_games_percent": fmt.Sprintf("%.0f", bluePercent),
		"avg_goals_per_game": fmt.Sprintf("%.1f", user.AverageGoalsPerGame),
		"antigoals":          user.AntiGoals,
		"goals":              user.Goals,
	})
}
