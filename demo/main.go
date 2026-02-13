// Demo: Python-as-Configuration for a game server.
//
// Run from the repo root:
//
//	go run demo/main.go
//	RAGE_ENV=production go run demo/main.go
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ATSOTECK/rage/pkg/rage"
)

func main() {
	// Create a state with only the modules our config scripts need.
	state := rage.NewStateWithModules(
		rage.WithModules(rage.ModuleMath, rage.ModuleDataclasses, rage.ModuleCollections),
	)
	defer state.Close()

	// Register Go functions callable from Python.
	// env(name, default="") — read an environment variable.
	state.Register("env", func(_ *rage.State, args ...rage.Value) rage.Value {
		name, _ := rage.AsString(args[0])
		def := ""
		if len(args) > 1 {
			def, _ = rage.AsString(args[1])
		}
		if v := os.Getenv(name); v != "" {
			return rage.String(v)
		}
		return rage.String(def)
	})

	// log(msg) — config-time logging so scripts can report what they loaded.
	state.Register("log", func(_ *rage.State, args ...rage.Value) rage.Value {
		if len(args) > 0 {
			fmt.Printf("  [config] %s\n", args[0].String())
		}
		return nil
	})

	// Inject a host value the Python side can use for computed settings.
	state.SetGlobal("cpu_count", rage.Int(int64(runtime.NumCPU())))

	// ── Load config files ──────────────────────────────────────────────
	fmt.Println("Loading game server configuration from Python scripts...")
	fmt.Println()

	configs := []struct {
		path   string
		global string // name of the top-level dict each script exports
	}{
		{"demo/config/settings.py", "settings"},
		{"demo/config/items.py", "items"},
		{"demo/config/levels.py", "levels"},
	}

	for _, cfg := range configs {
		src, err := os.ReadFile(cfg.path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", cfg.path, err)
			os.Exit(1)
		}
		if _, err := state.RunWithFilename(string(src), cfg.path); err != nil {
			fmt.Fprintf(os.Stderr, "error in %s: %v\n", cfg.path, err)
			os.Exit(1)
		}
	}

	fmt.Println()

	// ── Extract and display settings ───────────────────────────────────
	printSettings(state)
	printItems(state)
	printLevels(state)
}

// ── Pretty-print helpers ───────────────────────────────────────────────

func printSettings(state *rage.State) {
	s, ok := rage.AsDict(state.GetGlobal("settings"))
	if !ok {
		fmt.Println("(settings not found)")
		return
	}

	section("Server Settings")
	kv("Environment", s["environment"])
	kv("Listen", fmt.Sprintf("%s:%s", s["host"], s["port"]))
	kv("Debug", s["debug"])
	kv("Log level", s["log_level"])
	kv("API URL", s["api_url"])
	kv("DB connection", s["db_connection"])
	kv("DB pool size", s["db_pool_size"])
	kv("Max connections", s["max_connections"])

	if features, ok := rage.AsDict(s["features"]); ok {
		fmt.Println()
		fmt.Println("  Feature flags:")
		for name, val := range features {
			fmt.Printf("    %-20s %s\n", name, val)
		}
	}
	fmt.Println()
}

func printItems(state *rage.State) {
	it, ok := rage.AsDict(state.GetGlobal("items"))
	if !ok {
		fmt.Println("(items not found)")
		return
	}

	section("Weapon Catalog")

	if weapons, ok := rage.AsList(it["weapons"]); ok {
		// Show a sample of weapons (first of each material).
		seen := map[string]bool{}
		for _, w := range weapons {
			wd, ok := rage.AsDict(w)
			if !ok {
				continue
			}
			mat := wd["material"].String()
			if seen[mat] {
				continue
			}
			seen[mat] = true
			fmt.Printf("  %-28s dmg=%-4s wt=%-5s val=%-6s [%s]\n",
				wd["name"], wd["damage"], wd["weight"], wd["value"], wd["rarity"])
		}
		fmt.Printf("\n  (%d weapons total across all tiers)\n", len(weapons))
	}

	fmt.Println()
	if loot, ok := rage.AsDict(it["loot_table"]); ok {
		fmt.Println("  Loot drop rates:")
		for rarity, pct := range loot {
			fmt.Printf("    %-12s %s%%\n", rarity, pct)
		}
	}

	if bonus, ok := rage.AsDict(it["set_bonus"]); ok {
		fmt.Printf("\n  Set bonus: %s — %s pieces, +%s%% damage (%s total base dmg)\n",
			bonus["name"], bonus["pieces_required"], bonus["bonus_damage_pct"], bonus["total_damage"])
	}
	fmt.Println()
}

func printLevels(state *rage.State) {
	lv, ok := rage.AsDict(state.GetGlobal("levels"))
	if !ok {
		fmt.Println("(levels not found)")
		return
	}

	section("Level Progression")

	if xp, ok := rage.AsList(lv["xp_curve"]); ok {
		fmt.Println("  XP required (sample):")
		samples := []int{0, 4, 9, 19, 29, 39, 49}
		for _, i := range samples {
			if i < len(xp) {
				fmt.Printf("    Level %-3d  %s XP\n", i+1, xp[i])
			}
		}
	}

	if stats, ok := rage.AsList(lv["level_stats"]); ok {
		fmt.Println("\n  Stats at key levels:")
		for _, lvl := range []int{0, 9, 24, 49} {
			if lvl < len(stats) {
				if st, ok := rage.AsDict(stats[lvl]); ok {
					fmt.Printf("    Level %-3d  HP=%-5s ATK=%-4s DEF=%-4s SPD=%s\n",
						lvl+1, st["hp"], st["attack"], st["defense"], st["speed"])
				}
			}
		}
	}

	if zones, ok := rage.AsDict(lv["zone_scaling"]); ok {
		fmt.Println("\n  Zone difficulty multipliers:")
		for zone, mult := range zones {
			fmt.Printf("    %-12s x%s\n", zone, mult)
		}
	}

	if bosses, ok := rage.AsDict(lv["bosses"]); ok {
		fmt.Println("\n  Bosses:")
		for lvl, boss := range bosses {
			if bd, ok := rage.AsDict(boss); ok {
				fmt.Printf("    Level %-3s  %-22s HP=%-6s ATK=%-4s XP=%s\n",
					lvl, bd["name"], bd["hp"], bd["attack"], bd["xp_reward"])
			}
		}
	}

	fmt.Println()
}

func section(title string) {
	fmt.Printf("── %s %s\n", title, strings.Repeat("─", 58-len(title)))
}

func kv(label string, val interface{}) {
	fmt.Printf("  %-20s %v\n", label, val)
}
