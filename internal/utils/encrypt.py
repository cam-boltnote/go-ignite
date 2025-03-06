import os
import base64

# Generate 32 random bytes
raw_key = os.urandom(32)

# Base64 encode it for storage in .env file
encoded_key = base64.b64encode(raw_key)

print("Raw key length:", len(raw_key), "bytes")
print("Base64 encoded key:", encoded_key.decode())