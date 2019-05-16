package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetTeamForm(c *gin.Context) {
	if !EnsureLoggedIn(c) {
		return
	}

	id := c.Param("id")

	var team Team

	if err := dbase.Preload("Members").Preload("Tournament").First(&team, "id = ?", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	// Check team edit permissions
	general := c.GetStringMapString("general")
	userID := general["user_id"]

	hasEditPermissions := false

	for _, mbr := range team.Members {
		if mbr.UserID == userID {
			hasEditPermissions = true
		}
	}

	if team.Tournament.CreatedByID == userID {
		hasEditPermissions = true
	}

	if !hasEditPermissions {
		SendForbid(c, "Only team members or tournament managers can edit teams")
		return
	}

	SendHTML(http.StatusOK, c, "teamform", gin.H{
		"team": team,
	})
}

func PostEditTeam(c *gin.Context) {
	if !EnsureLoggedIn(c) {
		return
	}

	id := c.Param("id")

	var team Team

	if err := dbase.Preload("Members").Preload("Tournament").First(&team, "id = ?", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	// Check team edit permissions
	general := c.GetStringMapString("general")
	userID := general["user_id"]

	hasEditPermissions := false

	for _, mbr := range team.Members {
		if mbr.UserID == userID {
			hasEditPermissions = true
		}
	}

	if team.Tournament.CreatedByID == userID {
		hasEditPermissions = true
	}

	if !hasEditPermissions {
		SendForbid(c, "Only team members or tournament managers can edit teams")
		return
	}

	name := c.PostForm("name")
	color := c.PostForm("color")

	team.Name = name
	team.Color = color

	if err := dbase.Save(&team).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/tournament/%d", team.TournamentID))
}
