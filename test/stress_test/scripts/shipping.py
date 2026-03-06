def process_item(item):
    name = item["name"]
    price = item["price"]
    quantity = item["quantity"]
    category = item["category"]
    weight = item["weight"]
    tags = item["tags"]
    address = item["address"]

    total_weight = weight * quantity
    subtotal = price * quantity

    # Weight-based shipping rates with graduated tiers
    shipping = 0.0
    remaining_weight = total_weight
    # Graduated weight tiers: (max_kg_in_tier, rate_per_kg)
    tiers = [
        (1, 5.99),
        (4, 2.50),
        (15, 1.75),
        (30, 1.25),
        (9999, 0.75),
    ]
    for limit, rate in tiers:
        if remaining_weight <= 0:
            break
        chunk = remaining_weight
        if chunk > limit:
            chunk = limit
        shipping = shipping + chunk * rate
        remaining_weight = remaining_weight - chunk

    # Fragile surcharge for electronics
    if category == "electronics":
        shipping = shipping + 4.99

    # Hazmat check via tags
    hazmat = False
    for tag in tags:
        if "hazmat" in tag.lower() or "flammable" in tag.lower() or "fragile" in tag.lower():
            hazmat = True
            shipping = shipping + 12.50
            break

    # Address-based zone calculation
    zone_rates = {"domestic": 1.0, "regional": 1.5, "international": 2.5}
    zone = address["zone"]
    zone_mult = zone_rates[zone] if zone in zone_rates else 1.0
    shipping = shipping * zone_mult

    # Distance surcharge from zip code
    zip_code = address["zip"]
    zip_num = 0
    for ch in zip_code:
        if ch >= "0" and ch <= "9":
            zip_num = zip_num * 10 + ord(ch) - ord("0")
    distance_surcharge = (zip_num % 50) * 0.10
    shipping = shipping + distance_surcharge

    # Free shipping for orders over $200 (domestic only)
    free_shipping = False
    if subtotal >= 200 and zone == "domestic":
        shipping = 0.0
        free_shipping = True

    total = subtotal + shipping

    # Estimated delivery (business days)
    base_days = {"domestic": 3, "regional": 7, "international": 14}
    est_days = base_days[zone] if zone in base_days else 5
    if total_weight > 20:
        est_days = est_days + 2
    if hazmat:
        est_days = est_days + 3

    # Package count estimate
    max_per_package = 10.0
    packages = int(total_weight / max_per_package) + 1

    return {
        "company": "shipping_co",
        "item_name": name,
        "total_weight": round(total_weight, 2),
        "subtotal": round(subtotal, 2),
        "shipping": round(shipping, 2),
        "free_shipping": free_shipping,
        "hazmat": hazmat,
        "zone": zone,
        "zone_multiplier": zone_mult,
        "distance_surcharge": round(distance_surcharge, 2),
        "estimated_days": est_days,
        "packages": packages,
        "total": round(total, 2),
    }
