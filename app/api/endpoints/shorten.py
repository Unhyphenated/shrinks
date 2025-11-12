def shorten_url_endpoint(url_in):
    """
    Handles the creation (write) request:
    1. Check for existing long URL.
    2. If new, insert into DB, get ID.
    3. Encode ID to short code.
    4. Return short code.
    """
    pass

def redirect_url_endpoint(short_code: str):
    """
    Handles the redirection (read) request using caching:
    1. Check Redis for short code (Cache-Aside Pattern).
    2. If found, issue HTTP 302 Redirect.
    3. If miss, query PostgreSQL.
    4. If found in DB, save to Redis and issue HTTP 302 Redirect.
    5. If not found anywhere, raise 404.
    """
    pass

# Note: The actual APIRouter object and dependency injections for DB/Redis 
# will be defined here, but are omitted per instructions.