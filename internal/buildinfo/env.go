package buildinfo

import "runtime"

var (
	OS   = runtime.GOOS
	ARCH = runtime.GOARCH
)
