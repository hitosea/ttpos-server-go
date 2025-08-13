package utility

import (
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func GenItemCode(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, gonanoid.MustGenerate(alphabet, 18))
}
