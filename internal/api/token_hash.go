package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const sepSymbol = "#"

type HashToken struct {
	Secret []byte
}

func NewHMACHashToken(secret string) *HashToken {
	return &HashToken{Secret: []byte(secret)}
}

func (tk *HashToken) Create(session string, tokenExpTime int64) (string, error) {
	h := hmac.New(sha256.New, tk.Secret)
	data := fmt.Sprintf("%s:%d", session, tokenExpTime)
	h.Write([]byte(data))
	token := hex.EncodeToString(h.Sum(nil)) + sepSymbol + strconv.FormatInt(tokenExpTime, 10)
	return token, nil
}

func (tk *HashToken) Check(session, inputToken string) (bool, error) {
	tokenData := strings.Split(inputToken, sepSymbol)
	if len(tokenData) != 2 {
		return false, fmt.Errorf("bad token data")
	}

	tokenExp, err := strconv.ParseInt(tokenData[1], 10, 64)
	if err != nil {
		return false, fmt.Errorf("bad token time")
	}

	if tokenExp < time.Now().Unix() {
		return false, fmt.Errorf("token expired")
	}

	h := hmac.New(sha256.New, tk.Secret)
	data := fmt.Sprintf("%s:%d", session, tokenExp)
	h.Write([]byte(data))
	expectedMAC := h.Sum(nil)
	messageMAC, err := hex.DecodeString(tokenData[0])
	if err != nil {
		return false, fmt.Errorf("cand hex decode token")
	}

	return hmac.Equal(messageMAC, expectedMAC), nil
}
