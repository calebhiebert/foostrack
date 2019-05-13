package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func renderTeamForm(c *gin.Context, data gin.H) {
	var users []User

	if err := dbase.Find(&users).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	data["users"] = users

	SendHTML(http.StatusOK, c, "teamform", data)
}

// GetTeamForm returns the form required to create a team
func GetTeamForm(c *gin.Context) {
	renderTeamForm(c, gin.H{})
}

// PostCreateTeam will accept input from the create team form and
// save a new team to the database
func PostCreateTeam(c *gin.Context) {
	name := c.PostForm("name")
	members := c.PostFormArray("members")

	name = strings.TrimSpace(name)

	if len(members) != 2 {
		renderTeamForm(c, gin.H{
			"errors": []error{errors.New("A team must have 2 members")},
		})
		return
	}

	if name == "" {
		renderTeamForm(c, gin.H{
			"errors": []error{errors.New("A team cannot have an empty name")},
		})
		return
	}

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
