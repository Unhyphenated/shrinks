# Skeletons for reading environment variables for PostgreSQL, Redis, etc.
# Note: You'll typically use a library like pydantic-settings here

class Settings:
    # 1. Database Configuration (PostgreSQL)
    DATABASE_URL: str = "your_postgres_connection_string"
    
    # 2. Caching Configuration (Redis)
    REDIS_URL: str = "redis://your_redis_host:port/0"
    
    # 3. Application Configuration
    # This is the domain name the short links will use, crucial for redirects
    BASE_URL: str = "https://shrinks.io/"
    
    # 4. Environment Setting (Useful for toggling behavior, e.g., debug mode)
    ENVIRONMENT: str = "development" # Should be read from OS environment

# Instantiate the settings object that the rest of the application will use
settings = Settings()