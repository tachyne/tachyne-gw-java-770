// Command gw runs the tachyne Java gateway for protocols 770-772
// (Minecraft 1.21.5-1.21.8). The entire gateway — front door and session
// pipeline — lives in tachyne-common/gwsession; this binary is only the
// version pinning + environment wiring.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tachyne/tachyne-common/access"
	"github.com/tachyne/tachyne-common/gwsession"
)

// Protocol pinning for this gateway build. The world is rendered as canonical
// 770; a per-connection translation chain then serves the accepted range.
// 770-772 are verified pure ID remaps (1.21.5-1.21.8). Raise MaxProto to widen
// coverage once the added step is client-verified (773+ also needs matching
// config-phase data / dispatch routing).
const (
	Protocol    = 770
	MinProto    = 770
	MaxProto    = 772
	VersionName = "1.21.5"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	s := &gwsession.Server{
		Listen:       envOr("TACHYNE_LISTEN", ":25565"),
		Backend:      envOr("TACHYNE_BACKEND", "127.0.0.1:25500"),
		WorldPattern: envOr("TACHYNE_WORLD_PATTERN", "tachyne-world-%d.tachyne-world.tachyne.svc:25500"),
		AttachToken:  os.Getenv("TACHYNE_ATTACH_TOKEN"),
		MOTD:         envOr("TACHYNE_MOTD", "tachyne — Minecraft "+VersionName+" gateway"),
		SID:          ordinal(os.Getenv("POD_NAME")),
		Name:         "gw-java-770",
		VersionName:  VersionName,
		Proto:        Protocol,
		MinProto:     MinProto,
		MaxProto:     MaxProto,
	}
	if url := os.Getenv("TACHYNE_ACCESS_URL"); url != "" {
		s.Access = access.New(url, os.Getenv("TACHYNE_ACCESS_TOKEN"), 30*time.Second)
		log.Printf("access control via %s (fail closed)", url)
	} else {
		log.Print("WARNING: TACHYNE_ACCESS_URL unset — running OPEN (no access control)")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Printf("tachyne-gw-java sid=%d proto=%d (%s) listening on %s, world %s",
		s.SID, Protocol, VersionName, s.Listen, s.Backend)
	if err := s.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func ordinal(pod string) int {
	i := strings.LastIndexByte(pod, '-')
	if i < 0 {
		return 0
	}
	n, err := strconv.Atoi(pod[i+1:])
	if err != nil || n < 0 {
		return 0
	}
	return n
}
