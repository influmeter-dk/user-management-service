package main

import (
	"crypto/rand"
	"fmt"
	"log"
	b64 "encoding/base64"
)

func main() {
	keyLength := 32 // bytes

	secret := make([]byte, keyLength)

	_, err := rand.Read(secret)
	if err != nil {
		log.Fatal(err)
	}
	secretStr := b64.StdEncoding.EncodeToString(secret)
	fmt.Println(secretStr)
}
