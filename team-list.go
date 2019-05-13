package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTeamList returns a list of teams
func GetTeamList(c *gin.Context) {
	SendHTML(http.StatusOK, c, "teamlist", gin.H{})
}
