package randgen

import (
	"crypto/rand"
	"fmt"
)

func Number(length int) (string, error) {
	const chars = "0123456789"
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("rand string generator error : %+v", err)
	}
	charsLength := len(chars)
	for i := 0; i < length; i++ {
		buffer[i] = chars[int(buffer[i])%charsLength]
	}
	return string(buffer), nil
}

func Alphabet(length int) (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("rand string generator error : %+v", err)
	}
	charsLength := len(chars)
	for i := 0; i < length; i++ {
		buffer[i] = chars[int(buffer[i])%charsLength]
	}
	return string(buffer), nil
}

