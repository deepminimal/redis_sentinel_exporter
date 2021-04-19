package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
)

type Options struct {
	Addr                string
	CaCertificates      *x509.CertPool
	ClientCertificates  []tls.Certificate
	ListenAddress       string
	MetricsNamespace    string
	MetricsPath         string
	Password            string
	SkipTLSVerification bool
}

func (o *Options) Validate() error {
	if len(o.Addr) == 0 {
		return errors.New("sentinel address must be specified")
	}
	return nil
}
