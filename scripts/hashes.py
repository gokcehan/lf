import os
import hashlib


def calculate_sha256(file_path):
    sha256_hash = hashlib.sha256()
    with open(file_path, "rb") as f:
        # Read and update hash in chunks to save memory
        for byte_block in iter(lambda: f.read(4096), b""):
            sha256_hash.update(byte_block)
    return sha256_hash.hexdigest()


def display_directory_tree(path, indent=0):
    if os.path.isdir(path):
        # Print the folder name
        print("    " * indent + f"<{os.path.basename(path)}>")

        # Recursively explore the subdirectories and files
        for item in os.listdir(path):
            item_path = os.path.join(path, item)
            display_directory_tree(item_path, indent + 1)

    elif os.path.isfile(path):
        # Print the SHA-256 hash and file name
        sha256 = calculate_sha256(path)
        print("    " * indent + f"{sha256}  {os.path.basename(path)}")


if __name__ == "__main__":
    display_directory_tree(os.getcwd())