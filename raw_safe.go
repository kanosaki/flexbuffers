//+build !unsafe

package flexbuffers

import (
	"fmt"
	"math"
)

func (b Raw) ReadUInt64(offset int, byteWidth uint8) (uint64, error) {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return 0, ErrOutOfRange
	}
	if byteWidth < 4 {
		if byteWidth < 2 {
			return uint64(b[offset]), nil
		} else {
			return uint64(b[offset+1])<<8 | uint64(b[offset]), nil
		}
	} else {
		if byteWidth < 8 {
			return uint64(b[offset+3])<<24 | uint64(b[offset+2])<<16 | uint64(b[offset+1])<<8 | uint64(b[offset]), nil
		} else {
			return uint64(b[offset+7])<<56 | uint64(b[offset+6])<<48 | uint64(b[offset+5])<<40 | uint64(b[offset+4])<<32 |
				uint64(b[offset+3])<<24 | uint64(b[offset+2])<<16 | uint64(b[offset+1])<<8 | uint64(b[offset]), nil
		}
	}
}

func (b Raw) ReadInt64(offset int, byteWidth uint8) (int64, error) {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return 0, ErrOutOfRange
	}
	if byteWidth < 4 {
		if byteWidth < 2 {
			return int64(int8(b[offset])), nil
		} else {
			return int64(int16(b[offset+1])<<8 | int16(b[offset])), nil
		}
	} else {
		if byteWidth < 8 {
			return int64(int32(b[offset+3])<<24 | int32(b[offset+2])<<16 | int32(b[offset+1])<<8 | int32(b[offset])), nil
		} else {
			return int64(b[offset+7])<<56 | int64(b[offset+6])<<48 | int64(b[offset+5])<<40 | int64(b[offset+4])<<32 |
				int64(b[offset+3])<<24 | int64(b[offset+2])<<16 | int64(b[offset+1])<<8 | int64(b[offset]), nil
		}
	}
}

func (b Raw) ReadDouble(offset int, byteWidth uint8) (float64, error) {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return 0.0, ErrOutOfRange
	}
	if byteWidth < 4 {
		if byteWidth < 2 {
			return 0.0, fmt.Errorf("float8 is not supported")
		} else {
			return 0.0, fmt.Errorf("float16 is not supported")
		}
	} else {
		ui, err := b.ReadUInt64(offset, byteWidth)
		if err != nil {
			return 0.0, err
		}
		if byteWidth < 8 {
			return float64(math.Float32frombits(uint32(ui))), nil
		} else {
			return math.Float64frombits(ui), nil
		}
	}
}

func (b Raw) WriteInt64(offset int, byteWidth uint8, value int64) error {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return ErrOutOfRange
	}
	valueWidth := WidthI(value)
	fits := (1 << valueWidth) <= byteWidth
	if !fits {
		return ErrUpdateDoesntFit
	}
	if valueWidth <= BitWidth8 {
		b[offset] = byte(value)
	}
	if valueWidth <= BitWidth16 {
		b[offset+1] = byte(value >> 8)
	}
	if valueWidth <= BitWidth32 {
		b[offset+2] = byte(value >> 16)
		b[offset+3] = byte(value >> 24)
	}
	if valueWidth <= BitWidth64 {
		b[offset+4] = byte(value >> 32)
		b[offset+5] = byte(value >> 40)
		b[offset+6] = byte(value >> 48)
		b[offset+7] = byte(value >> 56)
	}
	return nil
}

func (b Raw) WriteUInt64(offset int, byteWidth uint8, value uint64) error {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return ErrOutOfRange
	}
	valueWidth := WidthU(value)
	fits := (1 << valueWidth) <= byteWidth
	if !fits {
		return ErrUpdateDoesntFit
	}
	if valueWidth <= BitWidth8 {
		b[offset] = byte(value)
	}
	if valueWidth <= BitWidth16 {
		b[offset+1] = byte(value >> 8)
	}
	if valueWidth <= BitWidth32 {
		b[offset+2] = byte(value >> 16)
		b[offset+3] = byte(value >> 24)
	}
	if valueWidth <= BitWidth64 {
		b[offset+4] = byte(value >> 32)
		b[offset+5] = byte(value >> 40)
		b[offset+6] = byte(value >> 48)
		b[offset+7] = byte(value >> 56)
	}
	return nil
}

func (b Raw) WriteFloat(offset int, byteWidth uint8, value float64) error {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return ErrOutOfRange
	}
	valueWidth := WidthF(value)
	fits := (1 << valueWidth) <= byteWidth
	if !fits {
		return ErrUpdateDoesntFit
	}
	if byteWidth == 4 {
		bits := math.Float32bits(float32(value))
		b[offset] = byte(bits)
		b[offset+1] = byte(bits >> 8)
		b[offset+2] = byte(bits >> 16)
		b[offset+3] = byte(bits >> 24)
	} else if byteWidth == 8 {
		bits := math.Float64bits(value)
		b[offset] = byte(bits)
		b[offset+1] = byte(bits >> 8)
		b[offset+2] = byte(bits >> 16)
		b[offset+3] = byte(bits >> 24)
		b[offset+4] = byte(bits >> 32)
		b[offset+5] = byte(bits >> 40)
		b[offset+6] = byte(bits >> 48)
		b[offset+7] = byte(bits >> 56)
	}
	return nil
}

func (b Raw) Indirect(offset int, byteWidth uint8) (int, error) {
	if len(b) <= offset+int(byteWidth) || offset < 0 {
		return 0, ErrOutOfRange
	}
	var ind int
	if byteWidth < 4 {
		if byteWidth < 2 {
			ind = offset - int(b[offset])
		} else {
			ind = offset - int(uint64(b[offset+1])<<8|uint64(b[offset]))
		}
	} else {
		if byteWidth < 8 {
			ind = offset - int(uint64(b[offset+3])<<24|uint64(b[offset+2])<<16|uint64(b[offset+1])<<8|uint64(b[offset]))
		} else {
			ind = offset - int(
				uint64(b[offset+7])<<56|uint64(b[offset+6])<<48|uint64(b[offset+5])<<40|uint64(b[offset+4])<<32|
					uint64(b[offset+3])<<24|uint64(b[offset+2])<<16|uint64(b[offset+1])<<8|uint64(b[offset]),
			)
		}
	}
	if ind < 0 || len(b) <= ind {
		return 0, ErrOutOfRange
	}
	return ind, nil
}
