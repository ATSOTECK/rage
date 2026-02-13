# Item definitions — templated weapons with computed stats.
#
# A template function generates families of items from parameters.
# In TOML you'd copy-paste each variant by hand.

materials = {
    "wood":    {"damage": 1.0, "weight": 0.8, "value": 10},
    "iron":    {"damage": 1.5, "weight": 1.2, "value": 50},
    "steel":   {"damage": 2.0, "weight": 1.0, "value": 120},
    "mithril": {"damage": 2.5, "weight": 0.5, "value": 500},
    "dragon":  {"damage": 3.5, "weight": 1.5, "value": 2000},
}

rarities = ["common", "uncommon", "rare", "epic", "legendary"]

def make_weapon(name, base_damage, base_weight, material, tier=1):
    """Template: one function defines an entire weapon family."""
    mat = materials[material]
    damage = int(base_damage * mat["damage"] * (1 + tier * 0.25))
    weight = round(base_weight * mat["weight"], 1)
    value = int(mat["value"] * tier * (1 + damage / 10))
    rarity_idx = min(tier - 1, len(rarities) - 1)

    if tier >= 4:
        display_name = f"Enchanted {material.title()} {name}"
    else:
        display_name = f"{material.title()} {name}"

    item = {
        "name": display_name,
        "damage": damage,
        "weight": weight,
        "value": value,
        "material": material,
        "tier": tier,
        "rarity": rarities[rarity_idx],
    }

    # Legendary items get a bonus effect
    if rarity_idx >= 4:
        item["bonus_effect"] = "lifesteal"
        item["bonus_damage"] = int(damage * 0.15)

    return item

# Generate 15 swords from a single template (5 materials x 3 tiers)
swords = [
    make_weapon("Sword", 15, 3.0, mat, tier)
    for mat in ["wood", "iron", "steel", "mithril", "dragon"]
    for tier in [1, 3, 5]
]

# Loot probability table — weights drop exponentially with rarity
loot_weights = {r: round(100 * (0.4 ** i), 2) for i, r in enumerate(rarities)}
total_weight = sum(loot_weights.values())
loot_table = {r: round(w / total_weight * 100, 1) for r, w in loot_weights.items()}

# Set bonus computed from items that share a material
dragon_set = [s for s in swords if s["material"] == "dragon"]
set_bonus = {
    "name": "Dragon Slayer Set",
    "pieces_required": 2,
    "total_damage": sum(s["damage"] for s in dragon_set),
    "bonus_damage_pct": 25,
}

# --- Final export ---

items = {
    "weapons": swords,
    "loot_table": loot_table,
    "set_bonus": set_bonus,
}

log(f"  Loaded {len(swords)} weapons across {len(materials)} materials")
