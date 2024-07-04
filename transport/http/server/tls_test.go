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

type certDef struct {
	Prefix      string
	IPAddresses []string
	DNSNames    []string
}

func (c certDef) Org() string {
	if c.Prefix == "" {
		return "Acme Co"
	}
	return c.Prefix + " " + "Acme Co"
}

func init() {
	certs := []certDef{
		certDef{
			Prefix:      "",
			IPAddresses: []string{"127.0.0.1", "::1"},
			DNSNames:    []string{"localhost"},
		},
		certDef{
			Prefix:      "example",
			IPAddresses: []string{"127.0.0.1"},
			DNSNames:    []string{"example.com"},
		},
	}

	for _, cd := range certs {
		if err := generateNamedCert(cd); err != nil {
			log.Fatal(err.Error())
		}
	}
}

func generateNamedCert(hostCert certDef) error {
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
			Organization: []string{hostCert.Org()},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, strIP := range hostCert.IPAddresses {
		if ip := net.ParseIP(strIP); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		}
	}
	template.DNSNames = append(template.DNSNames, hostCert.DNSNames...)

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

	serverCert := hostCert.Prefix + "cert.pem"
	serverKey := hostCert.Prefix + "key.pem"
	caCert := hostCert.Prefix + "ca.pem"

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
