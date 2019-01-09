package protocol

import (
	"testing"
	"unsafe"
)

func TestOrderMsgSize(t *testing.T) {

	var order OrderbookMsg
	const orderSize = unsafe.Sizeof(order)
	t.Logf("Size: %d", orderSize)
}
