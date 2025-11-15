package rendermonitor

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/util"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type RenderMonitor struct {
	groupName  string
	imageName  string
	width      int
	height     int
	connection net.Conn
	lock       sync.Mutex
}

type PixelData struct {
	ImageGroup  string  `msgpack:"imageGroup"`
	ImageName   string  `msgpack:"imageName"`
	ImageWidth  int     `msgpack:"imageWidth"`
	ImageHeight int     `msgpack:"imageHeight"`
	Progress    float64 `msgpack:"progress"`
	X           int     `msgpack:"x"`
	Y           int     `msgpack:"y"`
	PixelWidth  int     `msgpack:"pixelWidth"`
	PixelHeight int     `msgpack:"pixelHeight"`
	Color       []int   `msgpack:"color"`
}

func NewRenderMonitor() *RenderMonitor {
	// In IPv4, any address between 224.0.0.0 to 239.255.255.255 can be used as a multicast address.
	address := "127.0.0.1:5050"

	connection, err := net.Dial("udp", address)
	if err != nil {
		fmt.Printf("Could not create multicast connection to render monitor %v", err)
		os.Exit(2)
	}

	return &RenderMonitor{connection: connection}
}

func (renderMonitor *RenderMonitor) Close() {
	renderMonitor.lock.Lock()
	defer renderMonitor.lock.Unlock()
	renderMonitor.connection.Close()
}

func (renderMonitor *RenderMonitor) SetPixel(x int, y int, pixelWidth int, pixelHeight int, color *color.Color, amountSamples int, progress float64) {
	message := getMessage(
		renderMonitor.groupName, renderMonitor.imageName, renderMonitor.width, renderMonitor.height,
		x, y, pixelWidth, pixelHeight, color, amountSamples, progress)

	// fmt.Println("x:", x, "y:", y, "color:", color)

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

	renderMonitor.SetPixel(-1, -1, -1, -1, &color.Black, -1, 0)

	time.Sleep(100 * time.Millisecond)
}

func getMessage(imageGroup string, imageName string, imageWidth int, imageHeight int, x int, y int, pixelWidth int, pixelHeight int, c *color.Color, amountSamples int, progress float64) []byte {
	c = c.Copy()
	c.Multiply(1.0 / float32(amountSamples))
	c = c.GammaEncode(floatimage.GammaDefault)
	c.Multiply(255.0)

	r := int(util.ClampFloat64(0, 255, math.Round(float64(c.R))))
	g := int(util.ClampFloat64(0, 255, math.Round(float64(c.G))))
	b := int(util.ClampFloat64(0, 255, math.Round(float64(c.B))))

	rawColor := []int{r, g, b, 255}

	buffer := &bytes.Buffer{}
	enc := msgpack.NewEncoder(buffer)
	enc.SetCustomStructTag("msgpack")

	pd := PixelData{
		ImageGroup:  imageGroup,
		ImageName:   imageName,
		ImageWidth:  imageWidth,
		ImageHeight: imageHeight,
		Progress:    progress,
		X:           x,
		Y:           y,
		PixelWidth:  pixelWidth,
		PixelHeight: pixelHeight,
		Color:       rawColor,
	}

	if err := enc.Encode(pd); err != nil {
		fmt.Printf("Could not marshal data: %+v\n", pd)
	}

	return buffer.Bytes()
}
