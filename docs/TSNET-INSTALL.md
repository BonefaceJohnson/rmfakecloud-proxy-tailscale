# Tailscale (tsnet) Integration — Installation & Nutzung

Dieses Verzeichnis enthält ein kleines Binary (`cmd/tsnet-proxy`) das `tailscale.com/tsnet` verwendet, um ohne Kernel-TUN einen HTTP-Reverse-Proxy über dein Tailnet bereitzustellen.

Vorbereitung
1. Erzeuge einen Auth Key in der Tailscale Admin Console (empfohlen: eingeschränkte TTL).
2. Setze die Umgebungsvariablen:
   - `TS_AUTHKEY` = <dein auth key> (optional, für headless)
   - `TS_HOSTNAME` = gewünschter Hostname im Tailnet (default: rmfakeproxy)
   - `TS_STATE_DIR` = Pfad für persistenten Zustand (z.B. /var/lib/rmfakecloud/ts)
   - `TARGET_URL` = URL, zu der weitergeleitet werden soll (z. B. https://web.remarkable.com)
   - `TS_LISTEN_PORT` = Port, auf dem der Proxy innerhalb des tsnet-Listeners hört (default 8080)

Cross-Compile für reMarkable
- Für ARMv7 (ältere Geräte):
  GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -o rmtsproxy ./cmd/tsnet-proxy
- Für ARM64:
  GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o rmtsproxy ./cmd/tsnet-proxy

Start auf dem Gerät
- Kopiere das Binary aufs Gerät und starte:
  TS_AUTHKEY=tskey-... TS_HOSTNAME=rmfakeproxy TS_STATE_DIR=/home/root/.ts ./rmtsproxy &

Hinweis: `tsnet` funktioniert pro Prozess. Es ersetzt kein systemweites VPN-Interface; für HTTP/TCP-Proxys ist es aber ideal.
