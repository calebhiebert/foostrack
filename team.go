package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTeamForm returns the form required to create a team
func GetTeamForm(c *gin.Context) {

	var users []User

	if err := dbase.Find(&users).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	SendHTML(http.StatusOK, c, "teamform", gin.H{
		"users": users,
	})
}

// PostCreateTeam will accept input from the create team form and
// save a new team to the database
func PostCreateTeam(c *gin.Context) {
	name := c.PostForm("name")
	members := c.PostFormArray("members")

	// Create new transaction
	tx := dbase.Begin()

	// Create team
	team := Team{
		Name: name,
	}

	if err := tx.Create(&team).Error; err != nil {
		tx.Rollback()
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	// Create TeamUser entities for members
	for _, member := range members {
		teamUser := TeamUser{
			TeamID: team.ID,
			UserID: member,
		}

		if err := tx.Create(&teamUser).Error; err != nil {
			tx.Rollback()
			SendError(http.StatusInternalServerError, c, err)
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, "/teams")
}
