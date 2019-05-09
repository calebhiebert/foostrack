package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type User struct {
	ID         string `gorm:"primary_key;unique_index"`
	Username   string
	PictureURL string
	Events     []GameEvent `gorm:"foreignkey:UserID"`
}

type UserWithPosition struct {
	ID         string
	Username   string
	PictureURL string
	Team       string
	Position   string
}

// GetUser will render a user page
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

	fmt.Println(user)

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
	})
}
