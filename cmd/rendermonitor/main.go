package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"net"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/vmihailenco/msgpack/v5"
)

// PixelData mirrors the incoming msgpack structure described in the issue.
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

// imageCanvas encapsulates the UI image, its backing RGBA buffer and synchronization.
type imageCanvas struct {
	mu        sync.Mutex
	img       *image.RGBA
	fyneImage *PixelImage
}

// tabItemState keeps track of per-tab UI and state.
type tabItemState struct {
	canvas   *imageCanvas
	progress *widget.ProgressBar
	zoom     *widget.Label
}

// imageManager keeps track of all images (one per ImageName) and the tabs that display them.
type imageManager struct {
	mu      sync.Mutex
	items   map[string]*tabItemState
	tabs    *container.DocTabs
	winSize fyne.Size
}

// Zoom steps shared by all tabs.
var zoomSteps = []float64{0.125, 0.25, 0.3333333333, 0.5, 0.75, 1.0, 1.5, 2.0, 3.0, 4.0}

func zoomPercent(z float64) string { return fmt.Sprintf("%0.1f%%", z*100) }

func nextZoom(z float64) float64 {
	for _, s := range zoomSteps {
		if s > z+1e-9 {
			return s
		}
	}
	return zoomSteps[len(zoomSteps)-1]
}

func prevZoom(z float64) float64 {
	for i := len(zoomSteps) - 1; i >= 0; i-- {
		if zoomSteps[i] < z-1e-9 {
			return zoomSteps[i]
		}
	}
	return zoomSteps[0]
}

// getOrCreateTab returns the tabItemState for the given image name, creating a new
// tab with its own zoom controls and progress bar if necessary.
// Must be called from the UI goroutine.
func (m *imageManager) getOrCreateTab(name string, w, h int) *tabItemState {
	m.mu.Lock()
	if item, ok := m.items[name]; ok {
		m.mu.Unlock()
		return item
	}
	m.mu.Unlock()

	if w <= 0 {
		w = 640
	}
	if h <= 0 {
		h = 480
	}

	ic := newImageCanvas(w, h)

	// Per-tab progress bar
	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 1
	progressBar.SetValue(0)
	progressBar.TextFormatter = func() string { return fmt.Sprintf("%0.02f%%", progressBar.Value*100) }

	// Per-tab zoom controls
	zLabel := widget.NewLabel(zoomPercent(1.0))
	zLabel.Alignment = fyne.TextAlignCenter

	widest := widestZoomText(zoomSteps, zoomPercent)
	tmp := widget.NewLabel(widest)
	cellSize := tmp.MinSize()
	zoomCell := container.NewGridWrap(cellSize, zLabel)

	zoomOutBtn := widget.NewButtonWithIcon("", theme.ZoomOutIcon(), func() {
		if ic == nil || ic.fyneImage == nil {
			return
		}
		z := ic.fyneImage.Zoom()
		newZ := prevZoom(z)
		ic.fyneImage.SetZoom(newZ)
		zLabel.SetText(zoomPercent(newZ))
	})
	zoomInBtn := widget.NewButtonWithIcon("", theme.ZoomInIcon(), func() {
		if ic == nil || ic.fyneImage == nil {
			return
		}
		z := ic.fyneImage.Zoom()
		newZ := nextZoom(z)
		ic.fyneImage.SetZoom(newZ)
		zLabel.SetText(zoomPercent(newZ))
	})
	leftControls := container.NewHBox(zoomOutBtn, zoomCell, zoomInBtn)

	topBar := container.NewBorder(nil, nil, leftControls, nil, progressBar)

	scroll := container.NewScroll(ic.fyneImage)

	tabContent := container.NewBorder(topBar, nil, nil, nil, scroll)
	tab := container.NewTabItem(name, tabContent)

	item := &tabItemState{
		canvas:   ic,
		progress: progressBar,
		zoom:     zLabel,
	}

	m.mu.Lock()
	if existing, ok := m.items[name]; ok {
		m.mu.Unlock()
		return existing
	}
	m.items[name] = item
	m.tabs.Append(tab)

	// Always select the newly created tab.
	m.tabs.Select(tab)

	m.mu.Unlock()

	return item
}

// currentCanvas returns the imageCanvas for the currently selected tab, or nil if none.
func (m *imageManager) currentCanvas() *imageCanvas {
	m.mu.Lock()
	defer m.mu.Unlock()

	sel := m.tabs.Selected()
	if sel == nil {
		return nil
	}
	item, ok := m.items[sel.Text]
	if !ok {
		return nil
	}
	return item.canvas
}

func main() {
	port := flag.Int("port", 5050, "UDP port to listen for pixel data (msgpack)")
	addr := flag.String("addr", "230.0.0.0", "Address to bind the UDP server")
	flag.Parse()

	application := app.NewWithID("render-monitor")
	w := application.NewWindow("Render Monitor")
	windowSize := fyne.NewSize(800, 600)
	w.Resize(windowSize)

	// Tabs for each image (one per PixelData.ImageName)
	tabs := container.NewDocTabs()
	tabs.SetTabLocation(container.TabLocationTop)

	manager := &imageManager{
		items:   make(map[string]*tabItemState),
		tabs:    tabs,
		winSize: windowSize,
	}

	// Remove per-image state when a tab is closed.
	tabs.OnClosed = func(ti *container.TabItem) {
		manager.mu.Lock()
		delete(manager.items, ti.Text)
		manager.mu.Unlock()
	}

	// Each tab now has its own zoom controls and progress bar.
	w.SetContent(tabs)

	switchTab := func(delta int) {
		if len(tabs.Items) == 0 {
			return
		}
		selectedIndex := -1
		selected := tabs.Selected()
		for i, item := range tabs.Items {
			if item == selected {
				selectedIndex = i
				break
			}
		}
		if selectedIndex == -1 {
			selectedIndex = 0
		}

		newIndex := (selectedIndex + delta) % len(tabs.Items)
		if newIndex < 0 {
			newIndex += len(tabs.Items)
		}
		tabs.Select(tabs.Items[newIndex])
	}

	w.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyLeft, Modifier: fyne.KeyModifierAlt}, func(shortcut fyne.Shortcut) {
		switchTab(-1)
	})
	w.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyRight, Modifier: fyne.KeyModifierAlt}, func(shortcut fyne.Shortcut) {
		switchTab(1)
	})

	// Start UDP server in background
	go udpServer(net.JoinHostPort(*addr, itoa(*port)), manager)

	w.Show()
	refreshUI := func() {
		manager.mu.Lock()
		for _, item := range manager.items {
			if item.canvas != nil && item.canvas.fyneImage != nil {
				item.canvas.fyneImage.Refresh()
			}
			if item.progress != nil {
				item.progress.Refresh()
			}
		}
		manager.mu.Unlock()
	}

	go func() {
		// Periodic refresh to ensure UI stays responsive during bursts
		for range time.NewTicker(200 * time.Millisecond).C {
			app := fyne.CurrentApp()
			if app != nil && app.Driver() != nil {
				app.Driver().DoFromGoroutine(refreshUI, false)
			} else {
				// Fallback if no app yet (shouldn't happen post-start)
				go refreshUI()
			}
		}
	}()
	application.Run()
}

func widestZoomText(zoomSteps []float64, zoomPercent func(z float64) string) string {
	// Make the zoom label fixed width using the width of the widest zoom text.
	// We compute the widest text of our zoom steps (and a safety pattern) and wrap the label
	// in a GridWrap container with that cell size so it does not resize as text changes.
	var widest string
	{
		candidates := make([]string, 0, len(zoomSteps)+1)
		candidates = append(candidates, zoomPercent(1.0)) // safety in case of future zoom ranges
		for _, s := range zoomSteps {
			candidates = append(candidates, zoomPercent(s))
		}

		thSize := theme.Size(theme.SizeNameText)
		style := fyne.TextStyle{}

		maxW := float32(0)
		for _, t := range candidates {
			sz := fyne.MeasureText(t, thSize, style)
			if sz.Width > maxW {
				maxW = sz.Width
				widest = t
			}
		}
	}
	return widest
}

func newImageCanvas(w, h int) *imageCanvas {
	if w <= 0 {
		w = 640
	}
	if h <= 0 {
		h = 480
	}
	r := image.Rect(0, 0, w, h)
	rgba := image.NewRGBA(r)
	// Initialize as transparent black
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{C: color.RGBA{0, 0, 0, 0}}, image.Point{}, draw.Src)

	fyImg := NewPixelImageFromImage(rgba)

	return &imageCanvas{img: rgba, fyneImage: fyImg}
}

// resizeImage safely replaces the backing image and updates the UI object size.
func (ic *imageCanvas) resizeImage(w, h int) {
	if w <= 0 || h <= 0 {
		return
	}

	ic.mu.Lock()
	defer ic.mu.Unlock()

	ic.img = image.NewRGBA(image.Rect(0, 0, w, h))
	ic.fyneImage.SetImage(ic.img)
}

// applyPixelData writes the incoming pixels/rectangles into the backing image.
func (ic *imageCanvas) applyPixelData(pd PixelData) {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	if ic.img == nil || ic.img.Bounds().Dx() != pd.ImageWidth || ic.img.Bounds().Dy() != pd.ImageHeight {
		ic.img = image.NewRGBA(image.Rect(0, 0, pd.ImageWidth, pd.ImageHeight))
		ic.fyneImage.SetImage(ic.img)
	}

	// Clamp rectangle to image bounds
	x0 := max(0, pd.X)
	y0 := max(0, pd.Y)
	x1 := min(ic.img.Bounds().Dx(), pd.X+pd.PixelWidth)
	y1 := min(ic.img.Bounds().Dy(), pd.Y+pd.PixelHeight)

	if x1 <= x0 || y1 <= y0 {
		return
	}

	w := x1 - x0
	h := y1 - y0

	// Interpret color array
	switch {
	case len(pd.Color) == 4:
		// Single RGBA color fill
		c := color.RGBA{uint8(pd.Color[0]), uint8(pd.Color[1]), uint8(pd.Color[2]), uint8(pd.Color[3])}
		draw.Draw(ic.img, image.Rect(x0, y0, x1, y1), &image.Uniform{C: c}, image.Point{}, draw.Src)
	case len(pd.Color) == w*h*4:
		// Per-pixel RGBA
		idx := 0
		for yy := y0; yy < y1; yy++ {
			for xx := x0; xx < x1; xx++ {
				r := uint8(pd.Color[idx])
				g := uint8(pd.Color[idx+1])
				b := uint8(pd.Color[idx+2])
				a := uint8(pd.Color[idx+3])
				ic.img.SetRGBA(xx, yy, color.RGBA{R: r, G: g, B: b, A: a})
				idx += 4
			}
		}
	case len(pd.Color) == w*h:
		// Assume packed ARGB ints per pixel
		idx := 0
		for yy := y0; yy < y1; yy++ {
			for xx := x0; xx < x1; xx++ {
				argb := uint32(pd.Color[idx])
				a := uint8((argb >> 24) & 0xFF)
				r := uint8((argb >> 16) & 0xFF)
				g := uint8((argb >> 8) & 0xFF)
				b := uint8(argb & 0xFF)
				ic.img.SetRGBA(xx, yy, color.RGBA{R: r, G: g, B: b, A: a})
				idx++
			}
		}
	default:
		// Unknown format; do nothing
	}
}

// tcpServer listens for msgpack PixelData on the specified address and updates the image.
func tcpServer(addr string, ic *imageCanvas, progress *widget.ProgressBar) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("TCP listen failed on %s: %v", addr, err)
		return
	}
	defer ln.Close()

	log.Printf("Listening for pixel data on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleTcpConn(conn, ic, progress)
	}
}

func handleTcpConn(conn net.Conn, ic *imageCanvas, progress *widget.ProgressBar) {
	defer conn.Close()
	dec := msgpack.NewDecoder(conn)
	dec.SetCustomStructTag("msgpack")
	for {
		var pd PixelData
		if err := dec.Decode(&pd); err != nil {
			log.Printf("Decode error: %v", err)
			return
		}

		ic.applyPixelData(pd)
		fyne.Do(func() {
			progress.SetValue(pd.Progress)
		})
	}
}

// itoa is a small helper to avoid another import.
func itoa(i int) string { return fmtInt(i) }

func fmtInt(i int) string {
	// fast integer to string without strconv for minimal deps; for simplicity, use standard approach
	// But to keep dependencies minimal we implement a simple version here.
	if i == 0 {
		return "0"
	}
	sign := ""
	if i < 0 {
		sign = "-"
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + (i % 10))
		i /= 10
	}
	return sign + string(buf[pos:])
}

// udpServer listens for msgpack PixelData via UDP on the specified address and updates images.
// Each distinct PixelData.ImageName gets its own tab, backing image, zoom controls and progress bar.
func udpServer(addr string, mgr *imageManager) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf("UDP resolve failed on %s: %v", addr, err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf("UDP listen failed on %s: %v", addr, err)
		return
	}
	defer conn.Close()

	log.Printf("Listening for pixel data (UDP) on %s", addr)

	// Create a queue to buffer incoming messages
	packetQueue := make(chan []byte, 1024)

	// Worker goroutine to process messages from the queue
	go func() {
		for data := range packetQueue {
			var pd PixelData
			if err := msgpack.Unmarshal(data, &pd); err != nil {
				log.Printf("UDP msgpack unmarshal error: %v", err)
				continue
			}

			// Ensure the tab and image for this ImageName exist, then apply the update.
			var item *tabItemState

			done := make(chan *tabItemState, 1)
			fyne.Do(func() {
				item = mgr.getOrCreateTab(pd.ImageName, pd.ImageWidth, pd.ImageHeight)
				done <- item
			})
			item = <-done

			if item == nil || item.canvas == nil {
				continue
			}
			item.canvas.applyPixelData(pd)

			fyne.Do(func() {
				// Update the progress bar on this image's tab.
				if item.progress != nil {
					item.progress.SetValue(pd.Progress)
				}

				application := fyne.CurrentApp()
				if application == nil || application.Driver() == nil {
					return
				}

				title := fmt.Sprintf("Rendering %s (%.02f%%)", pd.ImageName, pd.Progress*100)
				for _, window := range application.Driver().AllWindows() {
					window.SetTitle(title)
				}
			})
		}
	}()

	buf := make([]byte, 65535)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("UDP read error: %v", err)
			continue
		}

		// Copy the data to a new slice so the buffer can be reused immediately
		packet := make([]byte, n)
		copy(packet, buf[:n])

		select {
		case packetQueue <- packet:
		default:
			// Queue is full; drop packet to keep the read loop running
		}
	}
}
