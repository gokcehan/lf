package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v3"
)

// imageExtensions lists file extensions that are treated as images for
// built-in Kitty protocol previews.
var imageExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
}

func isImageFile(path string) bool {
	return imageExtensions[strings.ToLower(filepath.Ext(path))]
}

type kittyScreen struct {
	lastFile   string
	lastWin    win
	forceClear bool
}

func (ks *kittyScreen) clearKitty(win *win, screen tcell.Screen, filePath string) {
	if ks.lastFile != "" && (filePath != ks.lastFile || *win != ks.lastWin || ks.forceClear) {
		// Delete all kitty images so they don't linger on screen
		// when the preview changes.
		fmt.Fprint(os.Stderr, "\033_Ga=d,d=a,q=2;\033\\")
	}
}

func (ks *kittyScreen) printKitty(win *win, screen tcell.Screen, reg *reg) {
	if reg.path == ks.lastFile && *win == ks.lastWin && !ks.forceClear {
		return
	}

	// Unlock any region locked by a previous kitty render so tcell
	// can redraw the full pane before we place the new image.
	if ks.lastFile != "" {
		screen.LockRegion(ks.lastWin.x, ks.lastWin.y, ks.lastWin.w, ks.lastWin.h, false)
	}

	cw, ch, err := cellSize(screen)
	if err != nil {
		cw, ch = 10, 20
	}

	y := win.y
	var b strings.Builder

	// Collect consecutive Kitty frames so that chunked transmission
	// (m=1 / m=0) is written as a single logical image at one position.
	var kittyBuf []string
	flushKitty := func() {
		if len(kittyBuf) == 0 {
			return
		}
		// Use the first frame that has dimension info for sizing.
		sw, sh := 0, 0
		for _, k := range kittyBuf {
			sw, sh = kittyCellSize(k, cw, ch)
			if sw > 0 && sh > 0 {
				break
			}
		}
		if sw <= 0 {
			sw = win.w
		}
		if sh <= 0 {
			sh = 1
		}

		for i, k := range kittyBuf {
			if i == 0 {
				fmt.Fprintf(&b, "\033[%d;%dH", y+1, win.x+1)
			}
			b.WriteString(k)
		}
		screen.LockRegion(win.x, y, sw, sh, true)
		y += sh
		kittyBuf = nil
	}

	for _, line := range reg.lines {
		if !strings.HasPrefix(line, "\033_G") {
			flushKitty()
			if y >= win.y+win.h {
				break
			}
			line = sanitizePreview(line)
			fmt.Fprintf(&b, "\033[%d;%dH", y+1, win.x+1)
			b.WriteString(line)
			y++
			continue
		}
		kittyBuf = append(kittyBuf, line)
	}
	flushKitty()

	// Write all output directly to stderr with synchronized update
	// so the image renders atomically without flickering.
	fmt.Fprint(os.Stderr, "\033[?2026h")
	fmt.Fprint(os.Stderr, "\0337")
	screen.Show()
	fmt.Fprint(os.Stderr, b.String())
	fmt.Fprint(os.Stderr, "\0338")
	fmt.Fprint(os.Stderr, "\033[?2026l")

	ks.lastFile = reg.path
	ks.lastWin = *win
	ks.forceClear = false
}

// kittyCellSize parses a Kitty graphics APC command to extract the display
// size in terminal cells. The command has the form:
//
//	\033_G<key=value,...>;<payload>\033\\
//
// S=/c=cols and V=/r=rows give the cell-based dimensions directly. If those
// are absent, s=w and v=h (pixel dimensions) are converted using cw and ch.
func kittyCellSize(cmd string, cw, ch int) (int, int) {
	if cw <= 0 {
		cw = 10
	}
	if ch <= 0 {
		ch = 20
	}

	// The control section is between "\033_G" and the first ';'.
	// "\033_G" is 3 bytes (\033, _, G).
	start := strings.IndexByte(cmd, ';')
	if start < 0 {
		return 1, 1
	}
	control := cmd[3:start]

	var sc, sr int // S= / V= (cells)
	var pw, ph int // s= / v= (pixels)

	for kv := range strings.SplitSeq(control, ",") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		switch k {
		case "S", "c":
			sc, _ = strconv.Atoi(v)
		case "V", "r":
			sr, _ = strconv.Atoi(v)
		case "s":
			pw, _ = strconv.Atoi(v)
		case "v":
			ph, _ = strconv.Atoi(v)
		}
	}

	if sc > 0 && sr > 0 {
		return sc, sr
	}
	if pw > 0 && ph > 0 {
		return (pw + cw - 1) / cw, (ph + ch - 1) / ch
	}

	return 0, 0
}

// generateKittyPreview builds a Kitty protocol preview for the image file at
// path. It decodes the image, scales it to fit the preview window, encodes it
// as PNG, and returns the Kitty APC command as a single line.
func generateKittyPreview(path string, win *win) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	bounds := img.Bounds()
	iw, ih := bounds.Dx(), bounds.Dy()
	if iw <= 0 || ih <= 0 {
		return nil, fmt.Errorf("invalid image dimensions: %dx%d", iw, ih)
	}

	// Estimate cell size (the same fallback used by sixel). The preview
	// goroutine does not have access to the tcell Screen, so we use the
	// historically safe defaults of 10×20 pixels per cell.
	const estCellW = 10
	const estCellH = 20

	maxW := win.w
	maxH := win.h

	// Compute the number of cells the image would occupy at its natural size.
	natCW := (iw + estCellW - 1) / estCellW
	natCH := (ih + estCellH - 1) / estCellH

	// Scale down to fit the preview window.
	scale := 1.0
	if natCW > maxW {
		scale = float64(maxW) / float64(natCW)
	}
	if float64(natCH)*scale > float64(maxH) {
		scale = float64(maxH) / float64(natCH)
	}

	targetW := max(int(float64(iw)*scale), 1)
	targetH := max(int(float64(ih)*scale), 1)

	// Resize using nearest-neighbour (fast, no extra dependencies).
	var resized image.Image
	if targetW == iw && targetH == ih {
		resized = img
	} else {
		resized = resizeNearest(img, targetW, targetH)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return nil, fmt.Errorf("encoding PNG: %w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	displayCW := (targetW + estCellW - 1) / estCellW
	displayCH := (targetH + estCellH - 1) / estCellH

	if displayCW > maxW {
		displayCW = maxW
	}
	if displayCH > maxH {
		displayCH = maxH
	}

	cmd := fmt.Sprintf(
		"\033_Ga=T,f=100,s=%d,v=%d,S=%d,V=%d,C=1,q=2;%s\033\\",
		targetW, targetH, displayCW, displayCH, b64,
	)

	return []string{cmd}, nil
}

// resizeNearest returns a new RGBA image that is a nearest-neighbour scaled
// copy of src.
func resizeNearest(src image.Image, dstW, dstH int) *image.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	sr := src.Bounds()
	sw, sh := sr.Dx(), sr.Dy()

	for y := range dstH {
		srcY := y * sh / dstH
		for x := range dstW {
			srcX := x * sw / dstW
			rgba.Set(x, y, src.At(sr.Min.X+srcX, sr.Min.Y+srcY))
		}
	}
	return rgba
}
