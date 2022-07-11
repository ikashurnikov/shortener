package service

import (
	"encoding/binary"
	"github.com/corvus-ch/zbase32"
	"github.com/ikashurnikov/shortener/internal/app/model"
)

type ZBase32LinkIDEncoder struct {
	impl *zbase32.Encoding
}

func NewZBase32LinkIDEncoder() *ZBase32LinkIDEncoder {
	return &ZBase32LinkIDEncoder{
		impl: zbase32.StdEncoding,
	}
}

func (e *ZBase32LinkIDEncoder) EncodeToString(linkID model.LinkID) (string, error) {
	var bytes [4]byte
	binary.LittleEndian.PutUint32(bytes[:], uint32(linkID))
	return e.impl.EncodeToString(bytes[:]), nil
}

func (e *ZBase32LinkIDEncoder) DecodeFromString(str string) (model.LinkID, error) {
	bytes, err := zbase32.StdEncoding.DecodeString(str)
	if err != nil {
		return 0, err
	}

	if len(bytes) != 4 {
		return 0, model.ErrDecodingShortURL
	}

	return model.LinkID(binary.LittleEndian.Uint32(bytes)), nil
}
