package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourname/rank233-server/internal/api"
	"github.com/yourname/rank233-server/internal/version"
	"github.com/yourname/rank233-server/ranker"
)

func main() {
	var (
		addr       string
		token      string
		dataDir    string
		persistSec int
		showVer    bool
	)
	flag.StringVar(&addr, "addr", envOr("RANK233_ADDR", "0.0.0.0:6320"), "listen address")
	flag.StringVar(&token, "token", envOr("RANK233_TOKEN", "neko233"), "auth token")
	flag.StringVar(&dataDir, "data", envOr("RANK233_DATA", "data"), "persistence directory")
	flag.IntVar(&persistSec, "persist-interval", 60, "persistence interval seconds")
	flag.BoolVar(&showVer, "version", false, "print version and exit")
	flag.Parse()

	if showVer {
		fmt.Println(version.Full())
		return
	}

	r := ranker.NewRanker()
	p := ranker.NewPersister(r, dataDir, time.Duration(persistSec)*time.Second)

	loaded, err := p.LoadAll()
	if err != nil {
		log.Printf("WARNING: load persisted data: %v", err)
	} else if loaded > 0 {
		log.Printf("loaded %d ranklists from %s", loaded, dataDir)
	}

	p.Start()
	defer p.Stop()

	srv := api.New(r, p, token)
	handler := srv.Middleware(srv.Router())

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}

	server := &http.Server{Handler: handler}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("shutting down, saving data...")
		p.SaveAll()
		server.Close()
	}()

	log.Printf("rank233-server listening on %s (token=%s)", addr, token)
	if err := server.Serve(ln); err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
