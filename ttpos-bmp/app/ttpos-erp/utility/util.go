package utility

import (
	"fmt"
	"ttpos-bmp/utility/uuid"
)

func GenItemCode(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, uuid.MustGetID())
}
