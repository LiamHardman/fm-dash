from PIL import Image, ImageDraw, ImageFont
from itertools import chain
import os

"""
This script is used in isolation to create lots of playstyle images,
based on the color_map provided. Each one will go to a folder with the same name in fifa_card_assets.
If you're on a slower PC, lower the scale_factor to 4 or 2. 1 is not recommended,
and anything above 16 is diminishing returns.
"""
color_map = {
    "normal": ((64,52,29,255), (226,192,117)),
    "icon": ((89, 76, 43, 255), (219, 219, 219, 255)),
    "totw": ((255, 226, 139, 255), (22, 22, 21, 255)),
    "motm": ((242,242,242, 255), (10,28,62, 255)),
    "hero": ((255,255,255, 255),  (216, 100, 0,255)),
    "ucl_hero": ((255,255,255, 255),  (5, 94, 162,255)),
    "toty": ((255, 222, 128,255), (19, 35, 85,255)),
    "if_silver": ((175,188,200, 255), (24,24,24, 255)),
    "if_bronze": ((226,164,135, 255), (20,20,19, 255)),
    "potm_epl":((250,250,250,255), (77,29,209,255)),
    "potm_bundesliga":((245,245,245,255), (25,25,25,255)),
    "potm_ligue1":((242,242,242,255), (10,28,62,255)),
    "potm_laliga":((242,242,242,255), (10,28,62,255)),
    "potm_serie_a":((0,232,217,255), (28,18,89,255)),
    "potm_mls":((242,242,242,255), (10,28,62,255)),
    "potm_eredivisie":((242,242,242,255), (54,96,58,255)),
    "totgs_ucl":((255,255,255,255), (4,18,90,255)),
    "totgs_uel":((255,255,255,255), (0,0,0,255)),
    "totgs_uecl":((255,255,255,255), (0,0,0,255)),
    "rttk_ucl":((255,255,255,255), (4,18,90,255)),
    "rttk_uel":((255,255,255,255), (0,0,0,255)),
    "rttk_uecl":((255,255,255,255), (0,0,0,255)),
    "fut_champs":((246,219,122,255), (103,23,16,255)),


}

name_map = {
    66: "first_touch",
    67: "flair",
    68: "press_proven",
    69: "rapid",
    70: "technical",
    71: "trickster",
    72: "anticipate",
    73: "block",
    74: "bruiser",
    75: "intercept",
    76: "jockey",
    77: "slide_tackle",
    78: "shield",
    79: "save",
    80: "catch",
    81: "footwork",
    82: "1on1",
    83: "rushout",
    84: "pinged_pass",
    85: "long_ball_pass",
    86: "incisive_pass",
    87: "tiki_taka",
    88: "whipped_pass",
    89: "acrobatic",
    90: "aerial",
    97: "giant_throw",
    98: "quick_step",
    99: "relentless",
    100: "trivela",
    101: "catch",
    102: "dead_ball",
    103: "finesse",
    104: "power_header",
    105: "power_shot"
}

base_width, base_height = 121,121

# Define scale factor for supersampling (e.g., 4x higher resolution)
scale_factor = 16
width, height = base_width * scale_factor + 70, base_height * scale_factor + 20

# Adjust the font size, border width, and vertical offset for the higher resolution
font_size_high_res = 96 * scale_factor
smaller_font_size = 64 * scale_factor  # Adjust as needed
border_width_high_res = 8 * scale_factor
vertical_offset_high_res = 10 * scale_factor

font_path = './fifa_card_assets/playstyles.ttf'
pil_font = ImageFont.truetype(font_path, font_size_high_res)

# Define the text and polygon fill color
text_and_border_color = (89, 76, 43, 255)  # Brownish color
polygon_color = (219, 219, 219, 255)  # Light grey color

# Define the scaled polygon points and center based on the new dimensions
scaled_polygon_points = [
    (25 * width / 100, 0),                        # Left top corner
    (75 * width / 100, 0),                        # Right top corner, adjusted y-coordinate
    (100 * width / 100, 40 * height / 100),       # Right side
    (50 * width / 100, 100 * height / 100),       # Bottom
    (0, 40 * height / 100)                        # Left side
]

# Calculate the scaled center of the polygon
scaled_polygon_center = (
    sum([point[0] for point in scaled_polygon_points]) / len(scaled_polygon_points),
    sum([point[1] for point in scaled_polygon_points]) / len(scaled_polygon_points)
)

for category, (text_and_border_color, polygon_color) in color_map.items():
    directory = f'./fifa_card_assets/playstyles/{category}_playstyles'
    os.makedirs(directory, exist_ok=True)

    for codepoint in chain(range(66, 91), range(97, 106)):
        char = chr(codepoint)
        name = name_map.get(codepoint, f"unknown_{codepoint}")  # Default to 'unknown' if codepoint not in map

        # Determine the font size based on the codepoint
        if codepoint in [66, 67, 68]:
            adjusted_font_size = smaller_font_size
        else:
            adjusted_font_size = font_size_high_res

        # Create the font object with the determined size
        pil_font = ImageFont.truetype(font_path, adjusted_font_size)
        img = Image.new('RGBA', (width, height))
        draw = ImageDraw.Draw(img)

        # Draw the polygon and border at high resolution
        draw.polygon(scaled_polygon_points, fill=polygon_color, outline=text_and_border_color, width=border_width_high_res)

        # Measure the size of the text at high resolution
        text_width, text_height = draw.textsize(char, font=pil_font)

        # Calculate the position for the text at high resolution
        text_position_high_res = (
            scaled_polygon_center[0] - text_width / 2,
            scaled_polygon_center[1] - text_height / 2 + vertical_offset_high_res  # Adjust vertical offset for high resolution
        )

        # Draw the text at high resolution
        draw.text(text_position_high_res, char, font=pil_font, fill=text_and_border_color)

        # Downsample the image back to the original size
        img = img.resize((base_width, base_height), resample=Image.Resampling.LANCZOS)
        directory = f'./fifa_card_assets/playstyles/{category}_playstyles'
        img.save(f'{directory}/{name}.png')