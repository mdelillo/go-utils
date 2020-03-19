package certutils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

type certConfig struct {
	keyUsage       x509.KeyUsage
	extKeyUsage    []x509.ExtKeyUsage
	isCA           bool
	maxPathLenZero bool
	useCA          bool
	caCertPEMBlock []byte
	caKeyPEMBlock  []byte
	maxPathLen     int
	commonName     string
	dnsNames       []string
	notBefore      time.Time
	notAfter       time.Time
}

type CertOption func(*certConfig)

const oneYear = time.Hour * 24 * 365

func WithIsCA(isCA bool) CertOption {
	return func(c *certConfig) {
		c.isCA = isCA
		c.keyUsage |= x509.KeyUsageCertSign
	}
}

func WithMaxPathLen(maxPathLen int) CertOption {
	return func(c *certConfig) {
		c.maxPathLen = maxPathLen
		c.maxPathLenZero = maxPathLen == 0
	}
}

func WithCA(caCertPEMBlock, caKeyPEMBlock []byte) CertOption {
	return func(c *certConfig) {
		c.useCA = true
		c.caCertPEMBlock = caCertPEMBlock
		c.caKeyPEMBlock = caKeyPEMBlock
	}
}

func WithCommonName(commonName string) CertOption {
	return func(c *certConfig) {
		c.commonName = commonName
	}
}

func WithDNSNames(dnsNames []string) CertOption {
	return func(c *certConfig) {
		c.dnsNames = dnsNames
	}
}

func WithNotBefore(notBefore time.Time) CertOption {
	return func(c *certConfig) {
		c.notBefore = notBefore
	}
}

func WithNotAfter(notAfter time.Time) CertOption {
	return func(c *certConfig) {
		c.notAfter = notAfter
	}
}

func GenerateCert(options ...CertOption) ([]byte, []byte, error) {
	config := certConfig{
		keyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		extKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		commonName:  "my-cert",
		notBefore:   time.Now(),
		notAfter:    time.Now().Add(oneYear),
	}

	for _, option := range options {
		option(&config)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		KeyUsage:              config.keyUsage,
		ExtKeyUsage:           config.extKeyUsage,
		Subject:               pkix.Name{CommonName: config.commonName},
		DNSNames:              config.dnsNames,
		NotBefore:             config.notBefore,
		NotAfter:              config.notAfter,
		IsCA:                  config.isCA,
		BasicConstraintsValid: config.isCA,
		MaxPathLen:            config.maxPathLen,
		MaxPathLenZero:        config.maxPathLenZero,
	}

	return generateCertAndKey(config, template)
}

func generateCertAndKey(config certConfig, template *x509.Certificate) ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key: %w", err)
	}

	var derBytes []byte

	if config.useCA {
		caTLS, err := tls.X509KeyPair(config.caCertPEMBlock, config.caKeyPEMBlock)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load CA key pair: %s", err)
		}

		ca, err := x509.ParseCertificate(caTLS.Certificate[0])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse CA certificate: %s", err)
		}

		csrDerBytes, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{}, key)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate csr: %s", err.Error())
		}

		csr, err := x509.ParseCertificateRequest(csrDerBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse csr: %s", err.Error())
		}

		derBytes, err = x509.CreateCertificate(rand.Reader, template, ca, csr.PublicKey, caTLS.PrivateKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create certificate: %s", err)
		}
	} else {
		derBytes, err = x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
		}
	}

	var certPEMBlock, keyPEMBlock bytes.Buffer

	err = pem.Encode(&certPEMBlock, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode certificate: %w", err)
	}

	err = pem.Encode(&keyPEMBlock, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode key: %w", err)
	}

	return certPEMBlock.Bytes(), keyPEMBlock.Bytes(), nil
}
