package main

import (
	"errors"
	"github.com/jinzhu/gorm"
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
	data["isMember"] = func(u User) bool {
		members := data["members"].([]TeamUser)

		for _, m := range members {
			if u.ID == m.UserID {
				return true
			}
		}

		return false
	}

	SendHTML(http.StatusOK, c, "teamform", data)
}

// GetTeamForm returns the form required to create a team
func GetTeamForm(c *gin.Context) {
	renderTeamForm(c, gin.H{})
}

// GetTeamEditForm returns a form to edit a team
func GetTeamEditForm(c *gin.Context) {
	var team Team

	if err := dbase.First(&team, "id = ?", c.Param("id")).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		} else {
			SendError(http.StatusInternalServerError, c, err)
			return
		}
	}

	var members []TeamUser

	if err := dbase.Find(&members, "team_id = ?", team.ID).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	renderTeamForm(c, gin.H{
		"editing": true,
		"team":    team,
		"members": members,
	})
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
