package generator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNumberGenerator_Generate(t *testing.T) {
	t.Run("generate", func(t *testing.T) {
		g := NewNumberGenerator(9999)
		number := g.Generate()
		t.Log(number)
		assert.Equal(t, 8, len(number))
	})
}
