# You will need to import your Base-62 function from app.core.utils
# and the standard Python testing framework (e.g., pytest).

def test_base62_encoding_basic_conversion():
    """
    Tests known inputs to ensure correct Base-62 conversion.
    For example: 
    1. Input ID 0 should result in '0'.
    2. Input ID 61 should result in 'Z' (the last single character).
    3. Input ID 62 should result in '10'.
    """
    # assert base62_encode(0) == "0"
    # assert base62_encode(61) == "Z"
    # assert base62_encode(62) == "10"
    pass

def test_base62_encoding_round_trip():
    """
    Tests encoding a large number and then decoding it back 
    to ensure no data loss during the process.
    """
    # large_id = 987654321
    # encoded = base62_encode(large_id)
    # assert base62_decode(encoded) == large_id # You might need a decode function
    pass

def test_base62_encoding_invalid_input():
    """
    Tests that the function handles edge cases and invalid inputs gracefully.
    For example, ensure negative integers raise an appropriate error.
    """
    # with pytest.raises(ValueError):
    #     base62_encode(-1)
    pass