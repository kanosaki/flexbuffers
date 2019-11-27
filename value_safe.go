//+build !unsafe

package flexbuffers

import (
	"math"
)

func newValueUInt(u uint64, t Type, bw BitWidth) value {
	return value{d: int64(u), typ: t, minBitWidth: bw}
}

func newValueInt(u int64, t Type, bw BitWidth) value {
	return value{d: u, typ: t, minBitWidth: bw}
}

func newValueFloat32(u float32) value {
	f := float64(u)
	return value{d: int64(math.Float64bits(f)), typ: FBTFloat, minBitWidth: BitWidth32}
}

func newValueFloat64(u float64) value {
	return value{d: int64(math.Float64bits(u)), typ: FBTFloat, minBitWidth: WidthF(u)}
}
