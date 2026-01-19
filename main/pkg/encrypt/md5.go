package encrypt

import (
	"crypto/md5"
	"fmt"
)

func CompanyMd5(companyUuid uint64, data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d%s", companyUuid, data))))
}
