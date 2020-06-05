package secret

import (
	"fmt"

	"github.com/tradeline-tech/workflow/types"
)

const (
	WorkflowNameKey = "secret"

	G            = "g"
	P            = "p"
	Y            = "y"
	X            = "x"
	GtoY         = "gy"
	GtoX         = "gx"
	GXtoY        = "gxy"
	GYtoX        = "gyx"
	IsSecretSame = "eq"
)

// GetModOfPow returns the modulo of raising an integer ^ exponent
// in a simple and moderately optimized manner using the property
// i^e%n => ((r1=1^i mod n) * (r2=r1^i mod n)* ... * (rexp^i mod n)) mod n
func GetModOfPow(integer, exponent, n int64) int64 {
	var res int64 = 1
	for i := res; i <= exponent; i++ {
		res = (res * integer) % n
	}

	return res
}

// IsSame answers whether the args are equal and prints it so.
func IsSame(gyx, gxy int64) bool {
	fmt.Println("g^xy :", gxy, "g^yx :", gyx, " equal?", gxy == gyx)

	return gyx == gxy
}

func GetValue(cfg types.TaskConfiguration, key string) (int64, error) {
	int64Val, err := cfg.GetInt64(key)
	if err != nil {
		return 0, err
	}

	// if key != X && key != Y && key != GXtoY && key != GYtoX {
	// 	fmt.Printf("\thint: %s : %d\n", key, int64Val)
	// }

	return int64Val, nil
}
