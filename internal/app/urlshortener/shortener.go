package urlshortener

type Shortener interface {
	EncodeLongURL(longURL string) (string, error)
	DecodeShortURL(shortURL string) (string, error)
}
