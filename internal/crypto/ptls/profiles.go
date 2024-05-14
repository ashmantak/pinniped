// Copyright 2021-2024 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build !fips_strict

package ptls

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"runtime"

	"k8s.io/apiserver/pkg/server/options"

	"go.pinniped.dev/internal/plog"
)

// init prints a log message to tell the operator how Pinniped was compiled. This makes it obvious
// that they are using Pinniped in FIPS-mode or not, which is otherwise hard to observe.
func init() { //nolint:gochecknoinits
	switch filepath.Base(os.Args[0]) {
	case "pinniped-server", "pinniped-supervisor", "pinniped-concierge", "pinniped-concierge-kube-cert-agent":
	default:
		return // do not print FIPS logs if we cannot confirm that we are running a server binary
	}

	// this init runs before we have parsed our config to determine our log level
	// thus we must use a log statement that will always print instead of conditionally print
	plog.Always("this server was not compiled in FIPS-only mode",
		"go version", runtime.Version())
}

// SecureTLSConfigMinTLSVersion is the minimum tls version in the format expected by tls.Config.
const SecureTLSConfigMinTLSVersion = tls.VersionTLS13

// Default TLS profile should be used by:
// A. servers whose clients are outside our control and who may reasonably wish to use TLS 1.2, and
// B. clients who need to interact with servers that might not support TLS 1.3.
// Note that this will behave differently when compiled in FIPS mode (see profiles_fips_strict.go).
// Default returns a tls.Config with a minimum of TLS1.2+ and a few ciphers that can be further constrained by configuration.
func Default(rootCAs *x509.CertPool) *tls.Config {
	return buildTLSConfig(rootCAs, cipherSuitesForDefault(), getUserConfiguredCiphersAllowList())
}

// DefaultLDAP TLS profile should be used by clients who need to interact with potentially old LDAP servers
// that might not support TLS 1.3 and that might use older ciphers.
// Note that this will behave differently when compiled in FIPS mode (see profiles_fips_strict.go).
func DefaultLDAP(rootCAs *x509.CertPool) *tls.Config {
	return buildTLSConfig(rootCAs, cipherSuitesForDefaultLDAP(), getUserConfiguredCiphersAllowList())
}

// Secure TLS profile should be used by:
// A. servers whose clients are entirely known by us and who may reasonably be told that they must use TLS 1.3, and
// B. clients who only need to interact with servers that are known by us to support TLS 1.3 (e.g. the Kubernetes API).
// Note that this will behave differently when compiled in FIPS mode (see profiles_fips_strict.go).
func Secure(rootCAs *x509.CertPool) *tls.Config {
	// as of 2021-10-19, Mozilla Guideline v5.6, Go 1.17.2, modern configuration, supports:
	// - Firefox 63
	// - Android 10.0
	// - Chrome 70
	// - Edge 75
	// - Java 11
	// - OpenSSL 1.1.1
	// - Opera 57
	// - Safari 12.1
	// https://ssl-config.mozilla.org/#server=go&version=1.17.2&config=modern&guideline=5.6
	c := Default(rootCAs)
	c.MinVersion = SecureTLSConfigMinTLSVersion // max out the security
	c.CipherSuites = nil                        // TLS 1.3 ciphers are not configurable
	return c
}

// SecureServing modifies the given options to have the appropriate MinTLSVersion and CipherSuites.
// This function should only be used by the implementation of ptls.SecureRecommendedOptions, which
// is called to help configure our aggregated API servers. This exists only because it needs
// to behave differently in FIPS mode.
// This function is only public so we can integration test it in ptls_fips_test.go.
// Note that this will behave differently when compiled in FIPS mode (see profiles_fips_strict.go).
func SecureServing(opts *options.SecureServingOptionsWithLoopback) {
	// secureServingOptionsMinTLSVersion is the minimum tls version in the format
	// expected by SecureServingOptions.MinTLSVersion from
	// k8s.io/apiserver/pkg/server/options.
	opts.MinTLSVersion = "VersionTLS13"
	opts.CipherSuites = nil
}

func hardcodedCipherSuites() []*tls.CipherSuite {
	return cipherSuitesForDefaultLDAP()
}

// cipherSuitesForDefault are the ciphers that Pinniped allows.
// It will be a strict subset of tls.CipherSuites.
func cipherSuitesForDefault() []*tls.CipherSuite {
	// the order does not matter in go 1.17+ https://go.dev/blog/tls-cipher-suites
	// we match crypto/tls.cipherSuitesPreferenceOrder because it makes unit tests easier to write
	// this list is ignored when TLS 1.3 is used
	//
	// as of 2021-10-19, Mozilla Guideline v5.6, Go 1.17.2, intermediate configuration, supports:
	// - Firefox 27
	// - Android 4.4.2
	// - Chrome 31
	// - Edge
	// - IE 11 on Windows 7
	// - Java 8u31
	// - OpenSSL 1.0.1
	// - Opera 20
	// - Safari 9
	// https://ssl-config.mozilla.org/#server=go&version=1.17.2&config=intermediate&guideline=5.6
	//
	// The Kubernetes API server must use approved cipher suites.
	// https://stigviewer.com/stig/kubernetes/2021-06-17/finding/V-242418

	// These are all AEADs with ECDHE, some use ChaCha20Poly1305 while others use AES-GCM,
	// which provides forward secrecy, confidentiality and authenticity of data.
	cipherSuiteIDsForDefault := []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	}

	return translateIDIntoSecureCipherSuites(cipherSuiteIDsForDefault)
}

// cipherSuitesForDefaultLDAP are some additional ciphers that Pinniped allows only for LDAP.
// It will be a strict subset of tls.CipherSuites.
func cipherSuitesForDefaultLDAP() []*tls.CipherSuite {
	// Add less secure ciphers to support the default AWS Active Directory config
	//
	// CBC with ECDHE
	// this provides forward secrecy and confidentiality of data but not authenticity
	// MAC-then-Encrypt CBC ciphers are susceptible to padding oracle attacks
	// See https://crypto.stackexchange.com/a/205 and https://crypto.stackexchange.com/a/224
	cipherSuiteIDsForDefaultLDAP := []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	}
	result := cipherSuitesForDefault()
	result = append(result, translateIDIntoSecureCipherSuites(cipherSuiteIDsForDefaultLDAP)...)
	return result
}
