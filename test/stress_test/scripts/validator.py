def process_item(item):
    errors = []
    warnings = []
    fixes = []

    name = item["name"]
    price = item["price"]
    quantity = item["quantity"]
    category = item["category"]
    weight = item["weight"]
    tags = item["tags"]
    address = item["address"]

    valid_categories = ["electronics", "clothing", "food", "books", "toys", "home", "sports"]

    # --- Name validation with sanitization ---
    if len(name) == 0:
        errors.append("name is required")
    elif len(name) > 100:
        errors.append("name must be 100 characters or less")

    # Check for suspicious characters
    clean_name = ""
    for ch in name:
        if ch.isalnum() or ch in " -_.":
            clean_name = clean_name + ch
        else:
            fixes.append("removed invalid char from name: " + ch)
    if clean_name != name:
        warnings.append("name contained invalid characters")

    # Check for profanity (simple blocklist iteration)
    blocked = ["badword1", "badword2", "offensive", "banned"]
    name_lower = name.lower()
    for word in blocked:
        if word in name_lower:
            errors.append("name contains blocked content")
            break

    # --- Price validation ---
    if price <= 0:
        errors.append("price must be positive")
    elif price > 99999:
        errors.append("price exceeds maximum allowed")
    if price > 1000:
        warnings.append("high-value item detected")
    if price != round(price, 2):
        fixes.append("price rounded to 2 decimal places")

    # --- Quantity validation ---
    if quantity <= 0:
        errors.append("quantity must be positive")
    elif quantity > 1000:
        warnings.append("bulk order detected")

    # --- Category validation ---
    if category not in valid_categories:
        # Fuzzy match: check prefix
        matched = False
        for vc in valid_categories:
            if len(category) >= 3 and vc.startswith(category[:3]):
                fixes.append("category corrected: " + category + " -> " + vc)
                matched = True
                break
        if not matched:
            errors.append("invalid category: " + category)

    # --- Weight validation ---
    if weight <= 0:
        errors.append("weight must be positive")
    elif weight > 100:
        warnings.append("heavy item detected")

    # --- Tag validation ---
    valid_tag_count = 0
    invalid_tags = []
    seen_tags = {}
    for tag in tags:
        stripped = tag.strip().lower()
        if len(stripped) == 0:
            invalid_tags.append(tag)
        elif stripped in seen_tags:
            warnings.append("duplicate tag: " + stripped)
        else:
            seen_tags[stripped] = True
            valid_tag_count = valid_tag_count + 1
    if len(invalid_tags) > 0:
        warnings.append("found " + str(len(invalid_tags)) + " empty/invalid tags")

    # --- Address validation ---
    required_fields = ["street", "city", "state", "zip", "zone"]
    missing_fields = []
    for field in required_fields:
        if field not in address or len(str(address[field])) == 0:
            missing_fields.append(field)
    if len(missing_fields) > 0:
        errors.append("address missing fields: " + ", ".join(missing_fields))

    # Zip code format check
    zip_code = address["zip"]
    if len(zip_code) > 0:
        digit_count = 0
        for ch in zip_code:
            if ch >= "0" and ch <= "9":
                digit_count = digit_count + 1
        if digit_count != 5:
            warnings.append("zip code may be invalid: " + zip_code)

    # --- Cross-field validation ---
    subtotal = price * quantity
    if subtotal > 10000 and category == "toys":
        warnings.append("unusually large toy order")
    if weight > 50 and category == "clothing":
        warnings.append("suspiciously heavy clothing item")

    # Compute risk score
    risk_score = len(errors) * 10 + len(warnings) * 3 + len(fixes)
    risk_level = "low"
    if risk_score >= 20:
        risk_level = "high"
    elif risk_score >= 10:
        risk_level = "medium"

    return {
        "company": "validator_co",
        "item_name": name,
        "clean_name": clean_name,
        "valid": len(errors) == 0,
        "errors": errors,
        "warnings": warnings,
        "fixes": fixes,
        "error_count": len(errors),
        "warning_count": len(warnings),
        "fix_count": len(fixes),
        "valid_tag_count": valid_tag_count,
        "risk_score": risk_score,
        "risk_level": risk_level,
    }
