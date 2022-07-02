package urlencoder

type Encoder interface {
	Shorten(id uint32) (string, error)
	Expand(shortURL string) (uint32, error)
}
