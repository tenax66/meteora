package shared

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

func ReadKeys(key_path string, pubkey_path string) (ed25519.PrivateKey, ed25519.PublicKey, error) {

	// read a secret key
	privateKeyPEM, err := os.ReadFile(key_path)
	if err != nil {
		log.Println("Error reading private key:", err)
		return nil, nil, err
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Println("Invalid private key")
		return nil, nil, err
	}

	// PEM blocks of type "PRIVATE KEY"
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Println("Error parsing private key:", err)
		return nil, nil, err
	}

	// read a public key
	publicKeyPEM, err := os.ReadFile(pubkey_path)
	if err != nil {
		log.Println("Error reading public key:", err)
		return nil, nil, err
	}

	block, _ = pem.Decode(publicKeyPEM)
	if block == nil || block.Type != "PUBLIC KEY" {
		log.Println("Invalid public key")
		return nil, nil, err
	}

	// PEM blocks of type "PUBLIC KEY"
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Println("Error parsing public key:", err)
		return nil, nil, err
	}

	return privateKey.(ed25519.PrivateKey), publicKey.(ed25519.PublicKey), nil
}
