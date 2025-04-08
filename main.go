package main

import (
	"github.com/AgDecode/mini-game/game"
)

func main() {
}

var gameState *game.State

func initGame() {
	gameState = game.InitGame()
}

func handleCommand(command string) string {
	return gameState.HandleCommand(command)
}
