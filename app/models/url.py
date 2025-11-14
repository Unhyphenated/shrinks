from sqlalchemy import Column, Integer, String
from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


class URLMapping(Base):

    id = Column("id", Integer, primary_key=True, autoincrement=True)

    long_url = Column("long_url", String, unique=True, index=True)

    short_code = Column("short_code", String, unique=True, index=True)

    clicks = Column("clicks", Integer, default=0)

    __tablename__ = "url_mappings"
