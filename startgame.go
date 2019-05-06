package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetStartGame(c *gin.Context) {
	var users []User

	if err := dbase.Find(&users).Error; err != nil {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"error":   err,
			"general": c.GetStringMapString("general"),
		})
		return
	}

	c.HTML(http.StatusOK, "startgame", gin.H{
		"users":   users,
		"general": c.GetStringMapString("general"),
	})
}

func PostStartGame(c *gin.Context) {
	fmt.Println("Stuff", c.PostForm("blue_goalie"))

	game := Game{}

	if err := dbase.Create(&game).Error; err != nil {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"error":   err,
			"general": c.GetStringMapString("general"),
		})
		return
	}

	c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/game/%d", game.ID))
}
