package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type TeamGoals struct {
	RedGoals  int `gorm:"column:redgoals"`
	BlueGoals int `gorm:"column:bluegoals"`
}

type GameInfo struct {
	gorm.Model
	TeamGoals
}

func SendHTML(statusCode int, c *gin.Context, page string, data gin.H) {

	if data == nil {
		data = gin.H{}
	}

	data["general"] = c.GetStringMapString("general")

	c.HTML(statusCode, page, data)
}

func SendError(code int, c *gin.Context, err error) {
	SendHTML(code, c, "error", gin.H{
		"error": err,
	})
}

func SendNotFound(c *gin.Context) {
	SendHTML(http.StatusNotFound, c, "notfound", nil)
}

func EnsureLoggedIn(c *gin.Context) bool {
	general := c.GetStringMapString("general")

	if general["isloggedin"] != "true" {
		SendHTML(http.StatusForbidden, c, "blocked", nil)
		return false
	}

	return true
}
