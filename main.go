package main

import (
	"fps/engine"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	os.Setenv("EBITEN_GRAPHICS_LIBRARY", "opengl")
	game := engine.NewGame()
	ebiten.SetWindowTitle("golang FPS")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}

}
