package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
)

var (
	isDebug   = flag.Bool("debug", GetBoolEnv("DEBUG", false), "Output verbose debug information.")
	logFormat = flag.String("log-format", GetStringEnv("LOG_FORMAT", "txt"), "Log format, valid options are txt and json.")

	listenAddress               = flag.String("web.listen-address", GetStringEnv("LISTEN_ADDRESS", ":9355"), "Address to listen on for web interface and telemetry.")
	metricsPath                 = flag.String("web.telemetry-path", GetStringEnv("TELEMETRY_PATH", "/metrics"), "Path under which to expose metrics.")
	namespace                   = flag.String("namespace", GetStringEnv("NAMESPACE", "redis_sentinel"), "Namespace for metrics.")
	sentinelAddr                = flag.String("sentinel.addr", GetStringEnv("SENTINEL_ADDR", "redis://127.0.0.1:26379"), "Redis Sentinel host:port.")
	sentinelConnectionTimeout   = flag.Duration("sentinel.connection-timeout", GetDurationEnv("SENTINEL_CONNECTION_TIMEOUT", 5*time.Second), "Timeout for connection to Redis Sentinel instance.")
	sentinelPassword            = flag.String("sentinel.password", GetStringEnv("SENTINEL_PASSWORD", ""), "Redis Sentinel password (optional).")
	sentinelPasswordFile        = flag.String("sentinel.password-file", GetStringEnv("SENTINEL_PASSWORD_FILE", ""), "Path to Redis Sentinel password file (optional).")
	sentinelSkipTLSVerification = flag.Bool("sentinel.skip-tls-verification", GetBoolEnv("SENTINEL_SKIP_TLS_VERIFICATION", false), "Skip TLS verification.")
	sentinelTLSCaCertFile       = flag.String("sentinel.tls-ca-cert-file", GetStringEnv("SENTINEL_TLS_CA_CERT_FILE", ""), "Name of the CA certificate file, including full path (optional).")
	sentinelTLSClientCertFile   = flag.String("sentinel.tls-client-cert-file", GetStringEnv("SENTINEL_TLS_CLIENT_CERT_FILE", ""), "Name of the client certificate file, including full path (optional).")
	sentinelTLSClientKeyFile    = flag.String("sentinel.tls-client-key-file", GetStringEnv("SENTINEL_TLS_CLIENT_KEY_FILE", ""), "Name of the client key file, including full path (optional).")
	versionPrint                = flag.Bool("version", false, "Prints version and exit.")
)

func main() {
	flag.Parse()

	if *versionPrint {
		fmt.Println(version.Print("redis_sentinel_exporter"))
		os.Exit(0)
	}

	switch *logFormat {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	logrus.Infof("Starting Redis Sentinel Exporter %s...", version.Version)

	if *isDebug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Enabling debug output")
	}

	var tlsClientCertificates []tls.Certificate
	if (*sentinelTLSClientKeyFile != "") != (*sentinelTLSClientCertFile != "") {
		logrus.Fatal("TLS client key file and cert file should both be present")
	}
	if *sentinelTLSClientKeyFile != "" && *sentinelTLSClientCertFile != "" {
		cert, err := tls.LoadX509KeyPair(*sentinelTLSClientCertFile, *sentinelTLSClientKeyFile)
		if err != nil {
			logrus.Fatalf("Couldn't load TLS client key pair, err: %s", err)
		}
		tlsClientCertificates = append(tlsClientCertificates, cert)
	}

	var tlsCaCertificates *x509.CertPool
	if *sentinelTLSCaCertFile != "" {
		caCert, err := ioutil.ReadFile(*sentinelTLSCaCertFile)
		if err != nil {
			logrus.Fatalf("Couldn't load TLS Ca certificate, err: %s", err)
		}
		tlsCaCertificates = x509.NewCertPool()
		tlsCaCertificates.AppendCertsFromPEM(caCert)
	}

	password := *sentinelPassword
	if len(*sentinelPasswordFile) > 0 {
		body, err := ioutil.ReadFile(*sentinelPasswordFile)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to load Redis Sentinel password file")
		}
		password = strings.TrimSpace(string(body))
	}

	options := &Options{
		Addr:                *sentinelAddr,
		CaCertificates:      tlsCaCertificates,
		ClientCertificates:  tlsClientCertificates,
		ConnectionTimeout:   *sentinelConnectionTimeout,
		ListenAddress:       *listenAddress,
		MetricsNamespace:    *namespace,
		MetricsPath:         *metricsPath,
		Password:            password,
		SkipTLSVerification: *sentinelSkipTLSVerification,
	}
	if err := options.Validate(); err != nil {
		logrus.WithError(err).Fatal("Validation failed")
	}

	exporter := NewRedisSentinelExporter(options)

	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector(options.MetricsNamespace + "_exporter"))
	// Deprecated and will be removed in 2.0, use `redis_sentinel_exporter_build_info`
	prometheus.MustRegister(version.NewCollector(options.MetricsNamespace))

	http.Handle(options.MetricsPath, promhttp.Handler())
	http.HandleFunc("/healthy", exporter.HealthyHandler)
	http.HandleFunc("/", exporter.IndexHandler)

	logrus.Printf("Providing metrics at %s%s", options.ListenAddress, options.MetricsPath)
	logrus.Fatal(http.ListenAndServe(options.ListenAddress, nil))
}
