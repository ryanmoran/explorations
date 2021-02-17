package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	// Generate a CA certificate and private key
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization: []string{"Paketo Buildpacks"},
			CommonName:   "Paketo Buildpacks Certificate Authority",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		panic(err)
	}

	err = encodeToFile("ssl/ca.pem", "CERTIFICATE", caBytes)
	if err != nil {
		panic(err)
	}

	// Generate a certificate and private key, signing it with the CA private key
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Paketo Buildpacks"},
			CommonName:   "Paketo Buildpacks Certificate",
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{"localhost"},
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		panic(err)
	}

	err = encodeToFile("ssl/cert.pem", "CERTIFICATE", certBytes)
	if err != nil {
		panic(err)
	}

	err = encodeToFile("ssl/key.pem", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(certPrivKey))
	if err != nil {
		panic(err)
	}
}

func encodeToFile(path, kind string, contents []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	err = pem.Encode(file, &pem.Block{
		Type:  kind,
		Bytes: contents,
	})
	if err != nil {
		return err
	}

	return nil
}
