package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/lucasb-eyer/go-colorful"
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
		Status:      TournamentStatusSignup,
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

	if err := dbase.Preload("TournamentUsers.User").Preload("User").Preload("Teams.Members.User").First(&tournament, "id = ?", id).Error; err != nil {
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
	canMakeTeams := tournament.Status == TournamentStatusSignup && len(tournament.TournamentUsers)%2 == 0

	teams := make([]map[string]interface{}, 0)

	for _, t := range tournament.Teams {
		tm := make(map[string]interface{})

		teams = append(teams, tm)

		tm["id"] = t.ID
		tm["color"] = t.Color
		tm["name"] = t.Name
	}

	var bracketPositions []*BracketPosition

	if err := dbase.Raw(`SELECT * FROM bracket_positions
												WHERE tournament_id = ?
												ORDER BY bracket_level ASC, bracket_position ASC;`, tournament.ID).Scan(&bracketPositions).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	for _, bp := range bracketPositions {
		for _, t := range tournament.Teams {
			if t.ID == bp.TeamID {
				bp.Team = t
			}
		}
	}

	SendHTML(http.StatusOK, c, "tournament", gin.H{
		"tournament":             tournament,
		"isUserJoinedTournament": isUserJoinedTournament,
		"isManager":              isTournamentManager,
		"canMakeTeams":           canMakeTeams,
		"unevenParticipants":     len(tournament.TournamentUsers)%2 != 0,
		"teams":                  teams,
		"canEditTeam": func(t Team) bool {
			for _, tm := range t.Members {
				if tm.UserID == userID {
					return true
				}
			}

			return tournament.CreatedByID == userID
		},
		"bracketPositions": bracketPositions,
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

func NukeTournament(c *gin.Context) {
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

	if tournament.CreatedByID != userID {
		SendForbid(c, "Only tournament managers can delete tournaments")
		return
	}

	if err := dbase.Unscoped().Where("tournament_id = ?", tournament.ID).Delete(&BracketPosition{}).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	if err := dbase.Unscoped().Where("tournament_id = ?", tournament.ID).Delete(&TournamentUser{}).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	if err := dbase.Unscoped().Where("tournament_id = ?", tournament.ID).Delete(&Team{}).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	if err := dbase.Unscoped().Where("id = ?", tournament.ID).Delete(&Tournament{}).Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, "/tournaments")
}

// CreateTeams will create teams for a tournament
func CreateTeams(c *gin.Context) {
	if !EnsureLoggedIn(c) {
		return
	}

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

	if tournament.CreatedByID != userID {
		SendForbid(c, "Only tournament managers can create teams")
		return
	}

	bracketPosition := 0
	teams := make([]*Team, 0)

	// Create Teams
	tx := dbase.Begin()

	for i := 0; i < len(tournament.TournamentUsers); i += 2 {
		u1 := &tournament.TournamentUsers[i]
		u2 := &tournament.TournamentUsers[i+1]

		team := &Team{
			TournamentID: tournament.ID,
			Name:         fmt.Sprintf("Team %d", i/2+1),
			Color:        colorful.HappyColor().Hex(),
		}

		if err := tx.Create(&team).Error; err != nil {
			tx.Rollback()
			SendError(http.StatusInternalServerError, c, err)
			return
		}

		u1.TeamID = &team.ID
		u2.TeamID = &team.ID

		if err := tx.Save(u1).Error; err != nil {
			tx.Rollback()
			SendError(http.StatusInternalServerError, c, err)
			return
		}

		if err := tx.Save(u2).Error; err != nil {
			tx.Rollback()
			SendError(http.StatusInternalServerError, c, err)
			return
		}

		team.Members = make([]TournamentUser, 2)
		team.Members[0] = *u1
		team.Members[1] = *u2

		teams = append(teams, team)
	}

	for i := 0; i < (len(teams) - (len(teams) % 2)); i += 2 {
		t1 := teams[i]
		t2 := teams[i+1]

		game, err := createGame(t1.Members[0].UserID, t1.Members[1].UserID, t2.Members[0].UserID, t2.Members[0].UserID, 10)
		if err != nil {
			tx.Rollback()
			SendError(http.StatusInternalServerError, c, err)
			return
		}

		b1 := BracketPosition{
			TournamentID:    tournament.ID,
			TeamID:          t1.ID,
			GameID:          &game.ID,
			BracketLevel:    0,
			BracketPosition: bracketPosition,
		}

		if err := tx.Create(&b1).Error; err != nil {
			tx.Rollback()
			SendError(http.StatusInternalServerError, c, err)
			return
		}

		bracketPosition++

		b2 := BracketPosition{
			TournamentID:    tournament.ID,
			TeamID:          t2.ID,
			GameID:          &game.ID,
			BracketLevel:    0,
			BracketPosition: bracketPosition,
		}

		if err := tx.Create(&b2).Error; err != nil {
			tx.Rollback()
			SendError(http.StatusInternalServerError, c, err)
			return
		}

		bracketPosition++
	}

	// There is an odd team
	if len(teams)%2 != 0 {
		b1 := BracketPosition{
			TournamentID:    tournament.ID,
			TeamID:          teams[len(teams)-1].ID,
			BracketLevel:    0,
			BracketPosition: bracketPosition,
		}

		if err := tx.Create(&b1).Error; err != nil {
			tx.Rollback()
			SendError(http.StatusInternalServerError, c, err)
			return
		}

		b2 := BracketPosition{
			TournamentID:    tournament.ID,
			TeamID:          teams[len(teams)-1].ID,
			BracketLevel:    1,
			BracketPosition: 0,
		}

		if err := tx.Create(&b2).Error; err != nil {
			tx.Rollback()
			SendError(http.StatusInternalServerError, c, err)
			return
		}

		bracketPosition++
	}

	tournament.Status = TournamentStatusUnderway

	if err := tx.Save(&tournament).Error; err != nil {
		tx.Rollback()
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	// Commit Teams
	if err := tx.Commit().Error; err != nil {
		SendError(http.StatusInternalServerError, c, err)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/tournament/%d", tournament.ID))
}
