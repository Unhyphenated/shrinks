package encoding

import (
	"fmt"
	"testing"
)

func TestMyLogic(t *testing.T) {
	// Call your function here
	for i := 0; i < 10; i++ {
		u, err := GenerateCode()
		if err != nil {
			fmt.Println("Err is : %w", err)
		}
		fmt.Println(string(u))
	}
}
