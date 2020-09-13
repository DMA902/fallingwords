package main

import (
	"log"

	"personal_projects/ebiten/typing/pkg/game"

	"github.com/hajimehoshi/ebiten"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Falling Words")

	theGame, err := game.NewGame(screenWidth, screenHeight)
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(&theGame); err != nil {
		log.Fatal(err)
	}
}
