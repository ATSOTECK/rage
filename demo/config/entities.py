# Game entity definitions using Go-defined classes.
#
# Classes available from the Go host:
#   Vec2(x, y)         — 2D position/vector
#   Color(r, g, b)     — RGB color
#   Inventory(capacity) — game container
#   GameSession(name)  — context manager + event logger

# ── Spawn points & paths ──────────────────────────────────────────

spawn_points = [Vec2(100, 200), Vec2(300, 50), Vec2.zero(), Vec2.ORIGIN]
patrol_path = [Vec2(0, 0), Vec2(10, 0), Vec2(10, 10), Vec2(0, 10)]

# Vec2 arithmetic for computed positions
center = Vec2(400, 300)
offsets = [Vec2(-50, -50), Vec2(50, -50), Vec2(0, 50)]
tower_positions = [center + offset for offset in offsets]

# Vec2 methods: distance, normalized, dot, lerp
patrol_total_distance = 0
for i in range(len(patrol_path) - 1):
    patrol_total_distance += patrol_path[i].distance_to(patrol_path[i + 1])

direction = Vec2(3, 4).normalized()
facing_dot = direction.dot(Vec2(1, 0))
midpoint = spawn_points[0].lerp(spawn_points[1], 0.5)
angled = Vec2.from_angle(45, 100)

# ── Color themes per biome ────────────────────────────────────────

biome_colors = {
    "forest": {"sky": Color(135, 206, 235), "ground": Color(34, 139, 34)},
    "desert": {"sky": Color(255, 223, 0), "ground": Color(210, 180, 140)},
    "volcano": {"sky": Color(178, 34, 34), "ground": Color(64, 64, 64)},
}

# ── Starter inventory ────────────────────────────────────────────

starter = Inventory(10)
starter["sword"] = {"damage": 5, "weight": 3}
starter["shield"] = {"damage": 0, "weight": 5}
starter["potion"] = {"damage": 0, "weight": 1}

# Inventory methods: keys, total_weight, drop
starter_keys = starter.keys()
carry_weight = starter.total_weight()
dropped = starter.drop("potion")
starter["health_potion"] = {"damage": 0, "weight": 1}

# ── Color methods: inverted, grayscale, mix, from_hex ─────────────

sunset = Color(255, 100, 50).mix(Color(100, 0, 150), t=0.4)
gray_sky = Color(135, 206, 235).grayscale()
alert = Color(255, 0, 0).inverted()
gold = Color.from_hex("#ffd700")

# ── Color formatting ─────────────────────────────────────────────

ui_colors = {
    "health_bar": format(Color(255, 0, 0), "hex"),
    "mana_bar": format(Color(0, 100, 255), "hex"),
    "sunset": format(sunset, "hex"),
    "gray_sky": format(gray_sky, "hex"),
    "alert_inv": format(alert, "hex"),
    "gold": format(gold, "hex"),
}

# ── GameSession context manager for setup logging ────────────────

with GameSession("config_load") as session:
    session(event="spawn_loaded", count=len(spawn_points))
    session(event="biomes_loaded", count=len(biome_colors))
    session(event="inventory_ready", slots=len(starter))
    session(event="spawn_loaded", count=len(tower_positions))

# GameSession methods: filter, event_count
spawn_events = session.filter("spawn_loaded")
total_events = session.event_count()

# ── Export ────────────────────────────────────────────────────────

entities = {
    "spawn_points": [str(p) for p in spawn_points],
    "tower_positions": [str(p) for p in tower_positions],
    "patrol_distance": round(patrol_total_distance, 2),
    "direction": str(direction),
    "facing_dot": facing_dot,
    "midpoint": str(midpoint),
    "angled": str(round(angled, 2)),
    "biome_colors": {name: {k: format(c, "hex") for k, c in colors.items()} for name, colors in biome_colors.items()},
    "starter_keys": starter_keys,
    "carry_weight": carry_weight,
    "dropped_potion": dropped is not None,
    "starter_items": len(starter),
    "starter_has_sword": "sword" in starter,
    "ui_colors": ui_colors,
    "session_events": session.events,
    "spawn_event_count": len(spawn_events),
    "total_events": total_events,
}
