package render

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BryanDGuy/halo/internal/grid"
)

var gradient = []rune{' ', '░', '▒', '▓', '█'}

// out buffers each frame so the entire content lands in one write(2) syscall.
var out = bufio.NewWriterSize(os.Stdout, 1<<17) // 128 KB

func toChar(v, scale float64) rune {
	if scale <= 0 || v <= 0 {
		return gradient[0]
	}
	norm := v / scale
	if norm >= 1 {
		return gradient[len(gradient)-1]
	}
	idx := int(norm * float64(len(gradient)-1))
	return gradient[idx]
}

// FrameTo writes the heatmap to w, normalizing to the grid's current max value.
func FrameTo(w io.Writer, g *grid.Grid) {
	var peak float64
	for _, v := range g.Data {
		if v > peak {
			peak = v
		}
	}
	var sb strings.Builder
	for r := range g.Rows {
		for c := range g.Cols {
			sb.WriteRune(toChar(g.At(r, c), peak))
		}
		sb.WriteByte('\n')
	}
	fmt.Fprint(w, sb.String())
}

// Init switches to the alternate screen buffer, clears it, and hides the cursor.
// The alternate screen preserves the user's scrollback and restores it on Cleanup,
// leaving no artifacts in the terminal after exit.
func Init() {
	fmt.Fprint(os.Stdout, "\033[?1049h\033[2J\033[H\033[?25l")
}

// Cleanup restores the cursor and switches back to the primary screen buffer.
func Cleanup() {
	fmt.Fprint(os.Stdout, "\033[?25h\033[?1049l")
}

// Frame renders g and status in-place. \033[H moves the cursor to the top-left
// of the alternate screen on every call — no line-counting needed. The full frame
// is assembled in a strings.Builder then written to a bufio.Writer so the OS
// receives it in a single write(2) syscall, preventing partial-frame flicker.
func Frame(g *grid.Grid, status string) {
	var peak float64
	for _, v := range g.Data {
		if v > peak {
			peak = v
		}
	}

	var sb strings.Builder
	sb.WriteString("\033[H")
	for r := range g.Rows {
		for c := range g.Cols {
			sb.WriteRune(toChar(g.At(r, c), peak))
		}
		sb.WriteByte('\n')
	}
	sb.WriteString("\r")
	sb.WriteString(status)
	sb.WriteString("\033[K")

	if _, err := out.WriteString(sb.String()); err != nil {
		return
	}
	if err := out.Flush(); err != nil {
		return
	}
}
