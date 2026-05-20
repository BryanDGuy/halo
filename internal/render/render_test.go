package render_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/BryanDGuy/halo/internal/render"
)

func TestFrameShape(t *testing.T) {
	g := grid.New(4, 8)
	var buf bytes.Buffer
	render.FrameTo(&buf, g)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 4 {
		t.Errorf("want 4 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if len([]rune(line)) != 8 {
			t.Errorf("line %d: want 8 chars, got %d", i, len([]rune(line)))
		}
	}
}

func TestFrameCharacters(t *testing.T) {
	g := grid.New(1, 3)
	g.Set(0, 0, 0.0)  // space
	g.Set(0, 1, 0.5)  // mid-range
	g.Set(0, 2, 1.0)  // full block

	var buf bytes.Buffer
	render.FrameTo(&buf, g)
	line := strings.TrimRight(buf.String(), "\n")
	runes := []rune(line)
	if runes[0] != ' ' {
		t.Errorf("0.0 should render as space, got %q", runes[0])
	}
	if runes[1] != '▒' {
		t.Errorf("0.5 should render as ▒, got %q", runes[1])
	}
	if runes[2] != '█' {
		t.Errorf("1.0 should render as █, got %q", runes[2])
	}
}
