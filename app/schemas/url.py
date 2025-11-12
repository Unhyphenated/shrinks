# You will need to import Pydantic's BaseModel and other necessary types (e.g., HttpUrl)
# from pydantic import BaseModel
# from pydantic.networks import HttpUrl
from typing import Optional

# --- Input Schema (What the Client Sends to POST /shorten) ---

class LongURLIn: 
    # The URL that the user wants to shorten.
    # Should use Pydantic validation (HttpUrl) to ensure it's a real URL format.
    long_url: str
    
    # Optional field if you want users to provide a custom short code
    # (Leaving it optional for the 70% plan simplifies creation logic)
    custom_short_code: Optional[str] = None

# --- Output Schemas (What the API Returns) ---

class ShortURLOut: 
    # The successful response after shortening a URL.
    short_code: str
    
    # Returning the original URL confirms the mapping.
    original_url: str 
    
    # Add the full link the user can copy
    shortened_link: str 

# --- Base Schema (For internal data sharing, e.g., with the database model) ---

class URLBase:
    # Fields common to both input and output
    long_url: str
    short_code: str