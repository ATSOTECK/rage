package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ATSOTECK/rage/pkg/rage"
	"github.com/ATSOTECK/rage/test/stress_test/internal/pool"
	"github.com/labstack/echo/v4"
)

// Config holds server configuration.
type Config struct {
	PoolSize   int
	ScriptsDir string
}

// companyScript maps company IDs to script filenames (without .py).
var companyScript = map[string]string{
	"discount_co":  "discount",
	"shipping_co":  "shipping",
	"validator_co": "validator",
	"transform_co": "transform",
	"tax_co":       "tax",
}

// Server is the multi-tenant item processing server.
type Server struct {
	echo  *echo.Echo
	pools map[string]*pool.StatePool
}

// processRequest is the JSON request body.
type processRequest struct {
	CompanyID string         `json:"company_id"`
	Item      map[string]any `json:"item"`
}

// New creates a new server, loading scripts and creating state pools.
func New(cfg Config) (*Server, error) {
	s := &Server{
		echo:  echo.New(),
		pools: make(map[string]*pool.StatePool),
	}
	s.echo.HideBanner = true
	s.echo.HidePort = true

	// Shared mutex for serializing RAGE state creation across all pools.
	// rage.NewStateWithModules calls runtime.ResetModules() (global state).
	var createMu sync.Mutex

	// Load and compile all scripts, create pools
	for companyID, scriptName := range companyScript {
		scriptPath := filepath.Join(cfg.ScriptsDir, scriptName+".py")
		src, err := os.ReadFile(scriptPath)
		if err != nil {
			return nil, fmt.Errorf("read script %s: %w", scriptPath, err)
		}

		// Compile once (immutable *rage.Code)
		// Use a temporary state just for compilation
		tmpState := rage.NewBareState()
		code, err := tmpState.Compile(string(src), scriptName+".py")
		tmpState.Close()
		if err != nil {
			return nil, fmt.Errorf("compile %s: %w", scriptName, err)
		}

		p, err := pool.New(companyID, cfg.PoolSize, code, &createMu)
		if err != nil {
			return nil, fmt.Errorf("create pool %s: %w", companyID, err)
		}
		s.pools[companyID] = p

		// Grab one state to validate process_item exists, then return it
		h, err := p.Borrow(context.Background())
		if err != nil {
			return nil, fmt.Errorf("borrow for validation %s: %w", companyID, err)
		}
		fn := h.State.GetGlobal("process_item")
		if fn == nil {
			p.Return(h)
			return nil, fmt.Errorf("script %s does not define process_item", scriptName)
		}
		p.Return(h)
	}

	// Register routes
	s.echo.POST("/process", s.handleProcess)
	s.echo.GET("/stats", s.handleStats)
	s.echo.POST("/gc", s.handleGC)
	s.echo.GET("/diag", s.handleDiag)

	return s, nil
}

func (s *Server) handleProcess(c echo.Context) error {
	var req processRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	companyID := strings.TrimSpace(req.CompanyID)
	p, ok := s.pools[companyID]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "unknown company_id: " + companyID})
	}

	h, err := p.Borrow(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "pool exhausted"})
	}
	defer p.Return(h)

	// Convert Go map to rage.Dict
	itemDict := rage.FromGo(req.Item)

	// Call process_item(item) — get function from the state's globals each time
	// since the function object is tied to the state
	fn := h.State.GetGlobal("process_item")
	result, err := h.State.Call(fn, itemDict)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result.GoValue())
}

func (s *Server) handleGC(c echo.Context) error {
	// Two GC passes: first collects, second cleans up finalizers.
	runtime.GC()
	runtime.GC()
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDiag(c echo.Context) error {
	result := map[string]any{
		"string_intern_pool_size": rage.StringInternPoolSize(),
	}
	for name, p := range s.pools {
		h, err := p.Borrow(context.Background())
		if err != nil {
			continue
		}
		globals := h.State.GetGlobals()
		result[name] = map[string]any{
			"globals_count": len(globals),
			"alloc_bytes":   h.State.AllocatedBytes(),
		}
		p.Return(h)
	}
	return c.JSON(http.StatusOK, result)
}

func (s *Server) handleStats(c echo.Context) error {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	stats := map[string]any{
		"heap_alloc_mb":     float64(mem.HeapAlloc) / 1024 / 1024,
		"heap_objects":      mem.HeapObjects,
		"total_alloc_mb":    float64(mem.TotalAlloc) / 1024 / 1024,
		"num_gc":            mem.NumGC,
		"gc_pause_total_ms": float64(mem.PauseTotalNs) / 1e6,
	}
	return c.JSON(http.StatusOK, stats)
}

// Start starts the HTTP server.
func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

// Shutdown gracefully shuts down the server and closes all pools.
func (s *Server) Shutdown(ctx context.Context) error {
	err := s.echo.Shutdown(ctx)
	for _, p := range s.pools {
		p.Close()
	}
	return err
}
