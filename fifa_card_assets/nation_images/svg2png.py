import cairosvg
import os

dir_path = './'

for filename in os.listdir(dir_path):
    if filename.endswith('.svg'):
        svg_file = os.path.join(dir_path, filename)
        png_file = svg_file.replace('.svg', '.png')  # Change file extension to .png
        cairosvg.svg2png(url=svg_file, write_to=png_file)