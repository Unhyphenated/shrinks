from pydantic import BaseModel
from pydantic.networks import HttpUrl
from typing import Optional


class LongURLIn(BaseModel):
    long_url: HttpUrl

    # Optional field if you want users to provide a custom short code
    # (Leaving it optional for the 70% plan simplifies creation logic)
    custom_short_code: Optional[str] = None


class ShortURLOut(BaseModel):
    short_code: str

    original_url: HttpUrl

    shortened_link: HttpUrl


# --- Base Schema (For internal data sharing, e.g., with the database model) ---


class URLBase(BaseModel):
    long_url: HttpUrl
    short_code: str
