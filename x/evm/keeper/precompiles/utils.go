package precompiles

import (
	"errors"
	"runtime"
	"strings"
)

func CheckCaller() (bool, error) {
	layer := 3
	pc, _, _, _ := runtime.Caller(layer)
	prefix := "github.com/ethereum/go-ethereum/core/vm.(*EVM)."
	caller := runtime.FuncForPC(pc).Name()
	readonly := false
	if strings.Index(caller, prefix) == 0 {
		fn := caller[len(prefix):]
		switch fn {
		case "Call":
			readonly = false
		case "CallCode", "DelegateCall", "StaticCall":
			readonly = true
		default:
			return readonly, errors.New("unknown caller")
		}
	}
	return readonly, nil
}
