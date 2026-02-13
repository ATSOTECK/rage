# Shared game balance constants — imported by other config scripts.
#
# This is where cross-script imports shine: change a rarity tier or
# add a material here and every config file picks it up automatically.
# With static formats you'd duplicate these tables in each file.

materials = {
    "wood":    {"damage": 1.0, "weight": 0.8, "value": 10},
    "iron":    {"damage": 1.5, "weight": 1.2, "value": 50},
    "steel":   {"damage": 2.0, "weight": 1.0, "value": 120},
    "mithril": {"damage": 2.5, "weight": 0.5, "value": 500},
    "dragon":  {"damage": 3.5, "weight": 1.5, "value": 2000},
}

rarities = ["common", "uncommon", "rare", "epic", "legendary"]

zones = ["Forest", "Desert", "Mountains", "Swamp", "Volcano"]
