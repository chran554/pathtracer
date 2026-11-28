package floatimage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mrjoshuak/go-openexr/exr"
	"github.com/mrjoshuak/go-openexr/exrmeta"
)

// EXRMetaData holds optional frame information. Zero-valued fields are omitted from the EXR header.
type EXRMetaData struct {
	CaptureTime time.Time
	Comments    string
	Owner       string
}

// WriteEXRImage writes linear RGB and converts the renderer's alpha hit counts to pixel coverage.
// Metadata can be nil.
func WriteEXRImage(filename string, image *FloatImage, samplesPerPixel int, metadata *EXRMetaData) error {
	if samplesPerPixel <= 0 {
		return fmt.Errorf("EXR alpha coverage requires a positive sample count, got %d", samplesPerPixel)
	}
	if !image.ContainImageData() {
		return errors.New("cannot write an empty EXR image")
	}

	file, err := os.CreateTemp(filepath.Dir(filename), ".frame-*.exr")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	writeErr := encodeEXR(file, image, samplesPerPixel, metadata)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	return os.Rename(file.Name(), filename)
}

func encodeEXR(file io.WriteSeeker, image *FloatImage, samplesPerPixel int, metadata *EXRMetaData) error {
	header := exr.NewScanlineHeader(image.Width, image.Height)
	if metadata != nil {
		if metadata.Owner != "" {
			exrmeta.SetOwner(header, metadata.Owner)
		}
		if metadata.Comments != "" {
			exrmeta.SetComments(header, metadata.Comments)
		}
		if !metadata.CaptureTime.IsZero() {
			exrmeta.SetCapDate(header, metadata.CaptureTime.Format("2006:01:02 15:04:05"))
			_, secondsEastOfUTC := metadata.CaptureTime.Zone()
			// OpenEXR defines utcOffset as the seconds to add to local time to obtain UTC.
			exrmeta.SetUTCOffset(header, float32(-secondsEastOfUTC))
		}
	}

	header.SetCompression(exr.CompressionZIP)
	channels := exr.NewChannelList()
	for _, name := range []string{"R", "G", "B", "A"} {
		channels.Add(exr.NewChannel(name, exr.PixelTypeFloat))
	}
	header.SetChannels(channels)

	buffer, _, err := exr.AllocateChannelsLimit(channels, header.DataWindow(), int64(len(image.pixels))*16)
	if err != nil {
		return err
	}
	red, green, blue, alpha := buffer.Get("R"), buffer.Get("G"), buffer.Get("B"), buffer.Get("A")
	for y := range image.Height {
		for x := range image.Width {
			pixel := image.GetPixel(x, y)
			red.SetFloat32(x, y, pixel.R)
			green.SetFloat32(x, y, pixel.G)
			blue.SetFloat32(x, y, pixel.B)
			alpha.SetFloat32(x, y, pixel.A/float32(samplesPerPixel))
		}
	}

	writer, err := exr.NewScanlineWriter(file, header)
	if err != nil {
		return err
	}
	writer.SetFrameBuffer(buffer)
	if err := writer.WritePixels(0, image.Height-1); err != nil {
		return err
	}
	return writer.Close()
}
