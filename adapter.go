package main

import (
	"crypto/subtle"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

type tokenDocument struct {
	Token string `json:"token"`
}

func main() {
	listen := flag.String("listen", "0.0.0.0:8181", "public listen address")
	upstreamText := flag.String("upstream", "http://127.0.0.1:8182", "private InfluxDB URL")
	tokenFile := flag.String("token-file", "/data/admin-token.json", "offline admin token file")
	flag.Parse()

	raw, err := os.ReadFile(*tokenFile)
	if err != nil {
		log.Fatalf("read protected token file: %v", err)
	}
	var doc tokenDocument
	if err := json.Unmarshal(raw, &doc); err != nil || doc.Token == "" {
		log.Fatal("protected token file is invalid")
	}
	internalBearer := "Bearer " + doc.Token
	externalToken := os.Getenv("INFLUXDB3_EXTERNAL_BEARER_TOKEN")
	if !utf8.ValidString(externalToken) || utf8.RuneCountInString(externalToken) < 32 {
		log.Fatal("external bearer token must contain at least 32 characters")
	}
	if subtle.ConstantTimeCompare([]byte(externalToken), []byte(doc.Token)) == 1 {
		log.Fatal("external bearer token must differ from the protected internal token")
	}
	externalBearer := "Bearer " + externalToken
	upstream, err := url.Parse(*upstreamText)
	if err != nil {
		log.Fatalf("parse upstream: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, proxyErr error) {
		log.Printf("upstream request failed: %v", proxyErr)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			r.URL.Path = "/health"
			r.Header.Set("Authorization", internalBearer)
			proxy.ServeHTTP(w, r)
			return
		}
		provided := r.Header.Get("Authorization")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(externalBearer)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		r.Header.Set("Authorization", internalBearer)
		proxy.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:              *listen,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	log.Printf("authenticated adapter listening on %s for %s", *listen, strings.TrimSuffix(upstream.String(), "/"))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Errorf("adapter server: %w", err))
	}
}
