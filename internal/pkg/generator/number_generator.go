//go:generate mockgen -source=number_generator.go -destination=./mocks/number_generator_mocks.go -package=generatormocks

package generator

import (
	"fmt"
	"math/rand"
)

type NumberGenerator interface {
	Generate() string
}

type numberGenerator struct {
	n int
}

func NewNumberGenerator(n int) NumberGenerator {
	return &numberGenerator{n: n}
}

func (g *numberGenerator) Generate() string {
	value := rand.Intn(g.n)
	return fmt.Sprintf("%08d", value)
}

var DefaultNumberGenerator = NewNumberGenerator(99999999)
