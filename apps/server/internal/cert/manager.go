package cert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
	"gorm.io/gorm"
)

type CertificateManager struct {
	db      *gorm.DB
	certDir string
	manager *autocert.Manager
}

type CertificateInfo struct {
	Domain      string    `json:"domain"`
	Issuer      string    `json:"issuer"`
	NotBefore   time.Time `json:"notBefore"`
	NotAfter    time.Time `json:"notAfter"`
	DNSNames    []string  `json:"dnsNames"`
	Fingerprint string    `json:"fingerprint"`
}

func NewCertificateManager(db *gorm.DB, certDir string) *CertificateManager {
	if err := os.MkdirAll(certDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create cert directory: %v", err))
	}
	return &CertificateManager{
		db:      db,
		certDir: certDir,
	}
}

func (m *CertificateManager) InitAutoCert(email string, domains []string) {
	m.manager = &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
		Cache:      autocert.DirCache(filepath.Join(m.certDir, "autocert")),
		Email:      email,
	}
}

func (m *CertificateManager) GetAutoCertManager() *autocert.Manager {
	return m.manager
}

func (m *CertificateManager) GenerateSelfSigned(domain string) (*CertificateInfo, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %v", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"WUI Self-Signed"},
			CommonName:   domain,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privBytes, _ := x509.MarshalECPrivateKey(privKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	certPath := filepath.Join(m.certDir, domain+".crt")
	keyPath := filepath.Join(m.certDir, domain+".key")

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return nil, err
	}

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		os.Remove(certPath)
		return nil, err
	}

	cert, _ := x509.ParseCertificate(derBytes)
	return &CertificateInfo{
		Domain:      domain,
		Issuer:      "WUI Self-Signed",
		NotBefore:   cert.NotBefore,
		NotAfter:    cert.NotAfter,
		DNSNames:    cert.DNSNames,
		Fingerprint: fmt.Sprintf("%x", sha256Sum(derBytes)),
	}, nil
}

func (m *CertificateManager) GetCertificate(domain string) (*CertificateInfo, error) {
	certPath := filepath.Join(m.certDir, domain+".crt")

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("certificate not found: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return &CertificateInfo{
		Domain:      domain,
		Issuer:      cert.Issuer.CommonName,
		NotBefore:   cert.NotBefore,
		NotAfter:    cert.NotAfter,
		DNSNames:    cert.DNSNames,
		Fingerprint: fmt.Sprintf("%x", sha256Sum(block.Bytes)),
	}, nil
}

func (m *CertificateManager) ListCertificates() ([]CertificateInfo, error) {
	var certs []CertificateInfo

	entries, err := os.ReadDir(m.certDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".crt") {
			domain := strings.TrimSuffix(entry.Name(), ".crt")
			info, err := m.GetCertificate(domain)
			if err == nil {
				certs = append(certs, *info)
			}
		}
	}

	return certs, nil
}

func (m *CertificateManager) DeleteCertificate(domain string) error {
	certPath := filepath.Join(m.certDir, domain+".crt")
	keyPath := filepath.Join(m.certDir, domain+".key")

	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (m *CertificateManager) CheckExpiry() (map[string]int, error) {
	certs, err := m.ListCertificates()
	if err != nil {
		return nil, err
	}

	expiry := make(map[string]int)
	for _, cert := range certs {
		days := int(time.Until(cert.NotAfter).Hours() / 24)
		expiry[cert.Domain] = days
	}

	return expiry, nil
}

func (m *CertificateManager) GetCertAndKey(domain string) ([]byte, []byte, error) {
	certPath := filepath.Join(m.certDir, domain+".crt")
	keyPath := filepath.Join(m.certDir, domain+".key")

	cert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, err
	}

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func (m *CertificateManager) UploadCertificate(domain string, certPEM, keyPEM []byte) error {
	certPath := filepath.Join(m.certDir, domain+".crt")
	keyPath := filepath.Join(m.certDir, domain+".key")

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("invalid certificate PEM")
	}

	_, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("invalid certificate: %v", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("invalid key PEM")
	}

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return err
	}

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		os.Remove(certPath)
		return err
	}

	return nil
}

func sha256Sum(data []byte) []byte {
	h := crypto.SHA256.New()
	h.Write(data)
	return h.Sum(nil)
}
