package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/kvartborg/vector"
	"github.com/solarlune/tetra3d"
)

//go:embed map01.gltf
var gltfData []byte

const ScreenWidth = 796
const ScreenHeight = 448

type Game struct {
	GameScene                *tetra3d.Scene
	Camera                   *tetra3d.Camera
	PrevMousePosition        vector.Vector
	CameraTilt, CameraRotate float64
	DrawDebugText            bool
	DrawDebugDepth           bool
	DrawDebugWireframe       bool
	DrawDebugNormals         bool
	Library                  *tetra3d.Library
	Gun, fouce, fire         *ebiten.Image
	Audio                    *audio.Player
	IsFire, IsFullscreen     bool
}

func NewGame() *Game {

	g := &Game{PrevMousePosition: vector.Vector{}}
	//加载图片
	m, _ := os.ReadFile("gun2.png")
	i, _, _ := image.Decode(bytes.NewReader(m))
	g.Gun = ebiten.NewImageFromImage(i)
	m, _ = os.ReadFile("crosshair.png")
	i, _, _ = image.Decode(bytes.NewReader(m))
	g.fouce = ebiten.NewImageFromImage(i)
	m, _ = os.ReadFile("fire.png")
	i, _, _ = image.Decode(bytes.NewReader(m))
	g.fire = ebiten.NewImageFromImage(i)
	//加载模型
	d, err := tetra3d.LoadGLTFData(gltfData, nil)
	g.Library = d
	g.GameScene = g.Library.ExportedScene
	//声音
	bgm, _ := os.ReadFile("gun.mp3")
	ss, _ := mp3.DecodeWithSampleRate(44100, bytes.NewReader(bgm))
	cont := audio.NewContext(44100)
	g.Audio, _ = cont.NewPlayer(ss)
	g.Audio.SetVolume(0)
	g.Audio.Play()

	if err != nil {
		panic(err)
	}

	g.Camera = tetra3d.NewCamera(ScreenWidth, ScreenHeight)

	g.Camera.SetLocalPosition(vector.Vector{0, 1, 15})

	g.GameScene.Root.AddChildren(g.Camera)
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	fmt.Println(g.GameScene.Root.HierarchyAsString())
	return g
}

func (g *Game) Update() error {
	var err error

	spin := g.GameScene.Root.Get("Spin").(*tetra3d.Model)
	spin.Rotate(0, 1, 0, 0.03)
	// light := g.GameScene.Root.Get("Point light").(*tetra3d.Node)
	// light.AnimationPlayer().Play(g.Library.Animations["LightAction"])
	// light.AnimationPlayer().Update(1.0 / 60.0)

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		err = errors.New("quit")
	}
	moveSpd := 0.1
	forward := g.Camera.LocalRotation().Forward().Invert()
	right := g.Camera.LocalRotation().Right()
	pos := g.Camera.LocalPosition()
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		pos = pos.Add(forward.Scale(moveSpd))
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		pos = pos.Add(right.Scale(moveSpd))
	}

	if ebiten.IsKeyPressed(ebiten.KeyS) {
		pos = pos.Add(forward.Scale(-moveSpd))
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		pos = pos.Add(right.Scale(-moveSpd))
	}

	// if ebiten.IsKeyPressed(ebiten.KeySpace) {
	// 	pos[1] += moveSpd
	// }
	// if ebiten.IsKeyPressed(ebiten.KeyControl) {
	// 	pos[1] -= moveSpd
	// }
	//开火
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if !g.IsFire {
			g.Audio.SetVolume(2)
			g.IsFire = true
			go func() {
				g.Audio.Rewind()
				g.Audio.Play()
			}()
		}
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.IsFire = false
	}
	//全屏
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		ebiten.SetFullscreen(!g.IsFullscreen)
	}

	g.Camera.SetLocalPosition(pos)

	mx, my := ebiten.CursorPosition()
	mv := vector.Vector{float64(mx), float64(my)}
	diff := mv.Sub(g.PrevMousePosition)
	g.CameraRotate -= diff[0] * 0.005
	rotate := tetra3d.NewMatrix4Rotate(0, 1, 0, g.CameraRotate)
	g.Camera.SetLocalRotation(rotate)
	g.PrevMousePosition = mv.Clone()
	return err
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	g.Camera.Clear()

	g.Camera.RenderNodes(g.GameScene, g.GameScene.Root)
	screen.DrawImage(g.Camera.ColorTexture(), nil)

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Translate(338, 270)
	screen.DrawImage(g.Gun, op)

	op1 := &ebiten.DrawImageOptions{}
	op1.Filter = ebiten.FilterLinear
	op1.GeoM.Translate(338, 224)
	screen.DrawImage(g.fouce, op1)

	if g.IsFire {
		op2 := &ebiten.DrawImageOptions{}
		op2.Filter = ebiten.FilterLinear
		op2.GeoM.Translate(330, 224)
		screen.DrawImage(g.fire, op2)
	}

	ebitenutil.DebugPrint(screen, strconv.FormatFloat(g.Camera.LocalPosition()[0], 'f', 0, 64)+"\n"+strconv.FormatFloat(g.Camera.LocalPosition()[1], 'f', 0, 64)+"\n"+strconv.FormatFloat(g.Camera.LocalPosition()[2], 'f', 0, 64))

}

func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	os.Setenv("EBITEN_GRAPHICS_LIBRARY", "opengl")
	game := NewGame()
	ebiten.SetWindowTitle("golang FPS")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}

}
