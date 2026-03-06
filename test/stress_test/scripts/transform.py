def process_item(item):
    name = item["name"]
    price = item["price"]
    quantity = item["quantity"]
    category = item["category"]
    weight = item["weight"]
    tags = item["tags"]
    address = item["address"]

    # --- Name normalization pipeline ---
    normalized = name.strip()
    # Title case
    words = normalized.split()
    title_words = []
    for w in words:
        if len(w) > 0:
            title_words.append(w[0].upper() + w[1:].lower())
    normalized_name = " ".join(title_words)

    # Slug generation
    slug = ""
    for ch in name.lower():
        if ch.isalnum():
            slug = slug + ch
        elif ch == " " or ch == "-":
            if len(slug) > 0 and slug[-1] != "-":
                slug = slug + "-"
    slug = slug.strip("-")

    # --- SKU generation with category + hash ---
    cat_prefix = category[:3].upper()
    name_hash = 0
    for ch in name:
        name_hash = (name_hash * 37 + ord(ch)) % 1000000
    price_code = int(price * 100) % 10000
    sku = cat_prefix + "-" + str(price_code).zfill(4) + "-" + str(name_hash).zfill(6)

    # --- Price tier classification ---
    if price >= 500:
        tier = "premium"
    elif price >= 100:
        tier = "standard"
    elif price >= 20:
        tier = "budget"
    else:
        tier = "economy"

    # --- Value density ---
    if weight > 0:
        value_density = round(price / weight, 2)
    else:
        value_density = 0.0

    subtotal = price * quantity

    # --- Tag normalization and categorization ---
    normalized_tags = []
    tag_categories = {"promo": [], "info": [], "shipping": [], "other": []}
    promo_keywords = ["sale", "clearance", "discount", "vip", "member", "seasonal"]
    ship_keywords = ["fragile", "hazmat", "flammable", "express", "overnight"]
    for tag in tags:
        t = tag.strip().lower()
        if len(t) == 0:
            continue
        normalized_tags.append(t)
        categorized = False
        for pk in promo_keywords:
            if pk in t:
                tag_categories["promo"].append(t)
                categorized = True
                break
        if not categorized:
            for sk in ship_keywords:
                if sk in t:
                    tag_categories["shipping"].append(t)
                    categorized = True
                    break
        if not categorized:
            # Check if it looks like a descriptor
            if len(t) > 3 and t[0].isalpha():
                tag_categories["info"].append(t)
            else:
                tag_categories["other"].append(t)

    # --- Address normalization ---
    addr_parts = []
    for field in ["street", "city", "state", "zip"]:
        val = address[field]
        if len(str(val)) > 0:
            addr_parts.append(str(val).strip())
    formatted_address = ", ".join(addr_parts)

    # State abbreviation normalization
    state = address["state"].upper().strip()
    state_map = {
        "CALIFORNIA": "CA", "NEW YORK": "NY", "TEXAS": "TX",
        "FLORIDA": "FL", "ILLINOIS": "IL", "OHIO": "OH",
    }
    normalized_state = state_map[state] if state in state_map else state

    # --- Composite scoring ---
    score = 0
    if tier == "premium":
        score = score + 10
    elif tier == "standard":
        score = score + 5
    else:
        score = score + 2
    score = score + len(tag_categories["promo"]) * 3
    if value_density > 100:
        score = score + 5
    elif value_density > 20:
        score = score + 2
    if quantity >= 10:
        score = score + 3
    else:
        score = score + 1

    return {
        "company": "transform_co",
        "original_name": name,
        "normalized_name": normalized_name,
        "slug": slug,
        "sku": sku,
        "tier": tier,
        "value_density": value_density,
        "subtotal": round(subtotal, 2),
        "normalized_tags": normalized_tags,
        "tag_categories": tag_categories,
        "formatted_address": formatted_address,
        "normalized_state": normalized_state,
        "composite_score": score,
    }
