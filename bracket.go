package main

import (
	"fmt"
	"sort"
	"strings"
)

// CheckBracket checks brackets and creates the required games and whatnot
func CheckBracket(tournamentID uint) error {
	var bracketPositions []BracketPosition

	if err := dbase.Find(&bracketPositions, "tournament_id = ?", tournamentID).Order("bracket_level ASC").Error; err != nil {
		return err
	}

	bracketLevels := make([][]BracketPosition, 0)

	for _, bp := range bracketPositions {
		if len(bracketLevels) < bp.BracketLevel+1 {
			bracketLevels = append(bracketLevels, make([]BracketPosition, 0))
		}

		bracketLevels[bp.BracketLevel] = append(bracketLevels[bp.BracketLevel], bp)
	}

	// For each bracket level
	for i := 0; i < len(bracketLevels); i++ {

		// Sort the new bracket level based on the team's position in the previous bracket
		// This is done to make sure the bracket progresses in the correct order
		if i > 0 {
			sort.Slice(bracketLevels[i], func(a, b int) bool {
				var aBracketPosition int
				var bBracketPosition int

				for _, pb := range bracketLevels[i-1] {
					if pb.TeamID == bracketLevels[i][a].TeamID {
						aBracketPosition = pb.BracketPosition
					}
				}

				for _, pb := range bracketLevels[i-1] {
					if pb.TeamID == bracketLevels[i][b].TeamID {
						bBracketPosition = pb.BracketPosition
					}
				}

				return aBracketPosition < bBracketPosition
			})
		}

		// For each bracket position
		for j := 0; j < len(bracketLevels[i]); j++ {
			bl := bracketLevels[i][j]

			if bl.BracketPosition != j {
				bracketLevels[i][j].BracketPosition = j

				if err := dbase.Exec(`UPDATE bracket_positions 
														SET bracket_position = ?
														WHERE tournament_id = ? AND team_id = ? AND bracket_level = ?`, j, bl.TournamentID, bl.TeamID, bl.BracketLevel).Error; err != nil {
					return err
				}
			}
		}

		// Create games for matchups where games have not been created yet
		for j := 0; j < len(bracketLevels[i])-(len(bracketLevels[i])%2); j += 2 {
			b1 := bracketLevels[i][j]
			b2 := bracketLevels[i][j+1]

			if b1.GameID == nil && b2.GameID == nil {
				var team1 Team

				if err := dbase.Preload("Members").Find(&team1, "id = ?", b1.TeamID).Error; err != nil {
					return err
				}

				var team2 Team

				if err := dbase.Preload("Members").Find(&team2, "id = ?", b2.TeamID).Error; err != nil {
					return err
				}

				tx := dbase.Begin()

				game, err := createGame(team1.Members[0].UserID, team1.Members[1].UserID, team2.Members[0].UserID, team2.Members[1].UserID, 10, &team1.ID, &team2.ID, tx)
				if err != nil {
					return err
				}

				if err := tx.Exec(`UPDATE bracket_positions 
														SET game_id = ?
														WHERE tournament_id = ? 
															AND team_id = ? 
															AND bracket_level = ?
															AND bracket_position = ?`,
					game.ID, b1.TournamentID, b1.TeamID, b1.BracketLevel, b1.BracketPosition).Error; err != nil {
					tx.Rollback()
					return err
				}

				if err := tx.Exec(`UPDATE bracket_positions 
														SET game_id = ?
														WHERE tournament_id = ? 
															AND team_id = ? 
															AND bracket_level = ?
															AND bracket_position = ?`,
					game.ID, b2.TournamentID, b2.TeamID, b2.BracketLevel, b2.BracketPosition).Error; err != nil {
					tx.Rollback()
					return err
				}

				if err := tx.Commit().Error; err != nil {
					return err
				}
			}
		}

		if len(bracketLevels[i])%2 != 0 && len(bracketLevels[i]) > 1 {
			bl := bracketLevels[i][len(bracketLevels[i])-1]

			// Find the next highest bracket position
			var count Count

			if err := dbase.Raw(`SELECT MAX(bracket_level) AS count
													FROM bracket_positions
													WHERE tournament_id = ?
														AND bracket_level = ?`, bl.TournamentID, bl.BracketLevel+1).Scan(&count).Error; err != nil {
				return err
			}

			newBracketPosition := BracketPosition{
				TournamentID:    bl.TournamentID,
				TeamID:          bl.TeamID,
				BracketLevel:    i + 1,
				BracketPosition: count.Count + 1,
			}

			if err := dbase.Create(&newBracketPosition).Error; err != nil {
				if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
					fmt.Println("Tried to create duplicate bracket position", newBracketPosition)
				} else {
					return err
				}
			}
		}
	}

	return nil
}
