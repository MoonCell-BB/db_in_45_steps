package database

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCellEncodeDecodeValI64(t *testing.T) {
	cell := Cell{Type: TypeI64, I64: -2}
	data := []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	assert.Equal(t, data, cell.EncodeVal(nil))
	decoded := Cell{Type: TypeI64}
	rest, err := decoded.DecodeVal(data)
	assert.True(t, len(rest) == 0 && err == nil)
	assert.Equal(t, cell, decoded)
}

func TestCellEncodeDecodeValStr(t *testing.T) {
	cell := Cell{Type: TypeStr, Str: []byte("asdf")}
	data := []byte{4, 0, 0, 0, 'a', 's', 'd', 'f'}
	assert.Equal(t, data, cell.EncodeVal(nil))
	decoded := Cell{Type: TypeStr}
	rest, err := decoded.DecodeVal(data)
	assert.True(t, len(rest) == 0 && err == nil)
	assert.Equal(t, cell, decoded)
}

func TestCellEncodeDecodeValStrEmpty(t *testing.T) {
	cell := Cell{Type: TypeStr, Str: []byte{}}
	data := []byte{0, 0, 0, 0}
	assert.Equal(t, data, cell.EncodeVal(nil))
	decoded := Cell{Type: TypeStr}
	rest, err := decoded.DecodeVal(data)
	assert.True(t, len(rest) == 0 && err == nil)
	assert.Empty(t, decoded.Str)
	assert.Equal(t, cell, decoded)
}

func TestCellDecodeValPartialData(t *testing.T) {
	cellI64 := Cell{Type: TypeI64}
	rest, err := cellI64.DecodeVal([]byte{0x01, 0x02, 0x03})
	assert.ErrorIs(t, err, ErrDataLen)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, rest)

	cellStr := Cell{Type: TypeStr}
	rest, err = cellStr.DecodeVal([]byte{2, 0, 0, 0, 'a'})
	assert.ErrorIs(t, err, ErrDataLen)
	assert.Equal(t, []byte{2, 0, 0, 0, 'a'}, rest)
}

func TestCellInvalidTypePanics(t *testing.T) {
	assert.Panics(t, func() {
		(&Cell{Type: CellType(99)}).EncodeVal(nil)
	})
	assert.Panics(t, func() {
		(&Cell{Type: CellType(99)}).DecodeVal([]byte{0, 0, 0, 0})
	})
	assert.Panics(t, func() {
		(&Cell{Type: CellType(99)}).EncodeKey(nil)
	})
	assert.Panics(t, func() {
		(&Cell{Type: CellType(99)}).DecodeKey([]byte{0, 0, 0, 0})
	})
}

func TestTableCell(t *testing.T) {
	cell := Cell{Type: TypeI64, I64: -2}
	data := []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	assert.Equal(t, data, cell.EncodeVal(nil))
	decoded := Cell{Type: TypeI64}
	rest, err := decoded.DecodeVal(data)
	assert.True(t, len(rest) == 0 && err == nil)
	assert.Equal(t, cell, decoded)

	cell = Cell{Type: TypeStr, Str: []byte("asdf")}
	data = []byte{4, 0, 0, 0, 'a', 's', 'd', 'f'}
	assert.Equal(t, data, cell.EncodeVal(nil))
	decoded = Cell{Type: TypeStr}
	rest, err = decoded.DecodeVal(data)
	assert.True(t, len(rest) == 0 && err == nil)
	assert.Equal(t, cell, decoded)
}

func randString() (out []byte) {
	sz := rand.Intn(256)
	for i := 0; i < sz; i++ {
		out = append(out, byte(rand.Uint32()%256))
	}
	return out
}

func TestTableCellKey(t *testing.T) {
	cell := Cell{Type: TypeI64, I64: -2}
	data := []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}
	assert.Equal(t, data, cell.EncodeKey(nil))
	decoded := Cell{Type: TypeI64}
	rest, err := decoded.DecodeKey(data)
	assert.True(t, len(rest) == 0 && err == nil)
	assert.Equal(t, cell, decoded)

	outKeys := []string{}
	for i := -2; i <= 2; i++ {
		cell = Cell{Type: TypeI64, I64: int64(i)}
		outKeys = append(outKeys, string(cell.EncodeKey(nil)))
	}
	assert.True(t, slices.IsSorted(outKeys))

	cell = Cell{Type: TypeStr, Str: []byte("a\x00s\x01d\x02f")}
	data = []byte{'a', 0x01, 0x01, 's', 0x01, 0x02, 'd', 0x02, 'f', 0}
	assert.Equal(t, data, cell.EncodeKey(nil))
	decoded = Cell{Type: TypeStr}
	rest, err = decoded.DecodeKey(data)
	assert.True(t, len(rest) == 0 && err == nil)
	assert.Equal(t, cell, decoded)

	strKeys := []string{}
	for i := 0; i < 10000; i++ {
		strKeys = append(strKeys, string(randString()))
	}
	slices.Sort(strKeys)

	outKeys = []string{}
	for _, s := range strKeys {
		cell := Cell{Type: TypeStr, Str: []byte(s)}
		outKeys = append(outKeys, string(cell.EncodeKey(nil)))

		decoded = Cell{Type: TypeStr}
		rest, err = decoded.DecodeKey([]byte(outKeys[len(outKeys)-1]))
		assert.True(t, len(rest) == 0 && err == nil && string(decoded.Str) == s)
	}
	assert.True(t, slices.IsSorted(outKeys))
}
