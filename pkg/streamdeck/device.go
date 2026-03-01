package streamdeck

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"sync"
	"time"

	"github.com/sstallion/go-hid"
)

// Device represents an opened Stream Deck device.
type Device struct {
	hid   *hid.Device
	Info  DeviceInfo
	Model Model
	mu    sync.Mutex // protects HID operations
}

// KeyEvent represents a key press or release event.
type KeyEvent struct {
	Key     int
	Pressed bool
}

// Open opens a Stream Deck device by its HID path.
func Open(path string) (*Device, error) {
	dev, err := hid.OpenPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %w", err)
	}

	// Get device info
	manufacturer, _ := dev.GetMfrStr()
	product, _ := dev.GetProductStr()
	serial, _ := dev.GetSerialNbr()

	// We need to get product ID - enumerate to find it
	var productID uint16
	err = hid.Enumerate(VendorID, 0x0000, func(info *hid.DeviceInfo) error {
		if info.Path == path {
			productID = info.ProductID
		}
		return nil
	})
	if err != nil {
		dev.Close()
		return nil, fmt.Errorf("failed to get product ID: %w", err)
	}

	model, _ := LookupModel(productID)
	if model.ProductID == 0 {
		model = Model{
			Name:      fmt.Sprintf("Unknown Stream Deck (PID: 0x%04X)", productID),
			ProductID: productID,
		}
	}

	d := &Device{
		hid:   dev,
		Model: model,
		Info: DeviceInfo{
			Path:         path,
			Serial:       serial,
			Manufacturer: manufacturer,
			Product:      product,
			Model:        model,
			Firmware:     getFirmwareVersion(dev),
		},
	}

	return d, nil
}

// OpenFirst opens the first Stream Deck device found.
func OpenFirst() (*Device, error) {
	devices, err := Enumerate()
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("no Stream Deck devices found")
	}
	return Open(devices[0].Path)
}

// Close closes the device.
func (d *Device) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.hid != nil {
		return d.hid.Close()
	}
	return nil
}

// SetBrightness sets the brightness of the Stream Deck (0-100).
func (d *Device) SetBrightness(percent int) error {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	data := make([]byte, 32)
	data[0] = 0x03
	data[1] = 0x08
	data[2] = byte(percent)

	_, err := d.hid.SendFeatureReport(data)
	return err
}

// Reset resets the Stream Deck to its default state.
func (d *Device) Reset() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	data := make([]byte, 32)
	data[0] = 0x03
	data[1] = 0x02

	_, err := d.hid.SendFeatureReport(data)
	return err
}

// SetImage sets the image on a specific key.
// The image is automatically resized to fit the key and rotated 180° for correct display.
func (d *Device) SetImage(keyIndex int, img image.Image) error {
	if keyIndex < 0 || keyIndex >= d.Model.Keys {
		return fmt.Errorf("key index %d out of range (0-%d)", keyIndex, d.Model.Keys-1)
	}
	if d.Model.PixelSize == 0 {
		return fmt.Errorf("device does not support images")
	}

	// Prepare the image: resize and rotate
	prepared := d.prepareImage(img)

	// Encode to appropriate format
	imageData, err := d.encodeImage(prepared)
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	return d.writeImageData(keyIndex, imageData)
}

// SetImageRaw sets raw image data on a key without any processing.
// Use this if you've already prepared the image correctly.
func (d *Device) SetImageRaw(keyIndex int, imageData []byte) error {
	if keyIndex < 0 || keyIndex >= d.Model.Keys {
		return fmt.Errorf("key index %d out of range (0-%d)", keyIndex, d.Model.Keys-1)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	return d.writeImageData(keyIndex, imageData)
}

// prepareImage resizes and rotates the image for Stream Deck display.
func (d *Device) prepareImage(src image.Image) image.Image {
	size := d.Model.PixelSize
	bounds := src.Bounds()

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, size, size))

	// If source is correct size, just copy with rotation
	if bounds.Dx() == size && bounds.Dy() == size {
		// Rotate 180 degrees
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dst.Set(size-1-x, size-1-y, src.At(bounds.Min.X+x, bounds.Min.Y+y))
			}
		}
		return dst
	}

	// Scale the image to fit
	scaleX := float64(bounds.Dx()) / float64(size)
	scaleY := float64(bounds.Dy()) / float64(size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Sample from source with rotation (180 degrees)
			srcX := int(float64(size-1-x) * scaleX)
			srcY := int(float64(size-1-y) * scaleY)
			dst.Set(x, y, src.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return dst
}

// encodeImage encodes the image to the appropriate format for this device.
func (d *Device) encodeImage(img image.Image) ([]byte, error) {
	var buf bytes.Buffer

	switch d.Model.ImageFormat {
	case "JPEG":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
		if err != nil {
			return nil, fmt.Errorf("jpeg encode: %w", err)
		}
	case "BMP":
		// BMP encoding for older devices
		err := encodeBMP(&buf, img)
		if err != nil {
			return nil, fmt.Errorf("bmp encode: %w", err)
		}
	default:
		// Default to JPEG
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
		if err != nil {
			return nil, fmt.Errorf("jpeg encode: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// writeImageData writes raw image data to a key.
func (d *Device) writeImageData(keyIndex int, imageData []byte) error {
	// Stream Deck MK.2/V2 uses 1024 byte pages with 8 byte header
	pageSize := 1024
	headerSize := 8
	payloadSize := pageSize - headerSize

	totalPages := (len(imageData) + payloadSize - 1) / payloadSize

	for page := 0; page < totalPages; page++ {
		start := page * payloadSize
		end := start + payloadSize
		if end > len(imageData) {
			end = len(imageData)
		}
		chunk := imageData[start:end]

		isLastPage := page == totalPages-1

		// Build the report
		report := make([]byte, pageSize)
		report[0] = 0x02           // Report ID for image
		report[1] = 0x07           // Command
		report[2] = byte(keyIndex) // Key index
		if isLastPage {
			report[3] = 0x01 // Last page flag
		} else {
			report[3] = 0x00
		}
		report[4] = byte(len(chunk) & 0xFF) // Payload length low byte
		report[5] = byte(len(chunk) >> 8)   // Payload length high byte
		report[6] = byte(page & 0xFF)       // Page number low byte
		report[7] = byte(page >> 8)         // Page number high byte

		copy(report[headerSize:], chunk)

		_, err := d.hid.Write(report)
		if err != nil {
			return fmt.Errorf("write page %d: %w", page, err)
		}
	}

	return nil
}

// Clear clears all keys on the Stream Deck (sets them to black).
func (d *Device) Clear() error {
	if d.Model.PixelSize == 0 {
		return nil // No display to clear
	}
	black := image.NewRGBA(image.Rect(0, 0, d.Model.PixelSize, d.Model.PixelSize))
	for i := 0; i < d.Model.Keys; i++ {
		if err := d.SetImage(i, black); err != nil {
			return fmt.Errorf("clear key %d: %w", i, err)
		}
	}
	return nil
}

// SetKeyColor sets a key to a solid color.
func (d *Device) SetKeyColor(keyIndex int, c color.Color) error {
	if d.Model.PixelSize == 0 {
		return fmt.Errorf("device does not support images")
	}
	size := d.Model.PixelSize
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
	return d.SetImage(keyIndex, img)
}

// ReadKeys reads the current state of all keys.
// Returns a slice of booleans where true means the key is pressed.
func (d *Device) ReadKeys() ([]bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Read buffer size depends on device, use generous buffer
	buf := make([]byte, 512)
	n, err := d.hid.ReadWithTimeout(buf, 100*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("read keys: %w", err)
	}
	if n == 0 {
		// No data available, return current state as all unpressed
		return make([]bool, d.Model.Keys), nil
	}

	// Parse key states - format depends on device generation
	// For MK.2/V2: first byte is report ID (0x01), then key states starting at offset 4
	keys := make([]bool, d.Model.Keys)
	keyOffset := 4 // MK.2/V2 offset
	for i := 0; i < d.Model.Keys && keyOffset+i < n; i++ {
		keys[i] = buf[keyOffset+i] != 0
	}

	return keys, nil
}

// WaitForKeyPress blocks until a key is pressed or the context is cancelled.
// Returns the index of the pressed key.
func (d *Device) WaitForKeyPress(ctx context.Context) (int, error) {
	prevState := make([]bool, d.Model.Keys)

	for {
		select {
		case <-ctx.Done():
			return -1, ctx.Err()
		default:
		}

		keys, err := d.ReadKeys()
		if err != nil {
			return -1, err
		}

		// Check for newly pressed keys
		for i, pressed := range keys {
			if pressed && !prevState[i] {
				return i, nil
			}
		}

		copy(prevState, keys)
		time.Sleep(10 * time.Millisecond)
	}
}

// ListenKeys starts listening for key events and sends them to the provided channel.
// Closes the channel when context is cancelled.
func (d *Device) ListenKeys(ctx context.Context, events chan<- KeyEvent) {
	go func() {
		defer close(events)
		prevState := make([]bool, d.Model.Keys)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			keys, err := d.ReadKeys()
			if err != nil {
				continue
			}

			// Detect state changes
			for i, pressed := range keys {
				if pressed != prevState[i] {
					select {
					case events <- KeyEvent{Key: i, Pressed: pressed}:
					case <-ctx.Done():
						return
					}
				}
			}

			copy(prevState, keys)
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

// KeyToCoord converts a key index to (col, row) coordinates.
func (d *Device) KeyToCoord(keyIndex int) (col, row int) {
	if d.Model.Cols == 0 {
		return 0, 0
	}
	return keyIndex % d.Model.Cols, keyIndex / d.Model.Cols
}

// CoordToKey converts (col, row) coordinates to a key index.
func (d *Device) CoordToKey(col, row int) int {
	return row*d.Model.Cols + col
}

// Cols returns the number of columns on the device.
func (d *Device) Cols() int {
	return d.Model.Cols
}

// Rows returns the number of rows on the device.
func (d *Device) Rows() int {
	return d.Model.Rows
}

// Keys returns the total number of keys on the device.
func (d *Device) Keys() int {
	return d.Model.Keys
}

// PixelSize returns the pixel dimensions for key images.
func (d *Device) PixelSize() int {
	return d.Model.PixelSize
}

// ResizeImage scales an image to fit the device's key size.
// Maintains aspect ratio and centers the image.
func (d *Device) ResizeImage(src image.Image) image.Image {
	size := d.Model.PixelSize
	if size == 0 {
		return src
	}

	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// If already correct size, return as-is
	if srcW == size && srcH == size {
		return src
	}

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, size, size))

	// Calculate scale to fit while maintaining aspect ratio
	scaleX := float64(size) / float64(srcW)
	scaleY := float64(size) / float64(srcH)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)

	// Center offset
	offsetX := (size - newW) / 2
	offsetY := (size - newH) / 2

	// Use nearest-neighbor for speed (called at 10fps)
	// Scale manually
	for y := 0; y < newH; y++ {
		srcY := srcBounds.Min.Y + int(float64(y)/scale)
		for x := 0; x < newW; x++ {
			srcX := srcBounds.Min.X + int(float64(x)/scale)
			dst.Set(offsetX+x, offsetY+y, src.At(srcX, srcY))
		}
	}

	return dst
}

// encodeBMP encodes an image to BMP format for older Stream Deck devices.
func encodeBMP(w *bytes.Buffer, img image.Image) error {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// BMP row size must be aligned to 4 bytes
	rowSize := ((width*3 + 3) / 4) * 4
	imageSize := rowSize * height
	fileSize := 54 + imageSize // 54 = header size

	// BMP File Header (14 bytes)
	w.Write([]byte{'B', 'M'})      // Magic number
	writeLE32(w, uint32(fileSize)) // File size
	writeLE16(w, 0)                // Reserved
	writeLE16(w, 0)                // Reserved
	writeLE32(w, 54)               // Offset to pixel data

	// DIB Header (40 bytes - BITMAPINFOHEADER)
	writeLE32(w, 40)                // Header size
	writeLE32(w, uint32(width))     // Width
	writeLE32(w, uint32(height))    // Height (positive = bottom-up)
	writeLE16(w, 1)                 // Color planes
	writeLE16(w, 24)                // Bits per pixel
	writeLE32(w, 0)                 // Compression (none)
	writeLE32(w, uint32(imageSize)) // Image size
	writeLE32(w, 2835)              // Horizontal resolution (72 DPI)
	writeLE32(w, 2835)              // Vertical resolution (72 DPI)
	writeLE32(w, 0)                 // Colors in palette
	writeLE32(w, 0)                 // Important colors

	// Pixel data (bottom-up, BGR format)
	row := make([]byte, rowSize)
	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			row[x*3+0] = byte(b >> 8) // B
			row[x*3+1] = byte(g >> 8) // G
			row[x*3+2] = byte(r >> 8) // R
		}
		w.Write(row)
	}

	return nil
}

func writeLE16(w *bytes.Buffer, v uint16) {
	w.WriteByte(byte(v))
	w.WriteByte(byte(v >> 8))
}

func writeLE32(w *bytes.Buffer, v uint32) {
	w.WriteByte(byte(v))
	w.WriteByte(byte(v >> 8))
	w.WriteByte(byte(v >> 16))
	w.WriteByte(byte(v >> 24))
}
