from typing import Generator

from app.core.config import Settings
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from redis import Redis

settings = Settings()

engine = create_engine(settings.DATABASE_URL)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

def get_db_session() -> Generator:
    try:
        db = SessionLocal()
        yield db

    finally:
        db.close()

# Global Redis client instance
redis_client: Redis = None

def get_redis_client() -> Redis:
    """
    Returns the initialized global Redis client instance.
    """
    global redis_client
    
    if redis_client is None:
        redis_client = Redis(
            host='redis-16264.c10.us-east-1-2.ec2.cloud.redislabs.com',
            port=16264,
            decode_responses=True,
            username="default",
            password="jKnj4aSNUH62Swx1aHpmQFiQsKHt3RLd",
        )
    return redis_client