package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/viper"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func flagBool(name, key string) bool {
	if rootCmd.PersistentFlags().Changed(name) {
		v, _ := rootCmd.PersistentFlags().GetBool(name)
		return v
	}
	return viper.GetBool(key)
}

func flagString(name, key string) string {
	if rootCmd.PersistentFlags().Changed(name) {
		v, _ := rootCmd.PersistentFlags().GetString(name)
		return v
	}
	return viper.GetString(key)
}

func tlsConfig() (*tls.Config, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if flagBool("insecure-skip-verify", "insecure_skip_verify") {
		cfg.InsecureSkipVerify = true
	}
	if ca := flagString("ca-file", "ca_file"); ca != "" {
		pem, err := os.ReadFile(ca)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("ca file %s has no certificates", ca)
		}
		cfg.RootCAs = pool
	}
	return cfg, nil
}

func transportCreds() (credentials.TransportCredentials, error) {
	if flagBool("insecure", "insecure") {
		return insecure.NewCredentials(), nil
	}
	cfg, err := tlsConfig()
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(cfg), nil
}

func httpClient() (*http.Client, error) {
	cfg, err := tlsConfig()
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: &http.Transport{TLSClientConfig: cfg}}, nil
}
