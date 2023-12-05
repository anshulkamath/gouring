package gouring

import (
	"syscall"
)

const MmapFlags = syscall.MAP_SHARED | syscall.MAP_POPULATE
