package render_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/BryanDGuy/halo/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameShape(t *testing.T) {
	g := grid.New(4, 8)
	var buf bytes.Buffer
	render.FrameTo(&buf, g)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Len(t, lines, 4)
	for i, line := range lines {
		assert.Len(t, []rune(line), 8, "line %d width", i)
	}
}

func TestFrameCharacters(t *testing.T) {
	g := grid.New(1, 3)
	g.Set(0, 0, 0.0) // space
	g.Set(0, 1, 0.5) // mid-range
	g.Set(0, 2, 1.0) // full block

	var buf bytes.Buffer
	render.FrameTo(&buf, g)
	runes := []rune(strings.TrimRight(buf.String(), "\n"))

	assert.Equal(t, ' ', runes[0], "0.0 should render as space")
	assert.Equal(t, '▒', runes[1], "0.5 should render as ▒")
	assert.Equal(t, '█', runes[2], "1.0 should render as █")
}
