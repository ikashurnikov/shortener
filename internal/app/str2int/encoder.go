package str2int

type Encoder interface {
	EncodeToString(value uint32) (string, error)
	DecodeString(str string) (uint32, error)
}
