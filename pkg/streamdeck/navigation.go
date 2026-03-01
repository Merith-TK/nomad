package streamdeck

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// PageItem represents an item on a page (folder or action).
type PageItem struct {
	Name     string // Display name
	Path     string // Full path to the item
	IsFolder bool   // True if this is a folder
	Script   string // Path to lua script (if action)
}

// Page represents a single page of items on the Stream Deck.
type Page struct {
	Path       string     // Current directory path
	Items      []PageItem // Items on this page
	ParentPath string     // Path to parent directory (empty if root)
	PageIndex  int        // Current page index (for pagination)
	TotalPages int        // Total number of pages
}

// Reserved key indices (column 0 on a 5-column deck)
// Layout: key index = row * cols + col
// Row 0: 0,1,2,3,4
// Row 1: 5,6,7,8,9
// Row 2: 10,11,12,13,14
//
// TODO: Reserved keys are currently hardcoded for MK.2 (5 cols x 3 rows).
// These should be dynamically calculated based on the device model's row count
// and column layout. Consider: ReservedKeys = [0, cols, cols*2, ...] for col 0.
const (
	KeyBack    = 0  // Row 0, Col 0 - Navigate back
	KeyToggle1 = 5  // Row 1, Col 0 - Reserved toggle (placeholder)
	KeyToggle2 = 10 // Row 2, Col 0 - Reserved toggle (placeholder)
)

// Navigator manages folder-based navigation on a Stream Deck.
type Navigator struct {
	dev          *Device
	rootPath     string
	currentDir   string
	pageIndex    int
	contentKeys  []int        // Key indices available for content (excludes column 0)
	reservedKeys []int        // Key indices for reserved functions (column 0)
	toggleStates map[int]bool // Toggle state for dummy keys
}

// NewNavigator creates a new navigator for the given device and root config path.
func NewNavigator(dev *Device, rootPath string) *Navigator {
	n := &Navigator{
		dev:          dev,
		rootPath:     rootPath,
		currentDir:   rootPath,
		pageIndex:    0,
		toggleStates: make(map[int]bool),
	}
	n.calculateKeyLayout()
	return n
}

// calculateKeyLayout determines which keys are for content vs reserved.
func (n *Navigator) calculateKeyLayout() {
	cols := n.dev.Cols()
	rows := n.dev.Rows()

	n.contentKeys = nil
	n.reservedKeys = nil

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			keyIndex := row*cols + col
			if col == 0 {
				// Column 0 is reserved
				n.reservedKeys = append(n.reservedKeys, keyIndex)
			} else {
				n.contentKeys = append(n.contentKeys, keyIndex)
			}
		}
	}
}

// ContentKeyCount returns the number of keys available for content.
func (n *Navigator) ContentKeyCount() int {
	return len(n.contentKeys)
}

// CurrentPath returns the current directory path.
func (n *Navigator) CurrentPath() string {
	return n.currentDir
}

// IsAtRoot returns true if we're at the root config directory.
func (n *Navigator) IsAtRoot() bool {
	return n.currentDir == n.rootPath
}

// LoadPage loads the current page and returns page info.
func (n *Navigator) LoadPage() (*Page, error) {
	entries, err := os.ReadDir(n.currentDir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", n.currentDir, err)
	}

	// Filter and sort entries
	var items []PageItem
	for _, entry := range entries {
		// Skip hidden files and special files (starting with . or _)
		if len(entry.Name()) > 0 && (entry.Name()[0] == '.' || entry.Name()[0] == '_') {
			continue
		}

		item := PageItem{
			Name:     entry.Name(),
			Path:     filepath.Join(n.currentDir, entry.Name()),
			IsFolder: entry.IsDir(),
		}

		// Check for lua script
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".lua" {
			item.Script = item.Path
			item.Name = entry.Name()[:len(entry.Name())-4] // Remove .lua extension
		}

		items = append(items, item)
	}

	// Sort: folders first, then alphabetically
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsFolder != items[j].IsFolder {
			return items[i].IsFolder
		}
		return items[i].Name < items[j].Name
	})

	// Calculate pagination using content keys only (excludes reserved column)
	keysAvailable := n.ContentKeyCount()

	totalPages := 1
	if len(items) > keysAvailable {
		totalPages = (len(items) + keysAvailable - 1) / keysAvailable
	}

	// Clamp page index
	if n.pageIndex >= totalPages {
		n.pageIndex = totalPages - 1
	}
	if n.pageIndex < 0 {
		n.pageIndex = 0
	}

	// Get items for current page
	start := n.pageIndex * keysAvailable
	end := start + keysAvailable
	if end > len(items) {
		end = len(items)
	}

	pageItems := items[start:end]

	// Determine parent path
	parentPath := ""
	if !n.IsAtRoot() {
		parentPath = filepath.Dir(n.currentDir)
	}

	return &Page{
		Path:       n.currentDir,
		Items:      pageItems,
		ParentPath: parentPath,
		PageIndex:  n.pageIndex,
		TotalPages: totalPages,
	}, nil
}

// NavigateInto enters a subdirectory.
func (n *Navigator) NavigateInto(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", path)
	}
	n.currentDir = path
	n.pageIndex = 0
	return nil
}

// NavigateBack goes to the parent directory.
func (n *Navigator) NavigateBack() bool {
	if n.IsAtRoot() {
		return false
	}
	n.currentDir = filepath.Dir(n.currentDir)
	n.pageIndex = 0
	return true
}

// NavigateToRoot returns to the root config directory.
func (n *Navigator) NavigateToRoot() {
	n.currentDir = n.rootPath
	n.pageIndex = 0
}

// NextPage moves to the next page.
func (n *Navigator) NextPage() bool {
	page, err := n.LoadPage()
	if err != nil {
		return false
	}
	if n.pageIndex < page.TotalPages-1 {
		n.pageIndex++
		return true
	}
	return false
}

// PrevPage moves to the previous page.
func (n *Navigator) PrevPage() bool {
	if n.pageIndex > 0 {
		n.pageIndex--
		return true
	}
	return false
}

// RenderPage renders the current page to the Stream Deck.
// Images are encoded concurrently, then written to the device serially.
// No Clear() pass is needed — every key is explicitly overwritten.
func (n *Navigator) RenderPage() error {
	page, err := n.LoadPage()
	if err != nil {
		return err
	}

	totalKeys := n.dev.Model.Keys
	type keyFrame struct {
		index int
		data  []byte
		err   error
	}

	frames := make([]keyFrame, totalKeys)
	for i := range frames {
		frames[i].index = i
	}

	// Build image for every key (nil = black / unused)
	images := make([]image.Image, totalKeys)

	// Reserved column
	if !n.IsAtRoot() {
		images[KeyBack] = n.createTextImage("<-", color.RGBA{100, 100, 100, 255})
	} else {
		images[KeyBack] = n.createTextImage("HOME", color.RGBA{50, 50, 50, 255})
	}
	if n.toggleStates[KeyToggle1] {
		images[KeyToggle1] = n.createTextImage("T1:ON", color.RGBA{0, 150, 0, 255})
	} else {
		images[KeyToggle1] = n.createTextImage("T1", color.RGBA{80, 80, 80, 255})
	}
	if n.toggleStates[KeyToggle2] {
		images[KeyToggle2] = n.createTextImage("T2:ON", color.RGBA{0, 150, 0, 255})
	} else {
		images[KeyToggle2] = n.createTextImage("T2", color.RGBA{80, 80, 80, 255})
	}

	// Content keys
	for i, item := range page.Items {
		if i >= len(n.contentKeys) {
			break
		}
		if item.IsFolder {
			images[n.contentKeys[i]] = n.createTextImage(truncateName(item.Name, 8), color.RGBA{30, 80, 180, 255})
		} else {
			images[n.contentKeys[i]] = n.createTextImage(truncateName(item.Name, 8), color.RGBA{30, 130, 80, 255})
		}
	}
	// Any remaining content keys (no item) stay nil → black

	// Encode all keys concurrently
	blackImg := func() image.Image {
		size := n.dev.PixelSize()
		img := image.NewRGBA(image.Rect(0, 0, size, size))
		draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
		return img
	}()

	var wg sync.WaitGroup
	wg.Add(totalKeys)
	for i := 0; i < totalKeys; i++ {
		i := i
		go func() {
			defer wg.Done()
			img := images[i]
			if img == nil {
				img = blackImg
			}
			data, err := n.dev.EncodeKeyImage(img)
			frames[i].data = data
			frames[i].err = err
		}()
	}
	wg.Wait()

	// Write serially (HID is not goroutine-safe for concurrent writes)
	for _, f := range frames {
		if f.err != nil {
			return fmt.Errorf("encode key %d: %w", f.index, f.err)
		}
		if err := n.dev.WriteKeyData(f.index, f.data); err != nil {
			return fmt.Errorf("write key %d: %w", f.index, err)
		}
	}

	return nil
}

// renderReservedKeys renders the reserved column buttons (column 0).
func (n *Navigator) renderReservedKeys() {
	// Key 0 (row 0, col 0): Back button
	if !n.IsAtRoot() {
		img := n.createTextImage("<-", color.RGBA{100, 100, 100, 255})
		n.dev.SetImage(KeyBack, img)
	} else {
		// At root - show home indicator
		img := n.createTextImage("HOME", color.RGBA{50, 50, 50, 255})
		n.dev.SetImage(KeyBack, img)
	}

	// Key 5 (row 1, col 0): Toggle 1 (placeholder)
	if n.toggleStates[KeyToggle1] {
		img := n.createTextImage("T1:ON", color.RGBA{0, 150, 0, 255})
		n.dev.SetImage(KeyToggle1, img)
	} else {
		img := n.createTextImage("T1", color.RGBA{80, 80, 80, 255})
		n.dev.SetImage(KeyToggle1, img)
	}

	// Key 10 (row 2, col 0): Toggle 2 (placeholder)
	if n.toggleStates[KeyToggle2] {
		img := n.createTextImage("T2:ON", color.RGBA{0, 150, 0, 255})
		n.dev.SetImage(KeyToggle2, img)
	} else {
		img := n.createTextImage("T2", color.RGBA{80, 80, 80, 255})
		n.dev.SetImage(KeyToggle2, img)
	}
}

// HandleKeyPress handles a key press and returns the action to take.
// Returns: (item *PageItem, navigated bool, err error)
// If navigated is true, the page changed. If item is non-nil, it's an action to execute.
func (n *Navigator) HandleKeyPress(keyIndex int) (*PageItem, bool, error) {
	page, err := n.LoadPage()
	if err != nil {
		return nil, false, err
	}

	// Check if this is a reserved key (column 0)
	switch keyIndex {
	case KeyBack:
		if n.NavigateBack() {
			return nil, true, nil
		}
		return nil, false, nil

	case KeyToggle1:
		// Toggle state and re-render
		n.toggleStates[KeyToggle1] = !n.toggleStates[KeyToggle1]
		n.renderReservedKeys()
		return nil, false, nil

	case KeyToggle2:
		// Toggle state and re-render
		n.toggleStates[KeyToggle2] = !n.toggleStates[KeyToggle2]
		n.renderReservedKeys()
		return nil, false, nil
	}

	// Check if this is a content key
	for i, ck := range n.contentKeys {
		if ck == keyIndex {
			if i < len(page.Items) {
				item := &page.Items[i]
				if item.IsFolder {
					if err := n.NavigateInto(item.Path); err != nil {
						return nil, false, err
					}
					return nil, true, nil
				}
				// It's an action/script
				return item, false, nil
			}
			return nil, false, nil // Empty key
		}
	}

	return nil, false, nil
}

// GetToggleState returns the state of a toggle key.
func (n *Navigator) GetToggleState(keyIndex int) bool {
	return n.toggleStates[keyIndex]
}

// SetToggleState sets the state of a toggle key.
func (n *Navigator) SetToggleState(keyIndex int, state bool) {
	n.toggleStates[keyIndex] = state
}

// GetVisibleScripts returns a map of script paths to key indices for visible scripts.
func (n *Navigator) GetVisibleScripts() map[string]int {
	result := make(map[string]int)

	page, err := n.LoadPage()
	if err != nil {
		return result
	}

	for i, item := range page.Items {
		if i >= len(n.contentKeys) {
			break
		}
		if !item.IsFolder && item.Script != "" {
			keyIndex := n.contentKeys[i]
			result[item.Script] = keyIndex
		}
	}

	return result
}

// createTextImage creates a simple image with text.
func (n *Navigator) createTextImage(text string, bgColor color.Color) image.Image {
	return n.CreateTextImageWithColors(text, bgColor, color.White)
}

// CreateTextImageWithColors creates an image with text and custom colors.
// This is exported for use by script passive updates.
func (n *Navigator) CreateTextImageWithColors(text string, bgColor, textColor color.Color) image.Image {
	size := n.dev.PixelSize()
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Draw text centered
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: basicfont.Face7x13,
	}

	// Calculate text position (roughly centered)
	textWidth := len(text) * 7 // basicfont is ~7px wide per char
	x := (size - textWidth) / 2
	if x < 2 {
		x = 2
	}
	y := size/2 + 4 // Center vertically

	d.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
	d.DrawString(text)

	return img
}

// truncateName truncates a name to fit on a button.
func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-1] + "."
}
