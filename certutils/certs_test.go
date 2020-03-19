package certutils_test

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"

	"github.com/mdelillo/go-utils/certutils"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCerts(t *testing.T) {
	spec.Run(t, "Certs", testCerts, spec.Report(report.Terminal{}))
}

func testCerts(t *testing.T, context spec.G, it spec.S) {
	parseCert := func(certPEMBlock, keyPEMBlock []byte) *x509.Certificate {
		tlsCert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
		require.NoError(t, err)

		x509Cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
		require.NoError(t, err)

		return x509Cert
	}

	context("GenerateCert", func() {
		it("generates a self-signed cert and key", func() {
			certPEMBlock, keyPEMBlock, err := certutils.GenerateCert()
			require.NoError(t, err)

			cert := parseCert(certPEMBlock, keyPEMBlock)
			assert.Equal(t, x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature, cert.KeyUsage)
			assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}, cert.ExtKeyUsage)
			assert.Equal(t, "my-cert", cert.Subject.CommonName)
			assert.Equal(t, "my-cert", cert.Issuer.CommonName)
			assert.Empty(t, cert.DNSNames)
			assert.WithinDuration(t, time.Now(), cert.NotBefore, time.Minute)
			assert.WithinDuration(t, time.Now().Add(time.Hour*24*365), cert.NotAfter, time.Minute)
			assert.False(t, cert.IsCA)
			assert.False(t, cert.BasicConstraintsValid)
		})

		context("WithIsCA", func() {
			it("generates a cert that can sign other certs", func() {
				caCertPEMBlock, caKeyPEMBlock, err := certutils.GenerateCert(
					certutils.WithIsCA(true),
					certutils.WithCommonName("ca"),
				)
				require.NoError(t, err)

				ca := parseCert(caCertPEMBlock, caKeyPEMBlock)
				assert.Equal(t, x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature|x509.KeyUsageCertSign, ca.KeyUsage)
				assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}, ca.ExtKeyUsage)
				assert.Equal(t, "ca", ca.Subject.CommonName)
				assert.Equal(t, "ca", ca.Issuer.CommonName)
				assert.Empty(t, ca.DNSNames)
				assert.WithinDuration(t, time.Now(), ca.NotBefore, time.Minute)
				assert.WithinDuration(t, time.Now().Add(time.Hour*24*365), ca.NotAfter, time.Minute)
				assert.True(t, ca.IsCA)
				assert.True(t, ca.BasicConstraintsValid)

				intermediateCertPEMBlock, intermediateKeyPEMBlock, err := certutils.GenerateCert(
					certutils.WithIsCA(true),
					certutils.WithCA(caCertPEMBlock, caKeyPEMBlock),
					certutils.WithCommonName("intermediate"),
				)
				require.NoError(t, err)

				intermediate := parseCert(intermediateCertPEMBlock, intermediateKeyPEMBlock)
				assert.Equal(t, "intermediate", intermediate.Subject.CommonName)
				assert.Equal(t, "ca", intermediate.Issuer.CommonName)

				leafCertPEMBlock, leafKeyPEMBlock, err := certutils.GenerateCert(
					certutils.WithIsCA(true),
					certutils.WithCA(intermediateCertPEMBlock, intermediateKeyPEMBlock),
					certutils.WithCommonName("leaf"),
				)
				require.NoError(t, err)

				leaf := parseCert(leafCertPEMBlock, leafKeyPEMBlock)
				assert.Equal(t, "leaf", leaf.Subject.CommonName)
				assert.Equal(t, "intermediate", leaf.Issuer.CommonName)
			})
		})

		context("WithCA", func() {
			it("generates a cert signed by the CA", func() {
				caCommonName := "some-ca"
				caCertPEMBlock, caKeyPEMBlock, err := certutils.GenerateCert(
					certutils.WithIsCA(true),
					certutils.WithCommonName(caCommonName),
				)
				require.NoError(t, err)

				certPEMBlock, keyPEMBlock, err := certutils.GenerateCert(
					certutils.WithCA(caCertPEMBlock, caKeyPEMBlock),
				)
				require.NoError(t, err)

				cert := parseCert(certPEMBlock, keyPEMBlock)
				assert.Equal(t, caCommonName, cert.Issuer.CommonName)
			})

			context("WithMaxPathLen", func() {
				it("sets the MaxPathLen", func() {
					certPEMBlock, keyPEMBlock, err := certutils.GenerateCert(
						certutils.WithIsCA(true),
						certutils.WithMaxPathLen(99),
					)
					require.NoError(t, err)

					cert := parseCert(certPEMBlock, keyPEMBlock)
					assert.Equal(t, 99, cert.MaxPathLen)
					assert.False(t, cert.MaxPathLenZero)
				})
			})

			context("WithMaxPathLen of 0", func() {
				it("sets MaxPathLenZero to true", func() {
					certPEMBlock, keyPEMBlock, err := certutils.GenerateCert(
						certutils.WithIsCA(true),
						certutils.WithMaxPathLen(0),
					)
					require.NoError(t, err)

					cert := parseCert(certPEMBlock, keyPEMBlock)
					assert.Equal(t, 0, cert.MaxPathLen)
					assert.True(t, cert.MaxPathLenZero)
				})
			})
		})

		context("WithOptions", func() {
			it("generates a cert with the specified options", func() {
				commonName := "some-common-name"
				notBefore := time.Now().Add(-1 * time.Hour)
				notAfter := time.Now().Add(time.Hour)
				dnsNames := []string{"some-dns-name"}

				certPEMBlock, keyPEMBlock, err := certutils.GenerateCert(
					certutils.WithCommonName(commonName),
					certutils.WithDNSNames(dnsNames),
					certutils.WithNotBefore(notBefore),
					certutils.WithNotAfter(notAfter),
				)
				require.NoError(t, err)

				cert := parseCert(certPEMBlock, keyPEMBlock)
				assert.Equal(t, commonName, cert.Subject.CommonName)
				assert.Equal(t, dnsNames, cert.DNSNames)
				assert.WithinDuration(t, notBefore, cert.NotBefore, time.Second)
				assert.WithinDuration(t, notAfter, cert.NotAfter, time.Second)
			})
		})
	})
}
