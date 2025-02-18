package unit_test

import (
	"fmt"
	"testing"
	"ttpos-server-go/pkg/auth"
)

func TestToken(t *testing.T) {
	// Example usage
	//http://127.0.0.1:8888/scan/#/?token=
	token := "YTM3MzIzYWYzYzZkMzYyMWJiNzcxZDRjZjgwYTBkNGRhNDk1ZGFlZmE5ZDE1M2E0MzNiNDg2NjMyNGY2ZjIxZi5leUpoSWpvNU5Ea3dOVGsyTVRZeU1EUTRNREF3TENKMElqb2lPVFE1TVRJd09EWTJORFkwT1RjeU9DSXNJbkVpT2lJeU9EUXlPVFlpZlE9PQ=="
	result, err := auth.DecodeDeskToken(token)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Decoded token data:", result)
	}
}
