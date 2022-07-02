package urlencoder

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

func (encoder ZBase32Encoder) Shorten(id uint32) (string, error) {
	var bytes [4]byte
	binary.LittleEndian.PutUint32(bytes[:], uint32(id))
	return encoder.impl.EncodeToString(bytes[:]), nil
}

func (encoder ZBase32Encoder) Expand(shortURL string) (uint32, error) {
	bytes, err := zbase32.StdEncoding.DecodeString(shortURL)
	if err != nil {
		return 0, err
	}

	if len(bytes) != 4 {
		return 0, errors.New("invalid string length")
	}

	return binary.LittleEndian.Uint32(bytes), nil
}
