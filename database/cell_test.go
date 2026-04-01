package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCellEncodeDecodeI64(t *testing.T) {
	cell := Cell{Type: TypeI64, I64: -2}
	data := []byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	assert.Equal(t, data, (&cell).Encode(nil))

	decoded := Cell{Type: TypeI64}
	rest, err := (&decoded).Decode(data)
	assert.NoError(t, err)
	assert.Empty(t, rest)
	assert.Equal(t, cell, decoded)
}

func TestCellEncodeDecodeStr(t *testing.T) {
	cell := Cell{Type: TypeStr, Str: []byte("asdf")}
	data := []byte{4, 0, 0, 0, 'a', 's', 'd', 'f'}
	assert.Equal(t, data, (&cell).Encode(nil))

	decoded := Cell{Type: TypeStr}
	rest, err := (&decoded).Decode(data)
	assert.NoError(t, err)
	assert.Empty(t, rest)
	assert.Equal(t, cell, decoded)
}

func TestCellEncodeDecodeStrEmpty(t *testing.T) {
	cell := Cell{Type: TypeStr, Str: []byte{}}
	data := []byte{0, 0, 0, 0}
	assert.Equal(t, data, (&cell).Encode(nil))

	decoded := Cell{Type: TypeStr}
	rest, err := (&decoded).Decode(data)
	assert.NoError(t, err)
	assert.Empty(t, rest)
	assert.Empty(t, decoded.Str)
	assert.Equal(t, cell, decoded)
}

func TestCellDecodePartialData(t *testing.T) {
	cellI64 := Cell{Type: TypeI64}
	rest, err := cellI64.Decode([]byte{0x01, 0x02, 0x03})
	assert.ErrorIs(t, err, ErrDataLen)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, rest)

	cellStr := Cell{Type: TypeStr}
	rest, err = cellStr.Decode([]byte{2, 0, 0, 0, 'a'})
	assert.ErrorIs(t, err, ErrDataLen)
	assert.Equal(t, []byte{2, 0, 0, 0, 'a'}, rest)
}

func TestCellInvalidTypePanics(t *testing.T) {
	assert.Panics(t, func() {
		(&Cell{Type: CellType(99)}).Encode(nil)
	})

	assert.Panics(t, func() {
		(&Cell{Type: CellType(99)}).Decode([]byte{0, 0, 0, 0})
	})
}
