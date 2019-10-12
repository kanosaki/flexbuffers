package flexbuffers

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWidthU(t *testing.T) {
	a := assert.New(t)
	a.Equal(BitWidth8, WidthU(0))
	a.Equal(BitWidth8, WidthU(math.MaxUint8))
	a.Equal(BitWidth16, WidthU(math.MaxUint8+1))
	a.Equal(BitWidth16, WidthU(math.MaxUint16))
	a.Equal(BitWidth32, WidthU(math.MaxUint16+1))
	a.Equal(BitWidth32, WidthU(math.MaxUint32))
	a.Equal(BitWidth64, WidthU(math.MaxUint32+1))
	a.Equal(BitWidth64, WidthU(math.MaxUint64))
}

func TestWidthI(t *testing.T) {
	a := assert.New(t)
	a.Equal(BitWidth8, WidthI(0))
	a.Equal(BitWidth8, WidthI(math.MaxInt8))
	a.Equal(BitWidth16, WidthI(math.MaxInt8+1))
	a.Equal(BitWidth16, WidthI(math.MaxInt16))
	a.Equal(BitWidth32, WidthI(math.MaxInt16+1))
	a.Equal(BitWidth32, WidthI(math.MaxInt32))
	a.Equal(BitWidth64, WidthI(math.MaxInt32+1))
	a.Equal(BitWidth64, WidthI(math.MaxInt64))

	a.Equal(BitWidth8, WidthI(math.MinInt8))
	a.Equal(BitWidth16, WidthI(math.MinInt8-1))
	a.Equal(BitWidth16, WidthI(math.MinInt16))
	a.Equal(BitWidth32, WidthI(math.MinInt16-1))
	a.Equal(BitWidth32, WidthI(math.MinInt32))
	a.Equal(BitWidth64, WidthI(math.MinInt32-1))
	a.Equal(BitWidth64, WidthI(math.MinInt64))
}
