package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tailscale.com/tsnet"
)

func mustEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	targetStr := mustEnv("TARGET_URL", "https://web.remarkable.com")
	target, err := url.Parse(targetStr)
	if err != nil {
		log.Fatalf("invalid TARGET_URL: %v", err)
	}

	hostname := mustEnv("TS_HOSTNAME", "rmfakeproxy")                // <hostname>.beta.tailscale.net
	authKey := os.Getenv("TS_AUTHKEY")                              // optional, für headless Auth
	stateDir := mustEnv("TS_STATE_DIR", "/var/lib/rmfakecloud/ts")  // persistenter Zustand
	listenPort := mustEnv("TS_LISTEN_PORT", "8080")                 // Port innerhalb des tsnet-Listeners

	var s tsnet.Server
	s.Hostname = hostname
	s.AuthKey = authKey
	s.Dir = stateDir
	// s.Logf = log.Printf // bei Bedarf aktivieren

	if err := s.Start(); err != nil {
		log.Fatalf("failed to start tsnet: %v", err)
	}
	defer s.Close()

	ln, err := s.Listen(ctx, "tcp", ":"+listenPort)
	if err != nil {
		log.Fatalf("tsnet.Listen failed: %v", err)
	}
	defer ln.Close()

	proxy := httputil.NewSingleHostReverseProxy(target)
	// Optional: Transport / TLS-Einstellungen anpassen
	server := &http.Server{
		Handler:      proxy,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Serve in goroutine, weil wir Listen/Accept über die tsnet-Listener-Conn nutzen.
	go func() {
		log.Printf("Serving reverse proxy for %s via Tailscale hostname %s on port %s", targetStr, hostname, listenPort)
		if err := server.Serve(&singleListener{ln}); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received, shutting down server...")
	shutdownCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	_ = server.Shutdown(shutdownCtx)
}

// singleListener adapts net.Listener (tsnet's listener) zu einem net.Listener, der Accept() direkt auf die zugrundeliegende Accept() nutzt.
// tsnet.Listen liefert bereits ein net.Listener; diese Adapter-Schicht ist defensiv/kompatibel.
type singleListener struct {
	net.Listener
}
