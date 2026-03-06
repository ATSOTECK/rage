def process_item(item):
    name = item["name"]
    price = item["price"]
    quantity = item["quantity"]
    category = item["category"]
    tags = item["tags"]

    subtotal = price * quantity

    # Tiered discounts based on subtotal
    if subtotal >= 500:
        discount_pct = 15
    elif subtotal >= 200:
        discount_pct = 10
    elif subtotal >= 50:
        discount_pct = 5
    else:
        discount_pct = 0

    # Tag-based bonus discounts (iterate + string matching)
    bonus_pct = 0
    matched_promos = []
    promo_keywords = {"clearance": 5, "sale": 3, "member": 2, "vip": 4, "seasonal": 1}
    for tag in tags:
        lower_tag = tag.lower().strip()
        for keyword in promo_keywords:
            if keyword in lower_tag:
                bonus_pct = bonus_pct + promo_keywords[keyword]
                matched_promos.append(keyword + "(" + str(promo_keywords[keyword]) + "%)")

    # Cap total discount at 30%
    total_discount_pct = discount_pct + bonus_pct
    if total_discount_pct > 30:
        total_discount_pct = 30

    discount_amount = subtotal * total_discount_pct / 100
    total = subtotal - discount_amount

    # Build quantity breakdown for bulk analysis
    breakdown = []
    remaining = quantity
    tier_sizes = [100, 50, 25, 10, 5, 1]
    for ts in tier_sizes:
        if remaining >= ts:
            count = remaining // ts
            breakdown.append({"tier": ts, "count": count, "subtotal": round(count * ts * price, 2)})
            remaining = remaining % ts

    # Category popularity score (simulate lookup table iteration)
    cat_scores = {
        "electronics": 95, "clothing": 88, "food": 72, "books": 65,
        "toys": 78, "home": 82, "sports": 70,
    }
    pop_score = cat_scores[category] if category in cat_scores else 50

    # Seasonal multiplier based on name hash
    name_hash = 0
    for ch in name:
        name_hash = (name_hash * 31 + ord(ch)) % 1000000
    seasonal_mult = 1.0 + (name_hash % 20) / 100.0

    return {
        "company": "discount_co",
        "item_name": name,
        "category": category,
        "subtotal": round(subtotal, 2),
        "base_discount_pct": discount_pct,
        "bonus_discount_pct": bonus_pct,
        "total_discount_pct": total_discount_pct,
        "matched_promos": matched_promos,
        "discount_amount": round(discount_amount, 2),
        "total": round(total, 2),
        "breakdown": breakdown,
        "popularity_score": pop_score,
        "seasonal_multiplier": round(seasonal_mult, 4),
    }
