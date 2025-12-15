package main

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// PixelImage is a widget that displays an image at 1:1 pixel size (no stretching).
// It is loosely based on fyne's canvas.Image but ensures the content is always
// rendered at the image's natural pixel size.
//
// Notes:
//   - The widget's MinSize equals the image's pixel dimensions.
//   - The internal canvas.Image is always laid out at its MinSize, ignoring any
//     larger space that a container may try to allocate, avoiding scaling up.
//   - When there is more space, the surrounding area remains empty. Use a Scroll
//     container if you want to navigate when the image is larger than the viewport.
type PixelImage struct {
	widget.BaseWidget

	// img holds the source pixels to render
	img  image.Image
	zoom float64
}

// NewPixelImageFromImage creates a new PixelImage from the provided image.
func NewPixelImageFromImage(img image.Image) *PixelImage {
	p := &PixelImage{img: img, zoom: 1.0}
	p.ExtendBaseWidget(p)
	return p
}

// SetImage replaces the image content and refreshes the widget.
func (p *PixelImage) SetImage(img image.Image) {
	p.img = img
	p.Refresh()
}

// SetZoom sets the zoom of the image
func (p *PixelImage) SetZoom(zoom float64) {
	p.zoom = zoom
	p.Refresh()
}

func (p *PixelImage) Refresh() {
	p.Resize(p.MinSize())
	p.BaseWidget.Refresh()
}

// Zoom is the zoom of the image
func (p *PixelImage) Zoom() float64 {
	return p.zoom
}

// CreateRenderer implements fyne.Widget.
func (p *PixelImage) CreateRenderer() fyne.WidgetRenderer {
	ci := canvas.NewImageFromImage(p.img)
	// We will stretch the image to the requested size but use pixel scaling
	// so zooming keeps crisp pixel edges.
	ci.FillMode = canvas.ImageFillStretch
	ci.ScaleMode = canvas.ImageScalePixels

	r := &pixelImageRenderer{
		p:   p,
		img: ci,
	}
	return r
}

// pixelImageRenderer handles drawing the PixelImage using an internal canvas.Image.
type pixelImageRenderer struct {
	p   *PixelImage
	img *canvas.Image
}

func (r *pixelImageRenderer) Layout(_ fyne.Size) {
	// Always keep the internal image at its natural size (times zoom) to avoid any stretching.
	fyne.Do(func() {
		r.img.Resize(r.MinSize())
		r.img.Move(fyne.NewPos(0, 0))
	})
}

func (r *pixelImageRenderer) MinSize() fyne.Size {
	if r.p.img == nil {
		return fyne.NewSize(0, 0)
	}
	b := r.p.img.Bounds()
	return fyne.NewSize(float32(float64(b.Dx())*r.p.zoom), float32(float64(b.Dy())*r.p.zoom))
}

func (r *pixelImageRenderer) Refresh() {
	// Update the backing canvas.Image to reflect current content
	r.img.Image = r.p.img
	// Keep scaling consistent with renderer setup
	r.img.FillMode = canvas.ImageFillStretch
	r.img.ScaleMode = canvas.ImageScalePixels

	fyne.Do(func() {
		r.img.Refresh()
	})
}

func (r *pixelImageRenderer) Destroy() {}

func (r *pixelImageRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.img}
}
