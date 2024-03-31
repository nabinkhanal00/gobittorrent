import bencodepy  # Install using `pip install bencodepy`

# Read and parse the .torrent file
with open("sample.torrent", "rb") as f:
    metadata = bencodepy.decode(f.read())

# Extract piece hashes
print(len(metadata[b"info"][b"pieces"]))
piece_hashes = [metadata[b"info"][b"pieces"][i:i+20] for i in range(0, len(metadata[b"info"][b"pieces"]), 20)]

# Convert each hash to hexadecimal format
piece_hashes_hex = [piece_hash.hex() for piece_hash in piece_hashes]

# Print the piece hashes
for i, hash_hex in enumerate(piece_hashes_hex, start=1):
    print(f"Piece {i} hash: {hash_hex}")
