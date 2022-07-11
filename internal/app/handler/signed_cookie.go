package handler

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"strconv"
)

type SignedCookie struct {
	key string
}

func NewSignedCookie(key string) SignedCookie {
	return SignedCookie{key: key}
}

func (c SignedCookie) Get(req *http.Request, name string) (string, bool) {
	cookie, err := req.Cookie(name)
	if err != nil {
		return "", false
	}

	signedCookie, err := req.Cookie(signedCookieName(name))
	if err != nil {
		return "", false
	}

	if c.verify(c.sign(cookie.Value), signedCookie.Value) {
		return cookie.Value, true
	}
	return "", false
}

func (c SignedCookie) GetInt(req *http.Request, name string) (int, bool) {
	value, ok := c.Get(req, name)
	if !ok {
		return 0, false
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return i, true
}

func (c SignedCookie) Set(rw http.ResponseWriter, name, value string) {
	http.SetCookie(rw, &http.Cookie{
		Name:  name,
		Value: value,
	})
	http.SetCookie(rw, &http.Cookie{
		Name:  signedCookieName(name),
		Value: c.sign(value),
	})
}

func (c SignedCookie) SetInt(rw http.ResponseWriter, name string, value int) {
	c.Set(rw, name, strconv.Itoa(value))
}

func (c *SignedCookie) Add(req *http.Request, name, value string) {
	req.AddCookie(&http.Cookie{Name: name, Value: value})
	req.AddCookie(&http.Cookie{Name: signedCookieName(name), Value: c.sign(value)})
}

func (c *SignedCookie) AddInt(req *http.Request, name string, value int) {
	c.Add(req, name, strconv.Itoa(value))
}

func (c *SignedCookie) sign(data string) string {
	h := hmac.New(sha1.New, []byte(c.key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (c SignedCookie) verify(data string, checkSum string) bool {
	if data == "" || checkSum == "" {
		return false
	}

	hexData, err := hex.DecodeString(data)
	if err != nil {
		return false
	}

	checkSumHex, err := hex.DecodeString(checkSum)
	if err != nil {
		return false
	}

	return hmac.Equal(hexData, checkSumHex)
}

func signedCookieName(name string) string {
	return name + ".sign"
}
