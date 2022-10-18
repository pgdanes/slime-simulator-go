package main

import (
	"math"
	"math/rand"

	"github.com/veandco/go-sdl2/sdl"
)

const WIDTH = 1200
const HEIGHT = 700
const AGENTS = 10000
const MOVE_SPEED = 1
const SENSE_ANGLE = math.Pi / 8
const SENSE_SIZE = 3
const SENSE_DISTANCE = 8.0
const TURN_SPEED = math.Pi / 4
const DIFFUSE_RATE = 0.4
const EVAPORATION_RATE = 0

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(
		"test",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		WIDTH, HEIGHT,
		sdl.WINDOW_SHOWN)

	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)
	frameOne := make([]uint8, WIDTH*HEIGHT)
	agents := make([]agent, AGENTS)
	rand.Seed(0)
	initRandomAgents(agents)

	running := true
	var i int64 = 0
	for running {
		rand.Seed(i)
		i += 1
		diffuse(frameOne, DIFFUSE_RATE)
		updateAgents(agents, frameOne)
		renderAgentsToFrame(agents, frameOne)
		for i := 0; i < len(frameOne); i++ {
			x, y := index1DTo2D(WIDTH, i)
			value := frameOne[i]
			surface.Set(x, y, sdl.ARGB8888{0xff, value, value, value})
		}
		window.UpdateSurface()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}

func safeSubtract(a, b uint8) uint8 {
	newValue := a - b
	if newValue > a {
		return 0
	} else {
		return newValue
	}
}

func safeAdd(a, b uint8) uint8 {
	newValue := a - b
	if newValue < a {
		return 255
	} else {
		return newValue
	}
}

func diffuse(frame []uint8, difuseSpeed float32) {
	for i := 0; i < len(frame); i++ {
		avg := getAverage(frame, i, WIDTH, HEIGHT)
		lerpedValue := lerp(frame[i], avg, difuseSpeed)
		evaporatedValue := safeSubtract(lerpedValue, EVAPORATION_RATE)
		frame[i] = evaporatedValue
	}
}

func lerp(a, b uint8, t float32) uint8 {
	if a > b {
		return b + uint8(float32(a-b)*t)
	} else {
		return a + uint8(float32(b-a)*t)
	}
}

func getAverage(frame []uint8, index, width, height int) uint8 {
	x, y := index1DTo2D(width, index)

	var sum uint16 = 0
	for offsetX := -1; offsetX <= 1; offsetX++ {
		for offsetY := -1; offsetY <= 1; offsetY++ {
			sampleX := x + offsetX
			sampleY := y + offsetY

			if sampleX >= 0 && sampleX < width && sampleY >= 0 && sampleY < height {
				i := index2DTo1D(width, sampleX, sampleY)
				sum += uint16(frame[i])
			}
		}
	}

	return uint8(sum / 9)
}

func index1DTo2D(width, index int) (x, y int) {
	return (index % width), (index / width)
}

func index2DTo1D(width, x, y int) int {
	return (width * y) + x
}

func updateAgents(agents []agent, frame []uint8) {
	for i := range agents {
		updateAgent(&agents[i], frame)
	}
}

func renderAgentsToFrame(agents []agent, frame []uint8) {
	for i := range agents {
		agent := agents[i]
		pixelIndex := index2DTo1D(WIDTH, int(agent.position.x), int(agent.position.y))
		if pixelIndex > 0 && pixelIndex < WIDTH*HEIGHT {
			frame[pixelIndex] = 255
		}
	}
}

func steerAgent(agent *agent, frame []uint8) {
	var size int = SENSE_SIZE
	var senseAngle float32 = SENSE_ANGLE
	var turnStrength float32 = rand.Float32() * TURN_SPEED

	var randomSteerStrength = rand.Float32() * 5

	weightForward := sense(*agent, frame, 0, size)
	weightLeft := sense(*agent, frame, -(senseAngle), size)
	weightRight := sense(*agent, frame, senseAngle, size)

	if weightForward > weightLeft && weightForward > weightRight {
		agent.angle += 0
	} else if weightForward < weightLeft && weightForward < weightRight {
		agent.angle += (randomSteerStrength - 0.5) * 2 * turnStrength
	} else if weightRight > weightLeft {
		agent.angle += randomSteerStrength * turnStrength
	} else if weightLeft > weightRight {
		agent.angle -= randomSteerStrength * turnStrength
	}
}

func sense(agent agent, frame []uint8, angleOffset float32, sensorSize int) uint32 {
	sensorAngle := agent.angle + angleOffset
	sensorOffset := float32(SENSE_DISTANCE)
	senseDirX := float32(math.Cos(float64(sensorAngle))) * sensorOffset
	senseDirY := float32(math.Sin(float64(sensorAngle))) * sensorOffset
	sensorDir := float2{senseDirX, senseDirY}
	sensorCenter := float2{agent.position.x + sensorDir.x, agent.position.y + sensorDir.y}
	senseX := int(sensorCenter.x)
	senseY := int(sensorCenter.y)

	var sum uint32 = 0

	for offsetX := -sensorSize; offsetX <= sensorSize; offsetX++ {
		for offsetY := -sensorSize; offsetY <= sensorSize; offsetY++ {
			x := senseX + offsetX
			y := senseY + offsetY
			if x >= 0 && x < WIDTH && y >= 0 && y < HEIGHT {
				sum += uint32(frame[index2DTo1D(WIDTH, x, y)])
			}
		}
	}

	return sum
}

func updateAgent(agent *agent, frame []uint8) {
	steerAgent(agent, frame)
	x := math.Cos(float64(agent.angle))
	y := math.Sin(float64(agent.angle))

	newPosX := agent.position.x + float32(x)*MOVE_SPEED
	newPosY := agent.position.y + float32(y)*MOVE_SPEED

	if newPosX < 0 || newPosX > WIDTH || newPosY < 0 || newPosY > HEIGHT {
		newPosX = float32(math.Min(WIDTH-0.01, math.Max(0, float64(newPosX))))
		newPosY = float32(math.Min(HEIGHT-0.01, math.Max(0, float64(newPosY))))
		agent.angle = rand.Float32() * (math.Pi * 2)
	}

	agent.position.x = newPosX
	agent.position.y = newPosY
}

func initRandomAgents(agents []agent) {
	for i := range agents {
		var x float32 = rand.Float32() * WIDTH
		var y float32 = rand.Float32() * HEIGHT
		var angle float32 = rand.Float32() * (math.Pi * 2)
		var pos float2 = float2{x, y}
		agents[i] = agent{pos, angle}
	}
}

type float2 struct {
	x, y float32
}

type agent struct {
	position float2
	angle    float32
}
