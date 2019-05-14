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

// Team represents the teams table
type Team struct {
	gorm.Model
	Name      string
	TeamUsers []TeamUser `gorm:"foreignkey:TeamID"`
}

// TeamUser represents the team_user table
type TeamUser struct {
	TeamID    uint   `gorm:"primary_key"`
	Team      Team   `gorm:"association_foreignkey:TeamID"`
	UserID    string `gorm:"primary_key"`
	User      User   `gorm:"association_foreignkey:UserID;foreignkey:ID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
