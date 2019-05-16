package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// GetTournamentList returns the tournament list page
func GetTournamentList(c *gin.Context) {
	var tournaments []Tournament

	if err := dbase.Find(&tournaments).Order("id DESC").Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	SendHTML(http.StatusOK, c, "tournaments", gin.H{
		"tournaments": tournaments,
	})
}

// GetTournamentForm returns the tournament creation form
func GetTournamentForm(c *gin.Context) {
	if !EnsureLoggedIn(c) {
		return
	}

	SendHTML(http.StatusOK, c, "tournamentform", gin.H{})
}

// PostTournamentForm captures the input from the create tournament form
func PostTournamentForm(c *gin.Context) {
	if !EnsureLoggedIn(c) {
		return
	}

	general := c.GetStringMapString("general")
	name := c.PostForm("name")

	tournament := Tournament{
		Name:        name,
		CreatedByID: general["user_id"],
	}

	if err := dbase.Create(&tournament).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/tournament/%d", tournament.ID))
}

// GetTournament returns the tournament landing page
func GetTournament(c *gin.Context) {
	id := c.Param("id")

	var tournament Tournament

	if err := dbase.Preload("TournamentUsers.User").Preload("User").First(&tournament, "id = ?", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	general := c.GetStringMapString("general")
	userID := general["user_id"]
	isUserJoinedTournament := false

	for _, tu := range tournament.TournamentUsers {
		if tu.UserID == userID {
			isUserJoinedTournament = true
		}
	}

	isTournamentManager := tournament.CreatedByID == userID

	SendHTML(http.StatusOK, c, "tournament", gin.H{
		"tournament":             tournament,
		"isUserJoinedTournament": isUserJoinedTournament,
		"isManager":              isTournamentManager,
	})
}

// PostJoinTournament will join a user to a tournament
func PostJoinTournament(c *gin.Context) {
	if !EnsureLoggedIn(c) {
		return
	}

	id := c.Param("id")
	general := c.GetStringMapString("general")
	userID := general["user_id"]

	var tournament Tournament

	if err := dbase.Preload("User").First(&tournament, "id = ?", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	tournamentUser := TournamentUser{
		TournamentID: tournament.ID,
		UserID:       userID,
	}

	if err := dbase.Create(&tournamentUser).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/tournament/%d", tournament.ID))
}

func GetTournamentUserSelect(c *gin.Context) {

	if !EnsureLoggedIn(c) {
		return
	}

	id := c.Param("id")

	var tournament Tournament

	if err := dbase.First(&tournament, "id = ?", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	general := c.GetStringMapString("general")

	if general["user_id"] != tournament.CreatedByID {
		SendForbid(c, "Only tournament managers can add users")
		return
	}

	RenderUserSelect(c, UserSelectTournament, fmt.Sprintf("Pick a user for %s", tournament.Name), func(u User) string {
		return fmt.Sprintf("/tournament/%d/adduser/%s", tournament.ID, u.ID)
	})
}

func AddUserToTournament(c *gin.Context) {
	tid := c.Param("id")
	uid := c.Param("uid")

	if !EnsureLoggedIn(c) {
		return
	}

	var tournament Tournament

	if err := dbase.First(&tournament, "id = ?", tid).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			SendNotFound(c)
			return
		}

		SendError(http.StatusInternalServerError, c, err)
		return
	}

	general := c.GetStringMapString("general")

	if general["user_id"] != tournament.CreatedByID {
		SendForbid(c, "Only tournament managers can add users")
		return
	}

	tUser := TournamentUser{
		TournamentID: tournament.ID,
		UserID:       uid,
	}

	if err := dbase.Create(&tUser).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/tournament/%d", tournament.ID))
}
