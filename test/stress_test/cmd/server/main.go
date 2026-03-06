package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/ATSOTECK/rage/test/stress_test/internal/server"
)

func main() {
	// Use GOMEMLIMIT with higher GOGC for smooth GC behavior under load.
	// The memory limit lets GC run more frequently only when approaching the cap,
	// while GOGC=200 provides a reasonable baseline frequency.
	debug.SetGCPercent(200)
	debug.SetMemoryLimit(256 * 1024 * 1024) // 256MB soft limit

	addr := flag.String("addr", ":8080", "listen address")
	poolSize := flag.Int("pool-size", 50, "number of RAGE states per company")
	scriptsDir := flag.String("scripts-dir", "../../scripts", "path to Python scripts directory")
	flag.Parse()

	// Start pprof server on a separate port for profiling
	go func() {
		fmt.Println("pprof available at http://localhost:6060/debug/pprof/")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			fmt.Fprintf(os.Stderr, "pprof server error: %v\n", err)
		}
	}()

	fmt.Printf("Starting server on %s (pool-size=%d, scripts=%s)\n", *addr, *poolSize, *scriptsDir)

	srv, err := server.New(server.Config{
		PoolSize:   *poolSize,
		ScriptsDir: *scriptsDir,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Println("\nShutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Shutdown error: %v\n", err)
		}
	}()

	if err := srv.Start(*addr); err != nil {
		// echo returns http.ErrServerClosed on graceful shutdown
		fmt.Printf("Server stopped: %v\n", err)
	}
}
