// SPDX-License-Identifier: Apache-2.0

package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func init() {
	if err := generateCerts(); err != nil {
		log.Fatal(err.Error())
	}
}

func generateCerts() error {
	hosts := []string{"127.0.0.1", "::1", "localhost"}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("Failed to generate private key: %v", err)
	}

	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment

	notBefore := time.Now()
	notAfter := notBefore.Add(1000000 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("Failed to create certificate: %v", err)
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("Failed to create ca: %v", err)
	}

	serverCert := "cert.pem"
	serverKey := "key.pem"
	caCert := "ca.pem"

	certOut, err := os.Create(serverCert)
	if err != nil {
		return fmt.Errorf("Failed to open %s for writing: %v", serverCert, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("Failed to write data to %s: %v", serverCert, err)
	}
	if err := certOut.Close(); err != nil {
		return fmt.Errorf("Error closing %s: %v", serverCert, err)
	}
	log.Printf("wrote %s\n", serverCert)

	keyOut, err := os.OpenFile(serverKey, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Failed to open %s for writing: %v", serverKey, err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("Failed to write data to %s: %v", serverKey, err)
	}
	if err := keyOut.Close(); err != nil {
		return fmt.Errorf("Error closing %s: %v", serverKey, err)
	}
	log.Printf("wrote %s\n", serverKey)

	caOut, err := os.Create(caCert)
	if err != nil {
		return fmt.Errorf("Failed to open %s for writing: %v", caCert, err)
	}
	if err := pem.Encode(caOut, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		return fmt.Errorf("Failed to write data to %s: %v", caCert, err)
	}
	if err := caOut.Close(); err != nil {
		return fmt.Errorf("Error closing %s: %v", caCert, err)
	}
	log.Printf("wrote %s\n", caCert)
	return nil
}
