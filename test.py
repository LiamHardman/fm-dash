from PIL import ImageFont
import logging

logging.basicConfig(filename='font_errors.log', level=logging.DEBUG, filemode='w')

font_path = "arial.ttf"
size = 40

try:
    font = ImageFont.truetype(font_path, size)
    logging.debug('Font loaded successfully.')
except Exception as e:
    logging.exception("Error loading font")