package streamdeck

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"sort"

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

// Navigator manages folder-based navigation on a Stream Deck.
type Navigator struct {
	dev        *Device
	rootPath   string
	currentDir string
	pageIndex  int
}

// NewNavigator creates a new navigator for the given device and root config path.
func NewNavigator(dev *Device, rootPath string) *Navigator {
	return &Navigator{
		dev:        dev,
		rootPath:   rootPath,
		currentDir: rootPath,
		pageIndex:  0,
	}
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
		// Skip hidden files
		if entry.Name()[0] == '.' {
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

	// Calculate pagination
	keysAvailable := n.dev.Keys()
	if !n.IsAtRoot() {
		keysAvailable-- // Reserve one key for "back" button
	}

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
func (n *Navigator) RenderPage() error {
	page, err := n.LoadPage()
	if err != nil {
		return err
	}

	// Clear all keys first
	if err := n.dev.Clear(); err != nil {
		return fmt.Errorf("clear: %w", err)
	}

	keyIndex := 0

	// If not at root, first key is "back"
	if !n.IsAtRoot() {
		img := n.createTextImage("<-", color.RGBA{100, 100, 100, 255})
		if err := n.dev.SetImage(keyIndex, img); err != nil {
			return fmt.Errorf("set back button: %w", err)
		}
		keyIndex++
	}

	// Render page items
	for _, item := range page.Items {
		if keyIndex >= n.dev.Keys() {
			break
		}

		var img image.Image
		if item.IsFolder {
			// Folder: blue background
			img = n.createTextImage(truncateName(item.Name, 8), color.RGBA{30, 80, 180, 255})
		} else {
			// Script/action: green background
			img = n.createTextImage(truncateName(item.Name, 8), color.RGBA{30, 130, 80, 255})
		}

		if err := n.dev.SetImage(keyIndex, img); err != nil {
			return fmt.Errorf("set key %d: %w", keyIndex, err)
		}
		keyIndex++
	}

	return nil
}

// HandleKeyPress handles a key press and returns the action to take.
// Returns: (item *PageItem, navigated bool, err error)
// If navigated is true, the page changed. If item is non-nil, it's an action to execute.
func (n *Navigator) HandleKeyPress(keyIndex int) (*PageItem, bool, error) {
	page, err := n.LoadPage()
	if err != nil {
		return nil, false, err
	}

	adjustedIndex := keyIndex

	// If not at root, first key is "back"
	if !n.IsAtRoot() {
		if keyIndex == 0 {
			n.NavigateBack()
			return nil, true, nil
		}
		adjustedIndex = keyIndex - 1
	}

	// Check if this is a valid item
	if adjustedIndex < 0 || adjustedIndex >= len(page.Items) {
		return nil, false, nil
	}

	item := &page.Items[adjustedIndex]

	if item.IsFolder {
		if err := n.NavigateInto(item.Path); err != nil {
			return nil, false, err
		}
		return nil, true, nil
	}

	// It's an action/script
	return item, false, nil
}

// createTextImage creates a simple image with text.
func (n *Navigator) createTextImage(text string, bgColor color.Color) image.Image {
	size := n.dev.PixelSize()
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Draw text centered
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
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
