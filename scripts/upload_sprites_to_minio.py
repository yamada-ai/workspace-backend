#!/usr/bin/env python3
"""
Upload placeholder sprite sheets to MinIO object storage.

This script:
1. Connects to MinIO
2. Creates 'sprites' bucket if it doesn't exist
3. Uploads all PNG files from sprites_placeholder/ directory
"""

import os
import sys
from pathlib import Path
from minio import Minio
from minio.error import S3Error


# MinIO configuration
MINIO_ENDPOINT = "localhost:9000"
MINIO_ACCESS_KEY = "minioadmin"
MINIO_SECRET_KEY = "minioadmin"
MINIO_SECURE = False  # Set to True if using HTTPS

BUCKET_NAME = "sprites"
SPRITES_DIR = "sprites_placeholder"


def main():
    """Upload all sprite sheets to MinIO."""

    # Initialize MinIO client
    print(f"🔌 Connecting to MinIO at {MINIO_ENDPOINT}...")
    client = Minio(
        endpoint=MINIO_ENDPOINT,
        access_key=MINIO_ACCESS_KEY,
        secret_key=MINIO_SECRET_KEY,
        secure=MINIO_SECURE
    )

    # Create bucket if it doesn't exist
    try:
        if not client.bucket_exists(bucket_name=BUCKET_NAME):
            print(f"📦 Creating bucket '{BUCKET_NAME}'...")
            client.make_bucket(bucket_name=BUCKET_NAME)
            print(f"✅ Bucket '{BUCKET_NAME}' created successfully")
        else:
            print(f"✅ Bucket '{BUCKET_NAME}' already exists")
    except S3Error as err:
        print(f"❌ Error creating bucket: {err}")
        sys.exit(1)

    # Set bucket policy to public read (for easier frontend access)
    # In production, you may want to use signed URLs instead
    policy = f"""{{
        "Version": "2012-10-17",
        "Statement": [
            {{
                "Effect": "Allow",
                "Principal": {{"AWS": ["*"]}},
                "Action": ["s3:GetObject"],
                "Resource": ["arn:aws:s3:::{BUCKET_NAME}/*"]
            }}
        ]
    }}"""

    try:
        client.set_bucket_policy(bucket_name=BUCKET_NAME, policy=policy)
        print(f"✅ Set bucket policy to public read")
    except S3Error as err:
        print(f"⚠️  Warning: Could not set bucket policy: {err}")

    # Upload sprites
    sprites_path = Path(SPRITES_DIR)
    if not sprites_path.exists():
        print(f"❌ Error: Directory '{SPRITES_DIR}' not found")
        print(f"   Please run generate_placeholder_sprites.py first")
        sys.exit(1)

    png_files = list(sprites_path.glob("*.png"))
    if not png_files:
        print(f"❌ Error: No PNG files found in '{SPRITES_DIR}'")
        sys.exit(1)

    print(f"\n📤 Uploading {len(png_files)} sprite sheets...")

    uploaded = 0
    failed = 0

    for png_file in png_files:
        object_name = png_file.name
        file_path = str(png_file)

        try:
            client.fput_object(
                bucket_name=BUCKET_NAME,
                object_name=object_name,
                file_path=file_path,
                content_type="image/png"
            )
            print(f"  ✓ Uploaded: {object_name}")
            uploaded += 1
        except S3Error as err:
            print(f"  ✗ Failed to upload {object_name}: {err}")
            failed += 1

    print(f"\n{'='*60}")
    print(f"✅ Upload complete!")
    print(f"   Uploaded: {uploaded} files")
    if failed > 0:
        print(f"   Failed: {failed} files")
    print(f"   Bucket: {BUCKET_NAME}")
    print(f"   MinIO Console: http://{MINIO_ENDPOINT.split(':')[0]}:9001")
    print(f"   Access URL format: http://{MINIO_ENDPOINT}/{BUCKET_NAME}/{{filename}}")
    print(f"{'='*60}")


if __name__ == '__main__':
    main()
