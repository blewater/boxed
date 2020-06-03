package secret

import (
	"fmt"

	"github.com/tradeline-tech/workflow/types"
)

const (
	WorkflowNameKey = "secret"

	G    = "g"
	P    = "p"
	GtoY = "gy"
	GtoX = "gx"
	Y    = "y"
	X    = "x"
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

func GetValueNotFoundErrFunc(v string) error {
	return fmt.Errorf("%s not found within configuration", v)
}

func GetValue(config types.TaskConfiguration, key string) (int64, error) {
	interfaceVal, ok := config.Get(key)
	if !ok {
		return 0, GetValueNotFoundErrFunc(key)
	}

	val := interfaceVal.(int64)
	fmt.Println(val)

	return val, nil
}
