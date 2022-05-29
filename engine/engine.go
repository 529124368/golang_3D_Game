package engine

import (
	"bytes"
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/kvartborg/vector"
	"github.com/solarlune/tetra3d"
)

//go:embed asset
var Asset embed.FS

const ScreenWidth = 796
const ScreenHeight = 448

type Enemy struct {
	Name   string
	HP     int
	shoted bool
}

func (e *Enemy) DelHP(img *ebiten.Image) {
	if e.HP > 0 {
		e.HP -= 20
	}
}

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
	Gun, fouce, fire, HP     *ebiten.Image
	Audio                    *audio.Player
	IsFire                   bool
	firePos                  vector.Vector
	enemyList                []*Enemy
}

func (g *Game) AddDelEm() {
	for _, v := range g.enemyList {
		if !g.GameScene.Root.Get(v.Name).Visible() {
			rand.Seed(time.Now().UnixNano())
			randomNum := rand.Intn(3)
			randomNum1 := rand.Intn(10)
			g.GameScene.Root.Get(v.Name).ResetLocalTransform()
			if v.Name == "en1" {
				g.GameScene.Root.Get(v.Name).SetLocalPosition(vector.Vector{21 + float64(randomNum1), 1, -9 - float64(randomNum)})
			} else {
				g.GameScene.Root.Get(v.Name).SetLocalPosition(vector.Vector{20 + float64(randomNum1), 1, 2 - float64(randomNum)})
			}

			g.GameScene.Root.Get(v.Name).SetVisible(true, false)
		}
	}
}

func GetAssetBytes(name string) []byte {
	b, err := Asset.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return b
}

func NewGame() *Game {

	g := &Game{PrevMousePosition: vector.Vector{}, firePos: vector.Vector{338, 184}}
	//加载图片
	m := GetAssetBytes("asset/gun2.png")
	i, _, _ := image.Decode(bytes.NewReader(m))
	g.Gun = ebiten.NewImageFromImage(i)
	m = GetAssetBytes("asset/crosshair.png")
	i, _, _ = image.Decode(bytes.NewReader(m))
	g.fouce = ebiten.NewImageFromImage(i)
	m = GetAssetBytes("asset/fire.png")
	i, _, _ = image.Decode(bytes.NewReader(m))
	g.fire = ebiten.NewImageFromImage(i)
	//HP
	m = GetAssetBytes("asset/hp_bar.png")
	i, _, _ = image.Decode(bytes.NewReader(m))
	g.HP = ebiten.NewImageFromImage(i)
	//敌人
	m = GetAssetBytes("asset/enmey.png")
	i, _, _ = image.Decode(bytes.NewReader(m))
	enemy := ebiten.NewImageFromImage(i)
	cubeMesh := tetra3d.NewPlane()
	mat := cubeMesh.MeshParts[0].Material
	mat.Shadeless = true
	mat.Texture = enemy
	parent := tetra3d.NewModel(cubeMesh, "en1")
	parent.Rotate(1, 0, 0, -1.6)
	parent.SetLocalPosition(vector.Vector{21, 1, -9})
	p1 := parent.Clone()
	p1.SetName("en2")
	p1.SetLocalPosition(vector.Vector{20, 1, 2})
	//敌人信息入库
	g.enemyList = append(g.enemyList, &Enemy{Name: "en1", HP: 48})
	g.enemyList = append(g.enemyList, &Enemy{Name: "en2", HP: 48})
	//加载模型
	d, err := tetra3d.LoadGLTFData(GetAssetBytes("asset/map01.gltf"), nil)
	g.Library = d
	g.GameScene = g.Library.ExportedScene
	//声音
	bgm := GetAssetBytes("asset/gun.mp3")
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

	g.GameScene.Root.AddChildren(g.Camera, parent, p1)

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
	//开火
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for _, v := range g.enemyList {
			if g.GameScene.Root.Get(v.Name).Visible() {
				s := g.GetScreenPos(v.Name)
				if s[0] >= 0 && s[0] <= ScreenWidth && s[1] >= 0 && s[1] <= ScreenHeight {
					if s.Sub(g.firePos).Magnitude() < 50 {
						fmt.Println("击中" + v.Name)
						v.DelHP(g.HP)
						v.shoted = true
					} else {
						fmt.Println("没击中")
						v.shoted = false
					}
				}
			} else {
				fmt.Println("没击中")
			}
		}

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
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	g.Camera.SetLocalPosition(pos)

	mx, my := ebiten.CursorPosition()

	mv := vector.Vector{float64(mx), float64(my)}

	diff := mv.Sub(g.PrevMousePosition)

	g.CameraTilt -= diff[1] * 0.005
	g.CameraRotate -= diff[0] * 0.005

	g.CameraTilt = math.Max(math.Min(g.CameraTilt, math.Pi/2-0.1), -math.Pi/2+0.1)

	rotate := tetra3d.NewMatrix4Rotate(0, 1, 0, g.CameraRotate)

	g.Camera.SetLocalRotation(rotate)

	//
	tilt := tetra3d.NewMatrix4Rotate(1, 0, 0, -1.6)
	rotate1 := tetra3d.NewMatrix4Rotate(0, 1, 0, g.CameraRotate+2)
	for _, v := range g.enemyList {
		g.GameScene.Root.Get(v.Name).SetLocalRotation(tilt.Mult(rotate1))
	}

	g.PrevMousePosition = mv.Clone()

	return err
}

//渲染图片
func (g *Game) RenderImg(screen, img *ebiten.Image, x, y, xs, ys float64) {
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(xs, ys)
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

//获取模型屏幕坐标
func (g *Game) GetScreenPos(name string) vector.Vector {
	return g.Camera.WorldToScreen(g.GameScene.Root.Get(name).WorldPosition())
}
func (g *Game) Draw(screen *ebiten.Image) {
	g.Camera.Clear()
	screen.Fill(color.Black)

	g.Camera.RenderNodes(g.GameScene, g.GameScene.Root)
	screen.DrawImage(g.Camera.ColorTexture(), nil)
	//枪
	g.RenderImg(screen, g.Gun, 338, 270, 1, 1)
	//准星
	g.RenderImg(screen, g.fouce, 338, 184, 1, 1)

	if g.IsFire {
		//火花
		g.RenderImg(screen, g.fire, 330, 184, 1, 1)

	}

	//显示敌人HP
	for _, v := range g.enemyList {
		if v.HP <= 0 {
			v.HP = 48
			v.shoted = false
			g.GameScene.Root.Get(v.Name).SetVisible(false, false)
			go func() {
				time.Sleep(time.Second * 2)
				g.AddDelEm()
			}()
		}
		if v.shoted {
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterLinear
			op.GeoM.Scale(2, 2)
			op.GeoM.Translate(ScreenWidth/2-50, 50)
			screen.DrawImage(g.HP.SubImage(image.Rectangle{image.Point{0, 0}, image.Point{v.HP, 4}}).(*ebiten.Image), op)
		}
	}

	ebitenutil.DebugPrint(screen, strconv.FormatFloat(g.Camera.LocalPosition()[0], 'f', 0, 64)+"\n"+strconv.FormatFloat(g.Camera.LocalPosition()[1], 'f', 0, 64)+"\n"+strconv.FormatFloat(g.Camera.LocalPosition()[2], 'f', 0, 64))

}

func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}
