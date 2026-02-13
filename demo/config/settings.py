# Server settings — environment-aware, computed configuration.
#
# In TOML/JSON/YAML you'd need separate files per environment
# and couldn't compute derived values. Here it's just Python.

environment = env("RAGE_ENV", "development")

# --- Per-environment overrides (try: RAGE_ENV=production) ---

if environment == "production":
    host = "0.0.0.0"
    port = 8443
    debug = False
    log_level = 2
    db_host = "db.prod.internal"
    db_pool_size = cpu_count * 4
elif environment == "staging":
    host = "0.0.0.0"
    port = 8080
    debug = True
    log_level = 3
    db_host = "db.staging.internal"
    db_pool_size = cpu_count * 2
else:
    host = "127.0.0.1"
    port = 5000
    debug = True
    log_level = 4
    db_host = "localhost"
    db_pool_size = 2

# --- Computed / derived values ---

max_connections = db_pool_size * 10
api_url = f"https://{host}:{port}/api/v1"
db_password = env("DB_PASSWORD", "dev_pass")
db_connection = f"postgres://app:{db_password}@{db_host}:5432/gamedb"

# --- Validation (static configs can't validate themselves) ---

assert 1 <= log_level <= 5, "log_level must be between 1 and 5"
assert db_pool_size > 0, "db_pool_size must be positive"

# --- Feature flags with conditional logic ---

features = {
    "websockets": True,
    "rate_limiting": environment == "production",
    "debug_toolbar": debug,
    "metrics": environment != "development",
    "ssl": port == 8443,
}

# --- Final export ---

settings = {
    "environment": environment,
    "host": host,
    "port": port,
    "debug": debug,
    "log_level": log_level,
    "api_url": api_url,
    "db_host": db_host,
    "db_pool_size": db_pool_size,
    "db_connection": db_connection,
    "max_connections": max_connections,
    "features": features,
}

log(f"  Settings loaded for '{environment}' environment")
