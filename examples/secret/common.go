package secret

import (
	"fmt"
	"strconv"
	"strings"

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

func GetValueNotFoundErrFunc(v string) error {
	return fmt.Errorf("%s not found within configuration", strings.TrimSpace(v))
}

func GetValue(config types.TaskConfiguration, key string) (int64, error) {
	interfaceVal, ok := config.Get(key)
	if !ok {
		return 0, GetValueNotFoundErrFunc(key)
	}

	var int64Val int64
	switch v := interfaceVal.(type) {
	case int64:
		int64Val = v
	case int:
		int64Val = int64(v)
	case string:
		intVal, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		int64Val = int64(intVal)
	default:
		return 0, fmt.Errorf("%v of type %T", interfaceVal, interfaceVal)
	}

	if key != X && key != Y && key != GXtoY && key != GYtoX {
		fmt.Printf("\thint: %s : %d\n", key, int64Val)
	}

	return int64Val, nil
}
