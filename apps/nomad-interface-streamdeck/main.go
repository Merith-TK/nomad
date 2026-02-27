package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"

	"github.com/sstallion/go-hid"
)

const (
	vendorID  = 0x0fd9 // Elgato
	productID = 0x0080 // Stream Deck MK.2
	pixels    = 72
	cols      = 5
	rows      = 3
)

func main() {
	// Initialize HID library
	if err := hid.Init(); err != nil {
		log.Fatal("Failed to init HID:", err)
	}
	defer hid.Exit()

	// Open the Stream Deck MK.2
	dev, err := hid.OpenFirst(vendorID, productID)
	if err != nil {
		log.Fatal("Failed to open Stream Deck MK.2:", err)
	}
	defer dev.Close()

	fmt.Println("Stream Deck MK.2 opened successfully!")

	// Get device info
	manufacturer, _ := dev.GetMfrStr()
	product, _ := dev.GetProductStr()
	serial, _ := dev.GetSerialNbr()
	fmt.Printf("Device: %s %s (Serial: %s)\n", manufacturer, product, serial)

	// Test: Set brightness to 75%
	fmt.Println("Setting brightness...")
	err = setBrightness(dev, 75)
	if err != nil {
		log.Printf("SetBrightness failed: %v", err)
	} else {
		fmt.Println("Brightness set!")
	}

	// Create a solid red image
	fmt.Println("Creating test image...")
	img := image.NewRGBA(image.Rect(0, 0, pixels, pixels))
	for y := 0; y < pixels; y++ {
		for x := 0; x < pixels; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	// Set image on button 0
	fmt.Println("Setting image on button 0...")
	err = setImage(dev, 0, img)
	if err != nil {
		log.Fatal("SetImage failed:", err)
	}

	fmt.Println("Success! Red square should be on button 0")
}

func setBrightness(dev *hid.Device, percent int) error {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	// Stream Deck MK.2 brightness command
	data := make([]byte, 32)
	data[0] = 0x03
	data[1] = 0x08
	data[2] = byte(percent)

	_, err := dev.SendFeatureReport(data)
	return err
}

func setImage(dev *hid.Device, keyIndex int, img image.Image) error {
	// Convert image to JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return fmt.Errorf("jpeg encode: %w", err)
	}
	imageData := buf.Bytes()

	// Stream Deck MK.2 uses 1024 byte pages with 8 byte header
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

		_, err := dev.Write(report)
		if err != nil {
			return fmt.Errorf("write page %d: %w", page, err)
		}
	}

	return nil
}
