package arch

import "runtime"

func IsAmd64() bool {
	return runtime.GOARCH == "amd64"
}
