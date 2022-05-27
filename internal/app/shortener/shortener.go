package shortener

type Shortener interface {
	Encode(longURL string) (string, error)
	Decode(shortURL string) (string, error)
}
