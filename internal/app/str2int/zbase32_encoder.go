package str2int

import (
	"encoding/binary"
	"errors"
	"github.com/corvus-ch/zbase32"
)

type ZBase32Encoder struct {
	impl *zbase32.Encoding
}

func NewZBase32Encoder() ZBase32Encoder {
	return ZBase32Encoder{
		impl: zbase32.StdEncoding,
	}
}

func (encoder ZBase32Encoder) EncodeToString(value uint32) (string, error) {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, value)
	return encoder.impl.EncodeToString(bytes), nil
}

func (encoder ZBase32Encoder) DecodeString(str string) (uint32, error) {
	bytes, err := zbase32.StdEncoding.DecodeString(str)
	if err != nil {
		return 0, err
	}

	if len(bytes) != 4 {
		return 0, errors.New("invalid string length")
	}

	return binary.LittleEndian.Uint32(bytes), nil
}
