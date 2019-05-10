package main

/*
	This file contains some predefined application constants
	Enums in go are weird, do some googling, this is how I chose to do them
*/

// GameEvent constants
const (
	GameEventStart              = "start"
	GameEventEnd                = "end"
	GameEventPlayerTakePosition = "ptp"
	GameEventGoal               = "goal"
	GameEventAntiGoal           = "antigoal"
	GameEventDeadBall           = "dead"
	GameEventOutOfBounds        = "oob"
)

// Game team constants
const (
	GameTeamRed  = "red"
	GameTeamBlue = "blue"
)

// Game positions
const (
	GamePositionForward = "forward"
	GamePositionGoalie  = "goalie"
)
