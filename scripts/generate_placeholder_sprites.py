#!/usr/bin/env python3
"""
Generate placeholder sprite sheets for testing animation system.

Format: tier{tierNum}-{iconId}_{motion_key}.png
Example: tier1-01_sleep.png

Each sprite is a horizontal strip: 4 frames × (32px × 32px) = 128px × 32px
"""

import os
from PIL import Image, ImageDraw, ImageFont


# Configuration
TIERS = [1, 2, 3, 4]
ICON_IDS = range(1, 11)  # 01-10
MOTIONS = {
    'sleep': '💤',
    'dance': '💃',
    'happy': '😊'
}

# Animation settings
FRAME_SIZE = 32  # 32px × 32px per frame
FRAME_COUNT = 4
OUTPUT_DIR = 'sprites_placeholder'

# Tier colors (background)
TIER_COLORS = {
    1: (100, 149, 237),  # Cornflower blue
    2: (60, 179, 113),   # Medium sea green
    3: (255, 215, 0),    # Gold
    4: (220, 20, 60)     # Crimson
}


def create_sprite_sheet(tier: int, icon_id: int, motion: str, emoji: str) -> Image.Image:
    """Create a single sprite sheet with 4 frames."""
    width = FRAME_SIZE * FRAME_COUNT
    height = FRAME_SIZE

    # Create image with transparency
    img = Image.new('RGBA', (width, height), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)

    base_color = TIER_COLORS[tier]

    for frame_idx in range(FRAME_COUNT):
        x_offset = frame_idx * FRAME_SIZE

        # Calculate opacity for blinking effect
        opacity = 255 if frame_idx % 2 == 0 else 200
        color = base_color + (opacity,)

        # Draw background rectangle
        draw.rectangle(
            [(x_offset, 0), (x_offset + FRAME_SIZE - 1, FRAME_SIZE - 1)],
            fill=color,
            outline=(255, 255, 255, 255),
            width=1
        )

        # Draw tier and icon ID text
        try:
            # Try to use a decent font
            font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 10)
        except:
            # Fallback to default
            font = ImageFont.load_default()

        label = f"T{tier}-{icon_id:02d}"

        # Calculate text position (top)
        bbox = draw.textbbox((0, 0), label, font=font)
        text_width = bbox[2] - bbox[0]
        text_x = x_offset + (FRAME_SIZE - text_width) // 2
        text_y = 2

        # Draw text with shadow for readability
        draw.text((text_x + 1, text_y + 1), label, fill=(0, 0, 0, 200), font=font)
        draw.text((text_x, text_y), label, fill=(255, 255, 255, 255), font=font)

        # Draw motion text (bottom)
        motion_label = f"{motion[0].upper()}"  # First letter of motion

        bbox = draw.textbbox((0, 0), motion_label, font=font)
        text_width = bbox[2] - bbox[0]
        text_x = x_offset + (FRAME_SIZE - text_width) // 2
        text_y = FRAME_SIZE - 14

        draw.text((text_x + 1, text_y + 1), motion_label, fill=(0, 0, 0, 200), font=font)
        draw.text((text_x, text_y), motion_label, fill=(255, 255, 255, 255), font=font)

    return img


def generate_all_sprites():
    """Generate all placeholder sprite sheets."""
    # Create output directory
    os.makedirs(OUTPUT_DIR, exist_ok=True)

    total = 0

    for tier in TIERS:
        for icon_id in ICON_IDS:
            for motion, emoji in MOTIONS.items():
                filename = f"tier{tier}-{icon_id:02d}_{motion}.png"
                filepath = os.path.join(OUTPUT_DIR, filename)

                # Create sprite sheet
                img = create_sprite_sheet(tier, icon_id, motion, emoji)

                # Save
                img.save(filepath, 'PNG')
                total += 1

                print(f"✓ Generated: {filename}")

    print(f"\n✅ Successfully generated {total} sprite sheets in '{OUTPUT_DIR}/'")
    print(f"   Tiers: {len(TIERS)}")
    print(f"   Icons per tier: {len(list(ICON_IDS))}")
    print(f"   Motions: {len(MOTIONS)}")
    print(f"   Total: {total} files")


if __name__ == '__main__':
    generate_all_sprites()
