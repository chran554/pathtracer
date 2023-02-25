package rendermonitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/image"
	"pathtracer/internal/pkg/util"
	"sync"
	"time"
)

type RenderMonitor struct {
	groupName  string
	imageName  string
	width      int
	height     int
	connection net.Conn
	lock       sync.Mutex
}

func NewRenderMonitor() RenderMonitor {
	// In IPv4, any address between 224.0.0.0 to 239.255.255.255 can be used as a multicast address.
	address := "230.0.0.0:9999"

	connection, err := net.Dial("udp", address)
	if err != nil {
		fmt.Printf("Could not create multicast connection to render monitor %v", err)
		os.Exit(2)
	}

	return RenderMonitor{connection: connection}
}

func (renderMonitor *RenderMonitor) Close() {
	renderMonitor.lock.Lock()
	defer renderMonitor.lock.Unlock()
	renderMonitor.connection.Close()
}

func (renderMonitor *RenderMonitor) SetPixel(x int, y int, pixelWidth int, pixelHeight int, color *color.Color, amountSamples int) {
	message := getMessage(
		renderMonitor.groupName, renderMonitor.imageName, renderMonitor.width, renderMonitor.height,
		x, y, pixelWidth, pixelHeight, color, amountSamples)

	//	fmt.Println("x:", x, "y:", y, "color:", color)

	renderMonitor.lock.Lock()
	defer renderMonitor.lock.Unlock()

	var err = errors.New("")
	for ok := true; ok; ok = err != nil {
		_, err = renderMonitor.connection.Write(message)
		if err != nil {
			// fmt.Printf("Render monitor connection write error: %v", err)
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (renderMonitor *RenderMonitor) Initialize(imageGroup string, imageName string, width int, height int) {
	renderMonitor.groupName = imageGroup
	renderMonitor.imageName = imageName
	renderMonitor.width = width
	renderMonitor.height = height

	renderMonitor.SetPixel(-1, -1, -1, -1, &color.Black, -1)

	time.Sleep(100 * time.Millisecond)
}

func getMessage(imageGroup string, imageName string,
	imageWidth int, imageHeight int,
	x int, y int, pixelWidth int, pixelHeight int, color *color.Color,
	amountSamples int) []byte {

	c := *color
	c.Multiply(1.0 / float32(amountSamples))
	c = image.GammaEncodeColor(&c, image.GammaDefault)
	c.Multiply(255.0)

	r := uint8(util.Clamp(0, 255, math.Round(float64(c.R))))
	g := uint8(util.Clamp(0, 255, math.Round(float64(c.G))))
	b := uint8(util.Clamp(0, 255, math.Round(float64(c.B))))

	rawColor := [3]uint8{r, g, b}

	message := struct {
		ImageGroup  string   `json:"imageGroup"`
		ImageName   string   `json:"imageName"`
		ImageWidth  int      `json:"imageWidth"`
		ImageHeight int      `json:"imageHeight"`
		X           int      `json:"x"`
		Y           int      `json:"y"`
		PixelWidth  int      `json:"pixelWidth"`
		PixelHeight int      `json:"pixelHeight"`
		Color       [3]uint8 `json:"color"`
	}{
		ImageGroup:  imageGroup,
		ImageName:   imageName,
		ImageWidth:  imageWidth,
		ImageHeight: imageHeight,
		X:           x,
		Y:           y,
		PixelWidth:  pixelWidth,
		PixelHeight: pixelHeight,
		Color:       rawColor,
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Could not marshal data: %+v\n", message)
	}

	return jsonMessage
}
