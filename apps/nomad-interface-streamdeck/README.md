# Nomad Interface Stream Deck

A Go application to control Elgato Stream Deck devices.

## Usage

Run the application to display "Hello" on the first button of the connected Stream Deck.

```bash
go run main.go
```

Or build the binary:

```bash
go build -o streamdeck main.go
./streamdeck
```

## Requirements

- Go 1.24+
- A connected Elgato Stream Deck
- CGO enabled (due to HID library dependency)

## Supported Models

- Original Stream Deck (PID 0x0060) - BMP format
- Stream Deck V2 (PID 0x006d) - JPEG format
- Stream Deck Mini (PID 0x0063) - JPEG format
- Stream Deck XL (PID 0x006c) - JPEG format
- Stream Deck Mini MK2 (PID 0x0090) - JPEG format

## Notes

- Uses the muesli/streamdeck library for Stream Deck communication
- Automatically detects device model and uses appropriate image format
- Renders text using basic Go font
- Requires CGO for HID access