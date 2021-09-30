package myio

const (
	MASK8 uint = 0x0000_0000_0000_00FF
)

func UintToByteSlice(v uint) (bs []byte) {
	bs = make([]byte, 8)
	for i := 0; i < 8; i++ {
		bs[i] = byte(v & MASK8)
		v >>= 8
	}
	return
}

func ByteSliceToUint(bs []byte) uint {
	var val uint
	base := 0
	for _, v := range bs {
		val |= uint(v) << base
		base += 8
	}
	return val
}