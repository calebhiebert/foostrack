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
	GameID    uint `gorm:"not null"`
	Game      Game `gorm:"association_foreignkey:GameID;"`
	UserID    *string
	User      User `gorm:"association_foreignkey:UserID;foreignkey:ID"`
	EventType string
	Team      string
	Position  string
	Elapsed   time.Duration `gorm:"-"`
}

// User represents the users table
type User struct {
	ID         string `gorm:"primary_key;unique_index"`
	Username   string
	PictureURL string
	Events     []GameEvent `gorm:"foreignkey:UserID"`
}

// Team represents the teams table
type Team struct {
	gorm.Model
	Name string
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
