import glob
import time
import os


def delete_old_cards():
    """
    Deletes files older than 30 minutes in the assets folder
    """
    current_time = time.time()

    age_minutes = 30
    age_seconds = age_minutes * 60
    directory = "./assets/"
    patterns = ["fifa_card_*.webp", "club_*.webp"]
    for pattern in patterns:
        for file_path in glob.glob(os.path.join(directory, pattern)):
            creation_time = os.path.getctime(file_path)
            if current_time - creation_time > age_seconds:
                os.remove(file_path)
                print(f"Deleted {file_path}")


def delete_old_cache():
    """
    Deletes files older than 60 days in the fifa_card_assets/player_img folder
    """
    current_time = time.time()
    age_days = 60
    age_seconds = age_days * 5184000
    directory = "./fifa_card_assets/player_img/"
    patterns = ["*.png"]
    for pattern in patterns:
        for file_path in glob.glob(os.path.join(directory, pattern)):
            creation_time = os.path.getctime(file_path)
            if current_time - creation_time > age_seconds:
                os.remove(file_path)
                print(f"Deleted {file_path}")


def delete_old_tmp():
    """
    Deletes files older than 2 hours in the temp folder
    """
    current_time = time.time()
    age_minutes = 120
    age_seconds = age_minutes * 60
    directory = "./tmp/"
    patterns = ["*.json"]
    for pattern in patterns:
        for file_path in glob.glob(os.path.join(directory, pattern)):
            creation_time = os.path.getctime(file_path)
            if current_time - creation_time > age_seconds:
                os.remove(file_path)
                print(f"Deleted {file_path}")


def delete_old_files():
    """
    Deletes files older than 30 minutes in the assets folder
    """
    current_time = time.time()

    age_days = 3
    age_seconds = age_days * 24 * 60 * 60
    directory = "./uploads/"
    patterns = ["*.html"]
    for pattern in patterns:
        for file_path in glob.glob(os.path.join(directory, pattern)):
            creation_time = os.path.getctime(file_path)
            if current_time - creation_time > age_seconds:
                os.remove(file_path)
                print(f"Deleted {file_path}")
