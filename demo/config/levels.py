# Level progression — generated tables and formulas.
#
# 50 levels of XP curves, stat scaling, milestones, and bosses,
# all from a handful of formulas. A static config would be 500+ lines.

import math
from common import zones

max_level = 50

# XP curve: smooth exponential ramp using math.pow/math.floor
xp_for_level = [math.floor(100 * math.pow(1.15, level)) for level in range(max_level)]

# Stat scaling at each level
def stats_at_level(level):
    return {
        "hp":      100 + level * 15 + math.floor(math.sqrt(level) * 10),
        "attack":  10 + level * 3,
        "defense": 5 + level * 2 + (level // 10) * 5,
        "speed":   10 + math.ceil(level * 0.8),
    }

level_stats = [stats_at_level(lvl) for lvl in range(max_level)]

# Milestone rewards at specific levels
milestones = {}
for lvl in range(max_level):
    if lvl == 5:
        milestones[lvl] = {"type": "unlock", "reward": "Dual Wield"}
    elif lvl == 10:
        milestones[lvl] = {"type": "unlock", "reward": "Magic System"}
    elif lvl == 25:
        milestones[lvl] = {"type": "title", "reward": "Veteran"}
    elif lvl == 49:
        milestones[lvl] = {"type": "title", "reward": "Grand Master"}
    elif lvl % 10 == 9 and lvl > 0:
        milestones[lvl] = {"type": "stat_boost", "reward": "All Stats +5"}

# Zone difficulty scaling (zones imported from common.py)
zone_scaling = {zones[i]: round(1.0 + i * 0.35, 2) for i in range(len(zones))}

# Boss at every 10th level with scaled stats
bosses = {}
for lvl in range(9, max_level, 10):
    base = stats_at_level(lvl)
    zone_idx = min(lvl // 10, len(zones) - 1)
    bosses[lvl] = {
        "name": f"{zones[zone_idx]} Guardian",
        "hp": base["hp"] * 10,
        "attack": base["attack"] * 3,
        "defense": base["defense"] * 2,
        "xp_reward": xp_for_level[lvl] * 5,
    }

# --- Final export ---

levels = {
    "max_level": max_level,
    "xp_curve": xp_for_level,
    "level_stats": level_stats,
    "milestones": milestones,
    "zone_scaling": zone_scaling,
    "bosses": bosses,
}

log(f"  Generated {max_level} levels, {len(milestones)} milestones, {len(bosses)} bosses")
