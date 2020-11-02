// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package main provides a authentication webhook program.
//
// This webhook is meant to be used in demo settings to play around with
// Pinniped. As well, it can come in handy in integration tests.
//
// This webhook is NOT meant for use in production systems.
package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	authenticationv1beta1 "k8s.io/api/authentication/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"go.pinniped.dev/internal/constable"
	"go.pinniped.dev/internal/controller/apicerts"
	"go.pinniped.dev/internal/controllerlib"
	"go.pinniped.dev/internal/dynamiccert"
)

const (
	// This string must match the name of the Namespace declared in the deployment yaml.
	namespace = "local-user-authenticator"
	// This string must match the name of the Service declared in the deployment yaml.
	serviceName = "local-user-authenticator"

	singletonWorker       = 1
	defaultResyncInterval = 3 * time.Minute

	invalidRequest = constable.Error("invalid request")
)

type webhook struct {
	certProvider   dynamiccert.Provider
	secretInformer corev1informers.SecretInformer
}

func newWebhook(
	certProvider dynamiccert.Provider,
	secretInformer corev1informers.SecretInformer,
) *webhook {
	return &webhook{
		certProvider:   certProvider,
		secretInformer: secretInformer,
	}
}

// start runs the webhook in a separate goroutine and returns whether or not the
// webhook was started successfully.
func (w *webhook) start(ctx context.Context, l net.Listener) error {
	server := http.Server{
		Handler: w,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			GetCertificate: func(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
				certPEM, keyPEM := w.certProvider.CurrentCertKeyContent()
				cert, err := tls.X509KeyPair(certPEM, keyPEM)
				return &cert, err
			},
		},
	}

	errCh := make(chan error)
	go func() {
		// Per ListenAndServeTLS doc, the {cert,key}File parameters can be empty
		// since we want to use the certs from http.Server.TLSConfig.
		errCh <- server.ServeTLS(l, "", "")
	}()

	go func() {
		select {
		case err := <-errCh:
			klog.InfoS("server exited", "err", err)
		case <-ctx.Done():
			klog.InfoS("server context cancelled", "err", ctx.Err())
			if err := server.Shutdown(context.Background()); err != nil {
				klog.InfoS("server shutdown failed", "err", err)
			}
		}
	}()

	return nil
}

func (w *webhook) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {
	username, password, err := getUsernameAndPasswordFromRequest(rsp, req)
	if err != nil {
		return
	}
	defer func() { _ = req.Body.Close() }()

	secret, err := w.secretInformer.Lister().Secrets(namespace).Get(username)
	notFound := k8serrors.IsNotFound(err)
	if err != nil && !notFound {
		klog.InfoS("could not get secret", "err", err)
		rsp.WriteHeader(http.StatusInternalServerError)
		return
	}

	if notFound {
		klog.InfoS("user not found")
		respondWithUnauthenticated(rsp)
		return
	}

	passwordMatches := bcrypt.CompareHashAndPassword(
		secret.Data["passwordHash"],
		[]byte(password),
	) == nil
	if !passwordMatches {
		klog.InfoS("authentication failed: wrong password")
		respondWithUnauthenticated(rsp)
		return
	}

	groups := []string{}
	groupsBuf := bytes.NewBuffer(secret.Data["groups"])
	if groupsBuf.Len() > 0 {
		groupsCSVReader := csv.NewReader(groupsBuf)
		groups, err = groupsCSVReader.Read()
		if err != nil {
			klog.InfoS("could not read groups", "err", err)
			rsp.WriteHeader(http.StatusInternalServerError)
			return
		}
		trimLeadingAndTrailingWhitespace(groups)
	}

	klog.InfoS("successful authentication")
	respondWithAuthenticated(rsp, secret.ObjectMeta.Name, string(secret.UID), groups)
}

func getUsernameAndPasswordFromRequest(rsp http.ResponseWriter, req *http.Request) (string, string, error) {
	if req.URL.Path != "/authenticate" {
		klog.InfoS("received request path other than /authenticate", "path", req.URL.Path)
		rsp.WriteHeader(http.StatusNotFound)
		return "", "", invalidRequest
	}

	if req.Method != http.MethodPost {
		klog.InfoS("received request method other than post", "method", req.Method)
		rsp.WriteHeader(http.StatusMethodNotAllowed)
		return "", "", invalidRequest
	}

	if !headerContains(req, "Content-Type", "application/json") {
		klog.InfoS("content type is not application/json", "Content-Type", req.Header.Values("Content-Type"))
		rsp.WriteHeader(http.StatusUnsupportedMediaType)
		return "", "", invalidRequest
	}

	if !headerContains(req, "Accept", "application/json") &&
		!headerContains(req, "Accept", "application/*") &&
		!headerContains(req, "Accept", "*/*") {
		klog.InfoS("client does not accept application/json", "Accept", req.Header.Values("Accept"))
		rsp.WriteHeader(http.StatusUnsupportedMediaType)
		return "", "", invalidRequest
	}

	if req.Body == nil {
		klog.InfoS("invalid nil body")
		rsp.WriteHeader(http.StatusBadRequest)
		return "", "", invalidRequest
	}

	var body authenticationv1beta1.TokenReview
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		klog.InfoS("failed to decode body", "err", err)
		rsp.WriteHeader(http.StatusBadRequest)
		return "", "", invalidRequest
	}

	if body.APIVersion != authenticationv1beta1.SchemeGroupVersion.String() {
		klog.InfoS("invalid TokenReview apiVersion", "apiVersion", body.APIVersion)
		rsp.WriteHeader(http.StatusBadRequest)
		return "", "", invalidRequest
	}

	if body.Kind != "TokenReview" {
		klog.InfoS("invalid TokenReview kind", "kind", body.Kind)
		rsp.WriteHeader(http.StatusBadRequest)
		return "", "", invalidRequest
	}

	tokenSegments := strings.SplitN(body.Spec.Token, ":", 2)
	if len(tokenSegments) != 2 {
		klog.InfoS("bad token format in request")
		rsp.WriteHeader(http.StatusBadRequest)
		return "", "", invalidRequest
	}

	return tokenSegments[0], tokenSegments[1], nil
}

func headerContains(req *http.Request, headerName, s string) bool {
	headerValues := req.Header.Values(headerName)
	for i := range headerValues {
		mimeTypes := strings.Split(headerValues[i], ",")
		for _, mimeType := range mimeTypes {
			mediaType, _, _ := mime.ParseMediaType(mimeType)
			if mediaType == s {
				return true
			}
		}
	}
	return false
}

func trimLeadingAndTrailingWhitespace(ss []string) {
	for i := range ss {
		ss[i] = strings.TrimSpace(ss[i])
	}
}

func respondWithUnauthenticated(rsp http.ResponseWriter) {
	rsp.Header().Add("Content-Type", "application/json")

	body := authenticationv1beta1.TokenReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TokenReview",
			APIVersion: authenticationv1beta1.SchemeGroupVersion.String(),
		},
		Status: authenticationv1beta1.TokenReviewStatus{
			Authenticated: false,
		},
	}
	if err := json.NewEncoder(rsp).Encode(body); err != nil {
		klog.InfoS("could not encode response", "err", err)
		rsp.WriteHeader(http.StatusInternalServerError)
	}
}

func respondWithAuthenticated(
	rsp http.ResponseWriter,
	username, uid string,
	groups []string,
) {
	rsp.Header().Add("Content-Type", "application/json")
	body := authenticationv1beta1.TokenReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TokenReview",
			APIVersion: authenticationv1beta1.SchemeGroupVersion.String(),
		},
		Status: authenticationv1beta1.TokenReviewStatus{
			Authenticated: true,
			User: authenticationv1beta1.UserInfo{
				Username: username,
				Groups:   groups,
				UID:      uid,
			},
		},
	}
	if err := json.NewEncoder(rsp).Encode(body); err != nil {
		klog.InfoS("could not encode response", "err", err)
		rsp.WriteHeader(http.StatusInternalServerError)
	}
}

func newK8sClient() (kubernetes.Interface, error) {
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("could not load in-cluster configuration: %w", err)
	}

	// Connect to the core Kubernetes API.
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("could not load in-cluster configuration: %w", err)
	}

	return kubeClient, nil
}

func startControllers(
	ctx context.Context,
	dynamicCertProvider dynamiccert.Provider,
	kubeClient kubernetes.Interface,
	kubeInformers kubeinformers.SharedInformerFactory,
) {
	aVeryLongTime := time.Hour * 24 * 365 * 100

	const certsSecretResourceName = "local-user-authenticator-tls-serving-certificate"

	// Create controller manager.
	controllerManager := controllerlib.
		NewManager().
		WithController(
			apicerts.NewCertsManagerController(
				namespace,
				certsSecretResourceName,
				map[string]string{
					"app": "local-user-authenticator",
				},
				kubeClient,
				kubeInformers.Core().V1().Secrets(),
				controllerlib.WithInformer,
				controllerlib.WithInitialEvent,
				aVeryLongTime,
				"local-user-authenticator CA",
				serviceName,
			),
			singletonWorker,
		).
		WithController(
			apicerts.NewCertsObserverController(
				namespace,
				certsSecretResourceName,
				dynamicCertProvider,
				kubeInformers.Core().V1().Secrets(),
				controllerlib.WithInformer,
			),
			singletonWorker,
		)

	kubeInformers.Start(ctx.Done())

	go controllerManager.Start(ctx)
}

func startWebhook(
	ctx context.Context,
	l net.Listener,
	dynamicCertProvider dynamiccert.Provider,
	secretInformer corev1informers.SecretInformer,
) error {
	return newWebhook(dynamicCertProvider, secretInformer).start(ctx, l)
}

func waitForSignal() os.Signal {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	return <-signalCh
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kubeClient, err := newK8sClient()
	if err != nil {
		return fmt.Errorf("cannot create k8s client: %w", err)
	}

	kubeInformers := kubeinformers.NewSharedInformerFactoryWithOptions(
		kubeClient,
		defaultResyncInterval,
		kubeinformers.WithNamespace(namespace),
	)

	dynamicCertProvider := dynamiccert.New()

	startControllers(ctx, dynamicCertProvider, kubeClient, kubeInformers)
	klog.InfoS("controllers are ready")

	//nolint: gosec // Intentionally binding to all network interfaces.
	l, err := net.Listen("tcp", ":8443")
	if err != nil {
		return fmt.Errorf("cannot create listener: %w", err)
	}
	defer func() { _ = l.Close() }()

	err = startWebhook(ctx, l, dynamicCertProvider, kubeInformers.Core().V1().Secrets())
	if err != nil {
		return fmt.Errorf("cannot start webhook: %w", err)
	}
	klog.InfoS("webhook is ready", "address", l.Addr().String())

	gotSignal := waitForSignal()
	klog.InfoS("webhook exiting", "signal", gotSignal)

	return nil
}

func main() {
	if err := run(); err != nil {
		klog.Fatal(err)
	}
}
