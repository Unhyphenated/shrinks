# You will need to import necessary components from SQLAlchemy or SQLModel
# e.g., from sqlalchemy import Column, Integer, String, Boolean
#       from sqlalchemy.ext.declarative import declarative_base

# Define the Base class/object for declarative models
# Base = declarative_base() 


class URLMapping: # This class represents the 'url_mappings' table
    
    # The auto-incrementing integer that serves as the basis for the short code
    id: int # Primary Key, autoincrement=True 
    
    # The original long URL. Must be indexed and unique for fast reverse lookups.
    long_url: str # Unique=True, Index=True 
    
    # The generated Base-62 short code (e.g., 'aBc1'). Must be unique.
    short_code: str # Unique=True, Index=True
    
    # (Optional 70% Feature: Tracking clicks is a good future metric)
    clicks: int = 0 
    
    # Define the actual table name in PostgreSQL
    __tablename__ = "url_mappings"