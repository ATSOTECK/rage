// Demo: Python-as-Configuration for a game server.
//
// Run from the repo root:
//
//	go run demo/main.go
//	RAGE_ENV=production go run demo/main.go
package main

import (
	"fmt"
	"math"
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

	// Register Go-defined game classes (Vec2, Color, Inventory, GameSession).
	if err := registerGameClasses(state); err != nil {
		fmt.Fprintf(os.Stderr, "error registering game classes: %v\n", err)
		os.Exit(1)
	}

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
		{"demo/config/entities.py", "entities"},
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
	printEntities(state)
}

// ── Go-defined game classes ────────────────────────────────────────────

// inventoryData stores ordered items for an Inventory instance.
type inventoryData struct {
	capacity int64
	keys     []string
	items    map[string]rage.Value
	iterIdx  int
}

// sessionData stores event log for a GameSession instance.
type sessionData struct {
	name   string
	events []rage.Value
}

func registerGameClasses(state *rage.State) error {
	// ── Vec2(x, y) — 2D position/vector ──────────────────────────────
	// Showcases: Init, Add, Sub, Mul, Neg, Abs, Eq, Hash, Str, Repr,
	//            Property, StaticMethod, Attr, Round
	// Methods:   distance_to, normalized, dot, lerp
	// ClassMethod: from_angle
	var vec2Class rage.ClassValue
	vec2Class = rage.NewClass("Vec2").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			if len(args) < 2 {
				return rage.TypeError("Vec2() requires x and y")
			}
			self.Set("x", args[0])
			self.Set("y", args[1])
			return nil
		}).
		Str(func(s *rage.State, self rage.Object) (string, error) {
			x, _ := rage.AsFloat(self.Get("x"))
			y, _ := rage.AsFloat(self.Get("y"))
			return fmt.Sprintf("Vec2(%g, %g)", x, y), nil
		}).
		Repr(func(s *rage.State, self rage.Object) (string, error) {
			x, _ := rage.AsFloat(self.Get("x"))
			y, _ := rage.AsFloat(self.Get("y"))
			return fmt.Sprintf("Vec2(%g, %g)", x, y), nil
		}).
		Add(func(s *rage.State, self rage.Object, other rage.Value) (rage.Value, error) {
			o, ok := rage.AsObject(other)
			if !ok {
				return nil, rage.TypeError("can only add Vec2 to Vec2")
			}
			x1, _ := rage.AsFloat(self.Get("x"))
			y1, _ := rage.AsFloat(self.Get("y"))
			x2, _ := rage.AsFloat(o.Get("x"))
			y2, _ := rage.AsFloat(o.Get("y"))
			result := self.Class().NewInstance()
			result.Set("x", rage.Float(x1+x2))
			result.Set("y", rage.Float(y1+y2))
			return result, nil
		}).
		Sub(func(s *rage.State, self rage.Object, other rage.Value) (rage.Value, error) {
			o, ok := rage.AsObject(other)
			if !ok {
				return nil, rage.TypeError("can only subtract Vec2 from Vec2")
			}
			x1, _ := rage.AsFloat(self.Get("x"))
			y1, _ := rage.AsFloat(self.Get("y"))
			x2, _ := rage.AsFloat(o.Get("x"))
			y2, _ := rage.AsFloat(o.Get("y"))
			result := self.Class().NewInstance()
			result.Set("x", rage.Float(x1-x2))
			result.Set("y", rage.Float(y1-y2))
			return result, nil
		}).
		Mul(func(s *rage.State, self rage.Object, other rage.Value) (rage.Value, error) {
			scalar, ok := rage.AsFloat(other)
			if !ok {
				return nil, rage.TypeError("can only multiply Vec2 by a number")
			}
			x, _ := rage.AsFloat(self.Get("x"))
			y, _ := rage.AsFloat(self.Get("y"))
			result := self.Class().NewInstance()
			result.Set("x", rage.Float(x*scalar))
			result.Set("y", rage.Float(y*scalar))
			return result, nil
		}).
		Neg(func(s *rage.State, self rage.Object) (rage.Value, error) {
			x, _ := rage.AsFloat(self.Get("x"))
			y, _ := rage.AsFloat(self.Get("y"))
			result := self.Class().NewInstance()
			result.Set("x", rage.Float(-x))
			result.Set("y", rage.Float(-y))
			return result, nil
		}).
		Abs(func(s *rage.State, self rage.Object) (rage.Value, error) {
			x, _ := rage.AsFloat(self.Get("x"))
			y, _ := rage.AsFloat(self.Get("y"))
			return rage.Float(math.Sqrt(x*x + y*y)), nil
		}).
		Eq(func(s *rage.State, self rage.Object, other rage.Value) (bool, error) {
			o, ok := rage.AsObject(other)
			if !ok {
				return false, nil
			}
			x1, _ := rage.AsFloat(self.Get("x"))
			y1, _ := rage.AsFloat(self.Get("y"))
			x2, _ := rage.AsFloat(o.Get("x"))
			y2, _ := rage.AsFloat(o.Get("y"))
			return x1 == x2 && y1 == y2, nil
		}).
		Hash(func(s *rage.State, self rage.Object) (int64, error) {
			x, _ := rage.AsFloat(self.Get("x"))
			y, _ := rage.AsFloat(self.Get("y"))
			xBits := math.Float64bits(x)
			yBits := math.Float64bits(y)
			return int64(xBits ^ (yBits * 0x9e3779b97f4a7c15)), nil
		}).
		Property("length", func(s *rage.State, self rage.Object) (rage.Value, error) {
			x, _ := rage.AsFloat(self.Get("x"))
			y, _ := rage.AsFloat(self.Get("y"))
			return rage.Float(math.Sqrt(x*x + y*y)), nil
		}).
		StaticMethod("zero", func(s *rage.State, args ...rage.Value) (rage.Value, error) {
			result := vec2Class.NewInstance()
			result.Set("x", rage.Int(0))
			result.Set("y", rage.Int(0))
			return result, nil
		}).
		Round(func(s *rage.State, self rage.Object, ndigits rage.Value) (rage.Value, error) {
			x, _ := rage.AsFloat(self.Get("x"))
			y, _ := rage.AsFloat(self.Get("y"))
			n := 0
			if ndigits != nil {
				if ni, ok := rage.AsInt(ndigits); ok {
					n = int(ni)
				}
			}
			factor := math.Pow(10, float64(n))
			result := self.Class().NewInstance()
			result.Set("x", rage.Float(math.Round(x*factor)/factor))
			result.Set("y", rage.Float(math.Round(y*factor)/factor))
			return result, nil
		}).
		// distance_to(other) — Euclidean distance between two points.
		Method("distance_to", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			o, ok := rage.AsObject(args[0])
			if !ok {
				return nil, rage.TypeError("distance_to() requires a Vec2")
			}
			x1, _ := rage.AsFloat(self.Get("x"))
			y1, _ := rage.AsFloat(self.Get("y"))
			x2, _ := rage.AsFloat(o.Get("x"))
			y2, _ := rage.AsFloat(o.Get("y"))
			dx, dy := x2-x1, y2-y1
			return rage.Float(math.Sqrt(dx*dx + dy*dy)), nil
		}).
		// normalized() — return unit vector (same direction, length 1).
		Method("normalized", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			x, _ := rage.AsFloat(self.Get("x"))
			y, _ := rage.AsFloat(self.Get("y"))
			length := math.Sqrt(x*x + y*y)
			if length == 0 {
				return nil, rage.ValueError("cannot normalize zero vector")
			}
			result := self.Class().NewInstance()
			result.Set("x", rage.Float(x/length))
			result.Set("y", rage.Float(y/length))
			return result, nil
		}).
		// dot(other) — dot product of two vectors.
		Method("dot", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			o, ok := rage.AsObject(args[0])
			if !ok {
				return nil, rage.TypeError("dot() requires a Vec2")
			}
			x1, _ := rage.AsFloat(self.Get("x"))
			y1, _ := rage.AsFloat(self.Get("y"))
			x2, _ := rage.AsFloat(o.Get("x"))
			y2, _ := rage.AsFloat(o.Get("y"))
			return rage.Float(x1*x2 + y1*y2), nil
		}).
		// lerp(other, t) — linear interpolation toward other by factor t.
		Method("lerp", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			o, ok := rage.AsObject(args[0])
			if !ok {
				return nil, rage.TypeError("lerp() requires a Vec2 and a number")
			}
			t, ok := rage.AsFloat(args[1])
			if !ok {
				return nil, rage.TypeError("lerp() requires a Vec2 and a number")
			}
			x1, _ := rage.AsFloat(self.Get("x"))
			y1, _ := rage.AsFloat(self.Get("y"))
			x2, _ := rage.AsFloat(o.Get("x"))
			y2, _ := rage.AsFloat(o.Get("y"))
			result := self.Class().NewInstance()
			result.Set("x", rage.Float(x1+(x2-x1)*t))
			result.Set("y", rage.Float(y1+(y2-y1)*t))
			return result, nil
		}).
		// from_angle(degrees, length=1) — create Vec2 from angle.
		ClassMethod("from_angle", func(s *rage.State, cls rage.ClassValue, args ...rage.Value) (rage.Value, error) {
			deg, _ := rage.AsFloat(args[0])
			length := 1.0
			if len(args) > 1 {
				if l, ok := rage.AsFloat(args[1]); ok {
					length = l
				}
			}
			rad := deg * math.Pi / 180
			result := cls.NewInstance()
			result.Set("x", rage.Float(math.Cos(rad)*length))
			result.Set("y", rage.Float(math.Sin(rad)*length))
			return result, nil
		}).
		Build(state)

	state.SetGlobal("Vec2", vec2Class)
	// Set ORIGIN as a class attribute using Python (demonstrates Go+Python interop).
	if _, err := state.Run("Vec2.ORIGIN = Vec2(0, 0)"); err != nil {
		return fmt.Errorf("setting Vec2.ORIGIN: %w", err)
	}

	// ── Color(r, g, b) — RGB color ──────────────────────────────────
	// Showcases: Init, Add, Mul, Eq, Str, Format, IntConv, FloatConv,
	//            ClassGetItem
	// Methods:   inverted, grayscale, mix (MethodKw)
	// ClassMethod: from_hex
	color := rage.NewClass("Color").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			if len(args) < 3 {
				return rage.TypeError("Color() requires r, g, b")
			}
			self.Set("r", args[0])
			self.Set("g", args[1])
			self.Set("b", args[2])
			return nil
		}).
		Str(func(s *rage.State, self rage.Object) (string, error) {
			r, _ := rage.AsInt(self.Get("r"))
			g, _ := rage.AsInt(self.Get("g"))
			b, _ := rage.AsInt(self.Get("b"))
			return fmt.Sprintf("Color(%d, %d, %d)", r, g, b), nil
		}).
		Add(func(s *rage.State, self rage.Object, other rage.Value) (rage.Value, error) {
			o, ok := rage.AsObject(other)
			if !ok {
				return nil, rage.TypeError("can only add Color to Color")
			}
			r1, _ := rage.AsInt(self.Get("r"))
			g1, _ := rage.AsInt(self.Get("g"))
			b1, _ := rage.AsInt(self.Get("b"))
			r2, _ := rage.AsInt(o.Get("r"))
			g2, _ := rage.AsInt(o.Get("g"))
			b2, _ := rage.AsInt(o.Get("b"))
			result := self.Class().NewInstance()
			result.Set("r", rage.Int(min(r1+r2, 255)))
			result.Set("g", rage.Int(min(g1+g2, 255)))
			result.Set("b", rage.Int(min(b1+b2, 255)))
			return result, nil
		}).
		Mul(func(s *rage.State, self rage.Object, other rage.Value) (rage.Value, error) {
			scalar, ok := rage.AsFloat(other)
			if !ok {
				return nil, rage.TypeError("can only multiply Color by a number")
			}
			r, _ := rage.AsInt(self.Get("r"))
			g, _ := rage.AsInt(self.Get("g"))
			b, _ := rage.AsInt(self.Get("b"))
			clamp := func(v float64) int64 {
				return max(0, min(int64(v), 255))
			}
			result := self.Class().NewInstance()
			result.Set("r", rage.Int(clamp(float64(r)*scalar)))
			result.Set("g", rage.Int(clamp(float64(g)*scalar)))
			result.Set("b", rage.Int(clamp(float64(b)*scalar)))
			return result, nil
		}).
		Eq(func(s *rage.State, self rage.Object, other rage.Value) (bool, error) {
			o, ok := rage.AsObject(other)
			if !ok {
				return false, nil
			}
			r1, _ := rage.AsInt(self.Get("r"))
			g1, _ := rage.AsInt(self.Get("g"))
			b1, _ := rage.AsInt(self.Get("b"))
			r2, _ := rage.AsInt(o.Get("r"))
			g2, _ := rage.AsInt(o.Get("g"))
			b2, _ := rage.AsInt(o.Get("b"))
			return r1 == r2 && g1 == g2 && b1 == b2, nil
		}).
		Format(func(s *rage.State, self rage.Object, spec string) (string, error) {
			r, _ := rage.AsInt(self.Get("r"))
			g, _ := rage.AsInt(self.Get("g"))
			b, _ := rage.AsInt(self.Get("b"))
			if spec == "hex" {
				return fmt.Sprintf("#%02x%02x%02x", r, g, b), nil
			}
			return fmt.Sprintf("Color(%d, %d, %d)", r, g, b), nil
		}).
		IntConv(func(s *rage.State, self rage.Object) (int64, error) {
			r, _ := rage.AsInt(self.Get("r"))
			g, _ := rage.AsInt(self.Get("g"))
			b, _ := rage.AsInt(self.Get("b"))
			return (r << 16) | (g << 8) | b, nil
		}).
		FloatConv(func(s *rage.State, self rage.Object) (float64, error) {
			r, _ := rage.AsInt(self.Get("r"))
			g, _ := rage.AsInt(self.Get("g"))
			b, _ := rage.AsInt(self.Get("b"))
			// ITU-R BT.709 luminance, normalized to 0.0–1.0
			return (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)) / 255.0, nil
		}).
		ClassGetItem(func(s *rage.State, cls rage.ClassValue, key rage.Value) (rage.Value, error) {
			return rage.String(fmt.Sprintf("Color[%s]", key.String())), nil
		}).
		// inverted() — return the complementary color.
		Method("inverted", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			r, _ := rage.AsInt(self.Get("r"))
			g, _ := rage.AsInt(self.Get("g"))
			b, _ := rage.AsInt(self.Get("b"))
			result := self.Class().NewInstance()
			result.Set("r", rage.Int(255-r))
			result.Set("g", rage.Int(255-g))
			result.Set("b", rage.Int(255-b))
			return result, nil
		}).
		// grayscale() — convert to grayscale using luminance weights.
		Method("grayscale", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			r, _ := rage.AsInt(self.Get("r"))
			g, _ := rage.AsInt(self.Get("g"))
			b, _ := rage.AsInt(self.Get("b"))
			gray := int64(0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b))
			result := self.Class().NewInstance()
			result.Set("r", rage.Int(gray))
			result.Set("g", rage.Int(gray))
			result.Set("b", rage.Int(gray))
			return result, nil
		}).
		// mix(other, t=0.5) — blend two colors by factor t (0.0 = self, 1.0 = other).
		MethodKw("mix", func(s *rage.State, self rage.Object, args []rage.Value, kwargs map[string]rage.Value) (rage.Value, error) {
			o, ok := rage.AsObject(args[0])
			if !ok {
				return nil, rage.TypeError("mix() requires a Color")
			}
			t := 0.5
			if tv, ok := kwargs["t"]; ok {
				if tf, ok := rage.AsFloat(tv); ok {
					t = tf
				}
			} else if len(args) > 1 {
				if tf, ok := rage.AsFloat(args[1]); ok {
					t = tf
				}
			}
			r1, _ := rage.AsInt(self.Get("r"))
			g1, _ := rage.AsInt(self.Get("g"))
			b1, _ := rage.AsInt(self.Get("b"))
			r2, _ := rage.AsInt(o.Get("r"))
			g2, _ := rage.AsInt(o.Get("g"))
			b2, _ := rage.AsInt(o.Get("b"))
			lerp := func(a, b int64, f float64) int64 {
				return int64(float64(a)*(1-f) + float64(b)*f)
			}
			result := self.Class().NewInstance()
			result.Set("r", rage.Int(lerp(r1, r2, t)))
			result.Set("g", rage.Int(lerp(g1, g2, t)))
			result.Set("b", rage.Int(lerp(b1, b2, t)))
			return result, nil
		}).
		// from_hex("#rrggbb") — create Color from hex string.
		ClassMethod("from_hex", func(s *rage.State, cls rage.ClassValue, args ...rage.Value) (rage.Value, error) {
			hex, _ := rage.AsString(args[0])
			if len(hex) > 0 && hex[0] == '#' {
				hex = hex[1:]
			}
			if len(hex) != 6 {
				return nil, rage.ValueError("from_hex() requires a 6-digit hex string")
			}
			var r, g, b int64
			if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
				return nil, rage.ValueError(fmt.Sprintf("invalid hex color: %q", hex))
			}
			result := cls.NewInstance()
			result.Set("r", rage.Int(r))
			result.Set("g", rage.Int(g))
			result.Set("b", rage.Int(b))
			return result, nil
		}).
		Build(state)

	state.SetGlobal("Color", color)

	// ── Inventory(capacity) — game container ─────────────────────────
	// Showcases: Init, Len, GetItem, SetItem, DelItem, Contains, Bool,
	//            Str, IAdd, Reversed, Iter, Next
	// Methods:   keys, total_weight, drop
	inventory := rage.NewClass("Inventory").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			cap := int64(16)
			if len(args) > 0 {
				if c, ok := rage.AsInt(args[0]); ok {
					cap = c
				}
			}
			self.Set("_data", rage.UserData(&inventoryData{
				capacity: cap,
				keys:     nil,
				items:    make(map[string]rage.Value),
			}))
			return nil
		}).
		Str(func(s *rage.State, self rage.Object) (string, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			return fmt.Sprintf("Inventory(%d/%d)", len(inv.keys), inv.capacity), nil
		}).
		Len(func(s *rage.State, self rage.Object) (int64, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			return int64(len(inv.keys)), nil
		}).
		GetItem(func(s *rage.State, self rage.Object, key rage.Value) (rage.Value, error) {
			name, _ := rage.AsString(key)
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			if v, ok := inv.items[name]; ok {
				return v, nil
			}
			return nil, rage.KeyError(fmt.Sprintf("'%s'", name))
		}).
		SetItem(func(s *rage.State, self rage.Object, key, val rage.Value) error {
			name, _ := rage.AsString(key)
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			if int64(len(inv.keys)) >= inv.capacity {
				if _, exists := inv.items[name]; !exists {
					return rage.ValueError("inventory is full")
				}
			}
			if _, exists := inv.items[name]; !exists {
				inv.keys = append(inv.keys, name)
			}
			inv.items[name] = val
			return nil
		}).
		DelItem(func(s *rage.State, self rage.Object, key rage.Value) error {
			name, _ := rage.AsString(key)
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			if _, ok := inv.items[name]; !ok {
				return rage.KeyError(fmt.Sprintf("'%s'", name))
			}
			delete(inv.items, name)
			for i, k := range inv.keys {
				if k == name {
					inv.keys = append(inv.keys[:i], inv.keys[i+1:]...)
					break
				}
			}
			return nil
		}).
		Contains(func(s *rage.State, self rage.Object, item rage.Value) (bool, error) {
			name, _ := rage.AsString(item)
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			_, ok := inv.items[name]
			return ok, nil
		}).
		Bool(func(s *rage.State, self rage.Object) (bool, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			return len(inv.keys) > 0, nil
		}).
		IAdd(func(s *rage.State, self rage.Object, other rage.Value) (rage.Value, error) {
			name, ok := rage.AsString(other)
			if !ok {
				return nil, rage.TypeError("can only += a string item name")
			}
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			if _, exists := inv.items[name]; !exists {
				if int64(len(inv.keys)) >= inv.capacity {
					return nil, rage.ValueError("inventory is full")
				}
				inv.keys = append(inv.keys, name)
				inv.items[name] = rage.None
			}
			return self, nil
		}).
		Reversed(func(s *rage.State, self rage.Object) (rage.Value, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			reversed := make([]rage.Value, len(inv.keys))
			for i, k := range inv.keys {
				reversed[len(inv.keys)-1-i] = rage.String(k)
			}
			return rage.List(reversed...), nil
		}).
		Iter(func(s *rage.State, self rage.Object) (rage.Value, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			inv.iterIdx = 0
			return self, nil
		}).
		Next(func(s *rage.State, self rage.Object) (rage.Value, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			if inv.iterIdx >= len(inv.keys) {
				return nil, rage.ErrStopIteration
			}
			key := inv.keys[inv.iterIdx]
			inv.iterIdx++
			return rage.String(key), nil
		}).
		// keys() — return list of item names in insertion order.
		Method("keys", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			items := make([]rage.Value, len(inv.keys))
			for i, k := range inv.keys {
				items[i] = rage.String(k)
			}
			return rage.List(items...), nil
		}).
		// total_weight() — sum the "weight" field across all items.
		Method("total_weight", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			total := int64(0)
			for _, v := range inv.items {
				if d, ok := rage.AsDict(v); ok {
					if w, ok := rage.AsInt(d["weight"]); ok {
						total += w
					}
				}
			}
			return rage.Int(total), nil
		}).
		// drop(name) — remove an item and return it (or None if not found).
		Method("drop", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			name, _ := rage.AsString(args[0])
			raw, _ := rage.AsUserData(self.Get("_data"))
			inv := raw.(*inventoryData)
			val, ok := inv.items[name]
			if !ok {
				return rage.None, nil
			}
			delete(inv.items, name)
			for i, k := range inv.keys {
				if k == name {
					inv.keys = append(inv.keys[:i], inv.keys[i+1:]...)
					break
				}
			}
			return val, nil
		}).
		Build(state)

	state.SetGlobal("Inventory", inventory)

	// ── GameSession(name) — context manager + event logger ───────────
	// Showcases: Init, Enter, Exit, CallKw, GetAttr
	// Methods:   filter, event_count
	session := rage.NewClass("GameSession").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			name := "unnamed"
			if len(args) > 0 {
				if n, ok := rage.AsString(args[0]); ok {
					name = n
				}
			}
			self.Set("_data", rage.UserData(&sessionData{
				name:   name,
				events: nil,
			}))
			return nil
		}).
		Str(func(s *rage.State, self rage.Object) (string, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			sd := raw.(*sessionData)
			return fmt.Sprintf("GameSession(%q, %d events)", sd.name, len(sd.events)), nil
		}).
		Enter(func(s *rage.State, self rage.Object) (rage.Value, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			sd := raw.(*sessionData)
			fmt.Printf("  [session] %s: started\n", sd.name)
			return self, nil
		}).
		Exit(func(s *rage.State, self rage.Object, excType, excVal, excTb rage.Value) (bool, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			sd := raw.(*sessionData)
			fmt.Printf("  [session] %s: closed (%d events logged)\n", sd.name, len(sd.events))
			return false, nil // don't suppress exceptions
		}).
		CallKw(func(s *rage.State, self rage.Object, args []rage.Value, kwargs map[string]rage.Value) (rage.Value, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			sd := raw.(*sessionData)
			// Build a dict from the keyword arguments.
			pairs := make([]any, 0, len(kwargs)*2)
			for k, v := range kwargs {
				pairs = append(pairs, k, v)
			}
			sd.events = append(sd.events, rage.Dict(pairs...))
			return rage.None, nil
		}).
		GetAttr(func(s *rage.State, self rage.Object, name string) (rage.Value, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			sd := raw.(*sessionData)
			switch name {
			case "name":
				return rage.String(sd.name), nil
			case "events":
				return rage.List(sd.events...), nil
			}
			return nil, fmt.Errorf("AttributeError: 'GameSession' has no attribute '%s'", name)
		}).
		// filter(event_name) — return events matching the given event name.
		Method("filter", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			name, _ := rage.AsString(args[0])
			raw, _ := rage.AsUserData(self.Get("_data"))
			sd := raw.(*sessionData)
			var matched []rage.Value
			for _, ev := range sd.events {
				if d, ok := rage.AsDict(ev); ok {
					if evName, ok := rage.AsString(d["event"]); ok && evName == name {
						matched = append(matched, ev)
					}
				}
			}
			return rage.List(matched...), nil
		}).
		// event_count() — return total number of logged events.
		Method("event_count", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			raw, _ := rage.AsUserData(self.Get("_data"))
			sd := raw.(*sessionData)
			return rage.Int(int64(len(sd.events))), nil
		}).
		Build(state)

	state.SetGlobal("GameSession", session)

	return nil
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

func printEntities(state *rage.State) {
	e, ok := rage.AsDict(state.GetGlobal("entities"))
	if !ok {
		fmt.Println("(entities not found)")
		return
	}

	section("Game Entities (ClassBuilder)")

	if sp, ok := rage.AsList(e["spawn_points"]); ok {
		fmt.Println("  Spawn points:")
		for _, p := range sp {
			fmt.Printf("    %s\n", p)
		}
	}

	if tp, ok := rage.AsList(e["tower_positions"]); ok {
		fmt.Println("  Tower positions (computed):")
		for _, p := range tp {
			fmt.Printf("    %s\n", p)
		}
	}

	fmt.Println("  Vec2 methods:")
	kv("    Patrol distance", e["patrol_distance"])
	kv("    Direction", e["direction"])
	kv("    Facing dot", e["facing_dot"])
	kv("    Midpoint", e["midpoint"])
	kv("    From angle(45)", e["angled"])

	if bc, ok := rage.AsDict(e["biome_colors"]); ok {
		fmt.Println("  Biome colors:")
		for name, colors := range bc {
			if cd, ok := rage.AsDict(colors); ok {
				fmt.Printf("    %-12s sky=%-9s ground=%s\n", name, cd["sky"], cd["ground"])
			}
		}
	}

	fmt.Println()
	if sk, ok := rage.AsList(e["starter_keys"]); ok {
		names := make([]string, len(sk))
		for i, k := range sk {
			names[i] = k.String()
		}
		kv("Starter keys", strings.Join(names, ", "))
	}
	kv("Carry weight", e["carry_weight"])
	kv("Dropped potion", e["dropped_potion"])
	kv("Starter items", e["starter_items"])
	kv("Has sword", e["starter_has_sword"])

	if ui, ok := rage.AsDict(e["ui_colors"]); ok {
		fmt.Println()
		fmt.Println("  UI colors:")
		for name, hex := range ui {
			fmt.Printf("    %-12s %s\n", name, hex)
		}
	}

	if events, ok := rage.AsList(e["session_events"]); ok {
		fmt.Println()
		fmt.Println("  Config session events:")
		for _, ev := range events {
			if ed, ok := rage.AsDict(ev); ok {
				parts := []string{}
				for k, v := range ed {
					parts = append(parts, fmt.Sprintf("%s=%s", k, v))
				}
				fmt.Printf("    {%s}\n", strings.Join(parts, ", "))
			}
		}
	}

	kv("Spawn events", e["spawn_event_count"])
	kv("Total events", e["total_events"])

	fmt.Println()
}

func section(title string) {
	fmt.Printf("── %s %s\n", title, strings.Repeat("─", 58-len(title)))
}

func kv(label string, val interface{}) {
	fmt.Printf("  %-20s %v\n", label, val)
}
