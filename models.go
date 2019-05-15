package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

/*
	This file contains models that directly map to datbase tables
*/

// Game represents the games table
type Game struct {
	gorm.Model
	Events   []GameEvent `gorm:"foreignkey:GameID"`
	WinGoals int         `gorm:"column:win_goals"`
}

// GameEvent represents the game_events table
type GameEvent struct {
	gorm.Model
	GameID    uint          `gorm:"not null" json:"gameId"`
	Game      Game          `gorm:"association_foreignkey:GameID;" json:"-"`
	UserID    *string       `json:"userId"`
	User      User          `gorm:"association_foreignkey:UserID;foreignkey:ID" json:"-"`
	EventType string        `json:"eventType"`
	Team      string        `json:"team"`
	Position  string        `json:"position"`
	Elapsed   time.Duration `gorm:"-" json:"elapsed"`
}

// User represents the users table
type User struct {
	ID         string      `gorm:"primary_key;unique_index" json:"id"`
	Username   string      `json:"username"`
	PictureURL string      `json:"pictureURL"`
	Events     []GameEvent `gorm:"foreignkey:UserID" json:"-"`
}

// Tournament represents the tournaments table
type Tournament struct {
	gorm.Model
	Name            string
	CreatedByID     string
	Status          string
	User            User             `gorm:"association_foreignkey:CreatedByID;foreignkey:ID"`
	TournamentUsers []TournamentUser `gorm:"foreignkey:TeamID"`
}

// Team represents the teams table
type Team struct {
	gorm.Model
	Name    string
	Members []TournamentUser `gorm:"foreignkey:TeamID"`
}

// TournamentUser represents the tournament_users table
type TournamentUser struct {
	TournamentID uint       `gorm:"primary_key" json:"tournamentId"`
	Tournament   Tournament `gorm:"association_foreignkey:TournamentID;foreignkey:ID"`
	TeamID       *uint      `json:"teamId"`
	Team         Team       `gorm:"association_foreignkey:TeamID"`
	UserID       string     `gorm:"primary_key"`
	User         User       `gorm:"association_foreignkey:UserID;foreignkey:ID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `sql:"index"`
}
