package rc

import "sync"

// Copy from net/http/server.go.
const copyBufPoolSize = 32 * 1024

var copyBufPool = sync.Pool{New: func() any { return new([copyBufPoolSize]byte) }}

func getCopyBuf() []byte { //nostyle:getters
	buf := copyBufPool.Get().(*[copyBufPoolSize]byte) //nolint:errcheck
	return buf[:]
}
func putCopyBuf(b []byte) {
	if len(b) != copyBufPoolSize {
		panic("trying to put back buffer of the wrong size in the copyBufPool") //nostyle:dontpanic
	}
	copyBufPool.Put((*[copyBufPoolSize]byte)(b))
}
