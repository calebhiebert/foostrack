package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTeamList returns a list of teams
func GetTeamList(c *gin.Context) {

	var teams []Team

	if err := dbase.Preload("TeamUsers").Preload("TeamUsers.User").Find(&teams).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	SendHTML(http.StatusOK, c, "teamlist", gin.H{
		"teams": teams,
	})
}
