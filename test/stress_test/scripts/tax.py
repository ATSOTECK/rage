def process_item(item):
    name = item["name"]
    price = item["price"]
    quantity = item["quantity"]
    category = item["category"]
    tags = item["tags"]
    address = item["address"]

    subtotal = price * quantity

    # --- State tax rates lookup ---
    state_rates = {
        "CA": 7.25, "NY": 8.0, "TX": 6.25, "FL": 6.0,
        "IL": 6.25, "OH": 5.75, "WA": 6.5, "PA": 6.0,
        "NJ": 6.625, "MA": 6.25, "VA": 5.3, "GA": 4.0,
        "NC": 4.75, "MI": 6.0, "AZ": 5.6, "CO": 2.9,
    }
    state = address["state"].upper().strip()
    # Normalize full state names
    state_name_map = {
        "CALIFORNIA": "CA", "NEW YORK": "NY", "TEXAS": "TX",
        "FLORIDA": "FL", "ILLINOIS": "IL", "OHIO": "OH",
    }
    if state in state_name_map:
        state = state_name_map[state]
    state_tax_rate = state_rates[state] if state in state_rates else 5.0

    # --- Category-based exemptions ---
    exempt_categories = ["food", "books"]
    # Reduced-rate categories (multiplier on state rate)
    reduced_categories = {"clothing": 0.5, "home": 0.75}

    if category in exempt_categories:
        effective_rate = 0.0
        tax_note = "tax-exempt category"
    elif category in reduced_categories:
        effective_rate = state_tax_rate * reduced_categories[category]
        tax_note = "reduced rate for " + category
    else:
        effective_rate = state_tax_rate
        tax_note = "standard rate for " + state

    # --- Progressive brackets on top of state rate ---
    if subtotal <= 50:
        bracket_mult = 0.8
        bracket = "low"
    elif subtotal <= 200:
        bracket_mult = 1.0
        bracket = "standard"
    elif subtotal <= 1000:
        bracket_mult = 1.1
        bracket = "elevated"
    else:
        bracket_mult = 1.2
        bracket = "high"

    applied_rate = effective_rate * bracket_mult
    tax_amount = subtotal * applied_rate / 100

    # --- Luxury surcharge for items over $500 unit price ---
    luxury_surcharge = 0.0
    if price > 500:
        luxury_surcharge = subtotal * 3.0 / 100

    # --- Special tag-based surcharges ---
    eco_credit = 0.0
    import_duty = 0.0
    for tag in tags:
        t = tag.lower().strip()
        if "eco" in t or "green" in t or "organic" in t:
            eco_credit = eco_credit + subtotal * 0.5 / 100
        if "imported" in t or "foreign" in t:
            import_duty = import_duty + subtotal * 2.0 / 100

    # --- Zone-based duty ---
    zone = address["zone"]
    if zone == "international":
        import_duty = import_duty + subtotal * 5.0 / 100
    elif zone == "regional":
        import_duty = import_duty + subtotal * 1.0 / 100

    total_tax = tax_amount + luxury_surcharge + import_duty - eco_credit
    if total_tax < 0:
        total_tax = 0.0
    total = subtotal + total_tax

    # --- Tax breakdown by component ---
    breakdown = []
    if tax_amount > 0:
        breakdown.append({"type": "state_tax", "rate": round(applied_rate, 4), "amount": round(tax_amount, 2)})
    if luxury_surcharge > 0:
        breakdown.append({"type": "luxury_surcharge", "rate": 3.0, "amount": round(luxury_surcharge, 2)})
    if import_duty > 0:
        breakdown.append({"type": "import_duty", "rate": 0.0, "amount": round(import_duty, 2)})
    if eco_credit > 0:
        breakdown.append({"type": "eco_credit", "rate": -0.5, "amount": round(-eco_credit, 2)})

    # --- Effective tax rate ---
    effective_total_rate = 0.0
    if subtotal > 0:
        effective_total_rate = round(total_tax / subtotal * 100, 4)

    return {
        "company": "tax_co",
        "item_name": name,
        "state": state,
        "subtotal": round(subtotal, 2),
        "state_tax_rate": state_tax_rate,
        "effective_rate": round(applied_rate, 4),
        "bracket": bracket,
        "tax_amount": round(tax_amount, 2),
        "luxury_surcharge": round(luxury_surcharge, 2),
        "import_duty": round(import_duty, 2),
        "eco_credit": round(eco_credit, 2),
        "total_tax": round(total_tax, 2),
        "total": round(total, 2),
        "effective_total_rate": effective_total_rate,
        "tax_note": tax_note,
        "breakdown": breakdown,
    }
