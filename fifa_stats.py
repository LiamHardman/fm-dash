from dash.dependencies import Input, Output, State
from PIL import Image, ImageDraw, ImageFont
from dash import no_update, callback_context
import pandas as pd
import math
from io import BytesIO
import random
import uuid
import os
import time
from bs4 import BeautifulSoup
import requests
from functools import lru_cache
from concurrent.futures import ThreadPoolExecutor, as_completed
from defaults import *
from cachetools import TTLCache
from ratelimit import limits, sleep_and_retry
from difflib import SequenceMatcher
from tracer_provider import tracer
import app_config
from opentelemetry import trace, context
from prometheus_client import Counter, Histogram

# Define counters
sofifa_api_requests_success = Counter(
    "sofifa_api_requests_success", "Total number of successful SOFIFA API requests"
)
sofifa_api_requests_failure = Counter(
    "sofifa_api_requests_failure", "Total number of failed SOFIFA API requests"
)
sofifa_api_requests_timeout = Counter(
    "sofifa_api_requests_timeout", "Total number of timed out SOFIFA API requests"
)
sofifa_api_requests_error = Counter(
    "sofifa_api_requests_error", "Total number of SOFIFA API requests with errors"
)
sofifa_api_request_duration = Histogram(
    "sofifa_api_request_duration", "Duration of SOFIFA API requests in seconds"
)
sortitoutsi_requests_success = Counter(
    "sortitoutsi_requests_success", "Total number of successful SortitoutSI requests"
)

sortitoutsi_requests_failure = Counter(
    "sortitoutsi_requests_failure", "Total number of failed SortitoutSI requests"
)

sortitoutsi_requests_cache_hit = Counter(
    "sortitoutsi_requests_cache_hit",
    "Total number of cache hits for SortitoutSI requests",
)

# Define histogram for tracking the duration of requests
sortitoutsi_request_duration = Histogram(
    "sortitoutsi_request_duration", "Duration of SortitoutSI requests in seconds"
)

yaml_config = app_config.load_config()
app_config.setup_logging(yaml_config["log_level"])
fallback_image_url = yaml_config["fallback_image_url"]

player_image_cache = TTLCache(maxsize=1000, ttl=6000)
cache = TTLCache(maxsize=500, ttl=6000)

max_calls_num = 40
max_calls_frequency = 1  # per second
potm_card_types = [
    "potm_epl",
    "potm_bundesliga",
    "potm_laliga",
    "potm_ligue1",
    "potm_serie_a",
    "potm_mls",
    "potm_eredivisie",
    "potm_laliga",
]


@lru_cache(maxsize=None)
def load_font(path, size):
    return ImageFont.truetype(path, size)


font_paths = {
    "bold": "./fifa_card_assets/bold.otf",
    "medium": "./fifa_card_assets/medium.otf",
}


def get_font(name, size):
    return load_font(font_paths[name], size)


def camel_case(s):
    return "".join(word.capitalize() for word in s.split())


def ensure_directories_exist():
    os.makedirs("league_data/team_img", exist_ok=True)
    os.makedirs("league_data/league_img", exist_ok=True)
    os.makedirs("league_data/temp", exist_ok=True)


def get_first_team_url(search_term):
    with tracer.start_as_current_span("get_first_team_url") as span:
        start_time = time.time()
        cached_url = cache.get(search_term)

        if cached_url:
            sortitoutsi_requests_cache_hit.inc()
            return cached_url

        base_url = "https://sortitoutsi.net/search/database"
        params = {"search": search_term, "type": "team"}
        response = requests.get(base_url, params=params)

        duration = time.time() - start_time
        sortitoutsi_request_duration.observe(duration)
        span.set_attribute("request.duration", duration)
        if response.status_code == 200:
            sortitoutsi_requests_success.inc()
            soup = BeautifulSoup(response.text, "html.parser")
            span.set_attribute("search_term", search_term)
            span.set_attribute("response.status_code", response.status_code)
            span.add_event("Cache hit" if cached_url else "Cache miss")
            team_links = soup.find_all("a", class_="item-title")
            best_match = None
            highest_ratio = 0.0

            for link in team_links:
                link_text = link.get_text(strip=True).lower()
                match_ratio = SequenceMatcher(
                    None, search_term.lower(), link_text
                ).ratio()
                if match_ratio > highest_ratio:
                    highest_ratio = match_ratio
                    best_match = link["href"].replace(
                        "football-manager", "football-manager-2024"
                    )

                if match_ratio == 1:
                    break

            if best_match:
                cache[search_term] = best_match
                return best_match
            else:
                return "No team found"
        else:
            sortitoutsi_requests_failure.inc()
            return "Failed to retrieve page"


def get_image_urls_and_league(team_page_url):
    """
    Retrieves the image URLs and league name associated with a team's page URL.

    Args:
        team_page_url (str): The URL of the team's page.

    Returns:
        tuple: A tuple containing the competition image URL, team image URL, and league name.
               If any of the values are not found, they will be set to None.
    """
    cached_data = cache.get(team_page_url)
    if cached_data:
        return cached_data
    response = requests.get(team_page_url)
    if response.status_code == 200:
        soup = BeautifulSoup(response.text, "html.parser")
        img_tags = soup.find_all("img")
        comp_img = next((img["src"] for img in img_tags if "comp" in img["src"]), None)
        team_img = next(
            (
                img["src"]
                for img in img_tags
                if "team/" in img["src"] and not "team_sm/" in img["src"]
            ),
            None,
        )
        league_link = soup.find("a", href=lambda x: x and "competition" in x)
        league_name = league_link.get_text(strip=True) if league_link else None

        cache[team_page_url] = (comp_img, team_img, league_name)
        return comp_img, team_img, league_name
    else:
        return None, None, None


def integrate_scraping(image, selected_row, render_club_images_values):
    ensure_directories_exist()
    club_name = selected_row.get("Club", "").strip()

    team_url = get_first_team_url(club_name)
    if "http" in team_url:
        comp_img_url, team_img_url, league_name = get_image_urls_and_league(team_url)
        if team_img_url and render_club_images_values:
            team_img_path = save_fm_image(
                team_img_url,
                "league_data/team_img",
                f"{club_name.replace(' ', '')}.png",
            )
            if team_img_path:
                try:
                    club_img = Image.open(team_img_path).convert("RGBA")
                    aspect_ratio = club_img.width / club_img.height
                    new_height = 48
                    new_width = int(new_height * aspect_ratio)
                    club_img = club_img.resize((new_width, new_height), Image.ANTIALIAS)
                    club_x = (image.width - new_width) // 2 + 60
                    club_y = image.height - new_height - 110
                    image.paste(club_img, (club_x, club_y), club_img)
                except FileNotFoundError:
                    print(f"Club image for {club_name} not found at {team_img_path}")

    else:
        print(f"URL not found for {club_name}")

    return image


def save_fm_image(image_url, directory, filename):
    cached_path = cache.get(image_url)
    if cached_path:
        return cached_path
    response = requests.get(image_url, timeout=1)
    if response.status_code == 200:
        image = Image.open(BytesIO(response.content))
        path = os.path.join(directory, filename)
        image.save(path)
        cache[image_url] = path
        return path
    else:
        print(f"Failed to retrieve image at {image_url}")
        return None


def integrate_league_image(image, selected_row):
    ensure_directories_exist()
    club_name = selected_row.get("Club", "").strip()
    max_width = 64
    max_height = 48
    team_url = get_first_team_url(club_name)
    if "http" in team_url:
        comp_img_url, team_img_url, league_name = get_image_urls_and_league(team_url)
        if comp_img_url and league_name:
            league_filename = f"{league_name.replace(' ', '_')}.png"
            league_img_path = save_fm_image(
                comp_img_url, "league_data/league_img", league_filename
            )

            if league_img_path:
                try:
                    league_img = Image.open(league_img_path).convert("RGBA")
                    orig_width, orig_height = league_img.size
                    aspect_ratio = orig_width / orig_height
                    if orig_width > orig_height:
                        new_width = min(orig_width, max_width)
                        new_height = int(new_width / aspect_ratio)
                    else:
                        new_height = min(orig_height, max_height)
                        new_width = int(new_height * aspect_ratio)

                    league_img = league_img.resize(
                        (new_width, new_height), Image.ANTIALIAS
                    )
                    league_x = (image.width - new_width) // 2
                    league_y = image.height - new_height - 110
                    image.paste(league_img, (league_x, league_y), league_img)
                except FileNotFoundError:
                    print(f"League image not found at {league_img_path}")

    else:
        print(f"URL not found for {club_name}")

    return image


@sleep_and_retry
@limits(calls=max_calls_num, period=max_calls_frequency)
def get_or_download_player_image(player_name, base_path="fifa_card_assets/player_img"):
    """
    Retrieves or downloads the image of a player.

    Args:
        player_name (str): The name of the player.
        base_path (str, optional): The base path where the player images are stored. Defaults to "fifa_card_assets/player_img".

    Returns:
        str: The file path of the player image if found or downloaded, None otherwise.
    """
    name_camel = camel_case(player_name)
    file_path = os.path.join(base_path, f"{name_camel}.png")
    if os.path.exists(file_path):
        print(f"Using cached image for {player_name}")
        return file_path
    player_url = get_player_url(player_name)
    if player_url:
        fetch_and_save_player_image(player_url, file_path)
        return file_path
    else:
        print(f"Player image for {player_name} could not be found.")
        return None


def save_image(image_url, file_path="player.png"):
    response = requests.get(image_url, timeout=1)
    if response.status_code == 200:
        with open(file_path, "wb") as file:
            file.write(response.content)
        print(f"Image saved as {file_path}")
    else:
        print("Failed to retrieve image:", response.status_code)


def fetch_and_save_player_image(player_page_url, file_path):
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
    }
    response = requests.get(player_page_url, headers=headers, timeout=1)

    if response.status_code == 200:
        soup = BeautifulSoup(response.content, "html.parser")
        img_tag = soup.find("img", {"data-type": "player"})

        if img_tag and "data-srcset" in img_tag.attrs:
            image_url = img_tag["data-srcset"].split(", ")[-1].split(" ")[0]
            save_image(image_url, file_path)
        else:
            print("Player image not found.")
    else:
        print("Failed to retrieve player page:", response.status_code)


def filter_players_by_author(data, excluded_authors):
    print("Data received for filtering: ", data)

    return [
        player
        for player in data
        if "author" in player
        and not any(ex_author in player["author"] for ex_author in excluded_authors)
    ]


@sleep_and_retry
@limits(calls=max_calls_num, period=max_calls_frequency)
def get_player_url(player_name):
    query = player_name.replace(" ", "%20")
    api_url = f"https://sofifa.com/api/player/suggestion?gender=0&hl=en-US&term={query}"
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
    }
    start_time = time.time()
    span = None

    try:
        response = requests.get(api_url, headers=headers, timeout=1)
        current_ctx = context.get_current()

        with tracer.start_as_current_span(
            "get_player_url", context=current_ctx
        ) as span:
            if response.status_code == 200:
                data = response.json()
                excluded_authors = ["07", "08", "09", "10", "11", "12"]
                filtered_data = filter_players_by_author(data, excluded_authors)

                if filtered_data:
                    player_id = filtered_data[0]["id"]
                    player_url = f"https://sofifa.com/player/{player_id}"
                    sofifa_api_requests_success.inc()
                    return player_url

            else:
                sofifa_api_requests_failure.inc()
                print("Failed to retrieve data:", response.status_code)
                return None

    except requests.Timeout:
        sofifa_api_requests_timeout.inc()
        print("Request timed out")
        return None
    except requests.RequestException as e:
        sofifa_api_requests_error.inc()
        print("Request failed:", e)
        return None
    finally:
        duration = time.time() - start_time
        sofifa_api_request_duration.observe(duration)

        if span is not None:
            span.set_attribute("request.duration", duration)
            span.end()

    return None


def generate_formation_options():
    return [{"label": formation, "value": formation} for formation in FORMATIONS.keys()]


dropdown_options = generate_formation_options()


def get_base_image_path(overall_score, card_type):
    try:
        overall_score = int(overall_score)
    except ValueError:
        print("Overall score is not a valid number")
        return None

    if card_type == "totw":
        if overall_score >= 75:
            return "./fifa_card_assets/cards/if_gold.png"
        elif 65 <= overall_score < 75:
            return "./fifa_card_assets/cards/if_silver.png"
        else:
            return "./fifa_card_assets/cards/if_bronze.png"
    elif card_type == "icon":
        return "./fifa_card_assets/cards/icon.png"
    elif card_type == "hero":
        return "./fifa_card_assets/cards/hero.png"
    elif card_type == "ucl_hero":
        return "./fifa_card_assets/cards/ucl_hero.png"
    elif card_type == "toty":
        return "./fifa_card_assets/cards/toty.png"
    elif card_type == "totgs_ucl":
        return "./fifa_card_assets/cards/totgs_ucl.png"
    elif card_type == "totgs_uel":
        return "./fifa_card_assets/cards/totgs_uel.png"
    elif card_type == "totgs_uecl":
        return "./fifa_card_assets/cards/totgs_uecl.png"
    elif card_type == "fut_champs":
        return "./fifa_card_assets/cards/fut_champs.png"
    elif card_type == "rttk_ucl":
        return "./fifa_card_assets/cards/rttk_ucl.png"
    elif card_type == "rttk_uel":
        return "./fifa_card_assets/cards/rttk_uel.png"
    elif card_type == "rttk_uecl":
        return "./fifa_card_assets/cards/rttk_uecl.png"
    elif card_type == "motm":
        return "./fifa_card_assets/cards/motm.png"
    elif card_type == "potm_epl":
        return "./fifa_card_assets/cards/potm_epl.png"
    elif card_type == "potm_bundesliga":
        return "./fifa_card_assets/cards/potm_bundesliga.png"
    elif card_type == "potm_ligue1":
        return "./fifa_card_assets/cards/potm_ligue1.png"
    elif card_type == "potm_serie_a":
        return "./fifa_card_assets/cards/potm_serie_a.png"
    elif card_type == "potm_laliga":
        return "./fifa_card_assets/cards/potm_laliga.png"
    elif card_type == "potm_mls":
        return "./fifa_card_assets/cards/potm_mls.png"
    elif card_type == "potm_eredivisie":
        return "./fifa_card_assets/cards/potm_eredivisie.png"
    elif overall_score >= 75:
        return "./fifa_card_assets/cards/1_gold.png"
    elif 65 <= overall_score < 75:
        return "./fifa_card_assets/cards/1_silver.png"
    else:
        return "./fifa_card_assets/cards/1_bronze.png"


def calculate_icon_enhancement(current_value):
    if current_value < 25:
        percent_increase = random.uniform(80, 120)
    elif 25 <= current_value < 40:
        percent_increase = random.uniform(55, 77)
    elif 40 <= current_value < 50:
        percent_increase = random.uniform(45, 60)
    elif 50 <= current_value < 60:
        percent_increase = random.uniform(30, 50)
    elif 60 <= current_value < 70:
        percent_increase = random.uniform(13, 20)
    elif 70 <= current_value < 75:
        percent_increase = random.uniform(12, 24)
    elif 75 <= current_value < 85:
        percent_increase = random.uniform(7, 12)
    elif 85 <= current_value < 90:
        percent_increase = random.uniform(6, 10)
    elif 90 <= current_value < 95:
        percent_increase = random.uniform(5, 8)
    else:
        percent_increase = random.uniform(4, 7)
    enhancement = math.ceil(current_value * (percent_increase / 100.0))

    # Debugging prints
    print(
        f"Current Value: {current_value}, Percent Increase: {percent_increase}, Enhancement: {enhancement}"
    )

    return enhancement


def calculate_icon_individual_stats(current_value):
    if current_value < 25:
        percent_increase = random.uniform(80, 120)
    elif 25 <= current_value < 40:
        percent_increase = random.uniform(50, 70)
    elif 40 <= current_value < 50:
        percent_increase = random.uniform(35, 50)
    elif 50 <= current_value < 60:
        percent_increase = random.uniform(28, 48)
    elif 60 <= current_value < 70:
        percent_increase = random.uniform(12, 23)
    elif 70 <= current_value < 75:
        percent_increase = random.uniform(14, 18)
    elif 75 <= current_value < 85:
        percent_increase = random.uniform(11, 16)
    elif 85 <= current_value < 90:
        percent_increase = random.uniform(7, 11)
    elif 90 <= current_value < 95:
        percent_increase = random.uniform(5, 7)
    else:
        percent_increase = random.uniform(2, 5)
    enhancement = math.ceil(current_value * (percent_increase / 100.0))
    return enhancement


def calculate_hero_enhancement(overall_rating):
    if overall_rating <= 45:
        return random.randint(30, 40)
    elif 45 < overall_rating <= 50:
        return random.randint(25, 35)
    elif 50 < overall_rating <= 55:
        return random.randint(19, 25)
    elif 55 < overall_rating <= 60:
        return random.randint(17, 24)
    elif 60 < overall_rating <= 65:
        return random.randint(16, 22)
    elif 65 < overall_rating <= 70:
        return random.randint(11, 16)
    elif 70 < overall_rating <= 75:
        return random.randint(7, 11)
    elif 75 < overall_rating <= 80:
        return random.randint(4, 9)
    elif 80 < overall_rating <= 85:
        return random.randint(3, 7)
    elif 85 < overall_rating <= 90:
        return random.randint(2, 5)
    elif 90 < overall_rating <= 95:
        return random.randint(1, 3)
    else:
        return random.randint(1, 2)


def calculate_hero_individual_stats(current_value):
    if current_value < 25:
        percent_increase = random.uniform(70, 110)
    elif 25 <= current_value < 40:
        percent_increase = random.uniform(50, 67)
    elif 40 <= current_value < 50:
        percent_increase = random.uniform(40, 55)
    elif 50 <= current_value < 60:
        percent_increase = random.uniform(25, 46)
    elif 60 <= current_value < 70:
        percent_increase = random.uniform(6, 13)
    elif 70 <= current_value < 75:
        percent_increase = random.uniform(6, 11)
    elif 75 <= current_value < 85:
        percent_increase = random.uniform(5, 8)
    elif 85 <= current_value < 90:
        percent_increase = random.uniform(4, 7)
    elif 90 <= current_value < 95:
        percent_increase = random.uniform(3, 5)
    else:
        percent_increase = random.uniform(2, 4)

    enhancement = math.ceil(current_value * (percent_increase / 100.0))

    return enhancement


def calculate_toty_individual_stats(current_value):
    if current_value < 25:
        percent_increase = random.uniform(90, 150)
    elif 25 <= current_value < 40:
        percent_increase = random.uniform(65, 120)
    elif 40 <= current_value < 50:
        percent_increase = random.uniform(55, 85)
    elif 50 <= current_value < 60:
        percent_increase = random.uniform(40, 60)
    elif 60 <= current_value < 70:
        percent_increase = random.uniform(35, 44)
    elif 70 <= current_value < 75:
        percent_increase = random.uniform(24, 38)
    elif 75 <= current_value < 85:
        percent_increase = random.uniform(18, 35)
    elif 85 <= current_value < 90:
        percent_increase = random.uniform(13, 25)
    elif 90 <= current_value < 95:
        percent_increase = random.uniform(7, 13)
    else:
        percent_increase = random.uniform(6, 10)

    enhancement = math.ceil(current_value * (percent_increase / 100.0))

    return enhancement


def calculate_potm_enhancement(overall_rating):
    if overall_rating <= 45:
        return random.randint(20, 30)
    elif 45 < overall_rating <= 50:
        return random.randint(18, 24)
    elif 50 < overall_rating <= 55:
        return random.randint(17, 22)
    elif 55 < overall_rating <= 60:
        return random.randint(15, 21)
    elif 60 < overall_rating <= 65:
        return random.randint(13, 19)
    elif 65 < overall_rating <= 70:
        return random.randint(12, 17)
    elif 70 < overall_rating <= 75:
        return random.randint(7, 12)
    elif 75 < overall_rating <= 80:
        return random.randint(5, 8)
    elif 80 < overall_rating <= 85:
        return random.randint(3, 4)
    elif 85 < overall_rating <= 90:
        return random.randint(2, 3)
    elif 90 < overall_rating <= 95:
        return 1
    else:
        return 1


def calculate_potm_individual_stats(current_value):
    if current_value < 50:
        percent_increase = random.uniform(20, 35)
    elif 50 <= current_value < 70:
        percent_increase = random.uniform(10, 19)
    elif 70 <= current_value < 85:
        percent_increase = random.uniform(6, 13)
    else:
        percent_increase = random.uniform(3, 10)

    enhancement = math.ceil(current_value * (percent_increase / 100.0))
    return enhancement


def calculate_special_individual_stats(current_value):
    if current_value <= 45:
        return random.randint(20, 30)
    elif 45 < current_value <= 50:
        return random.randint(18, 24)
    elif 50 < current_value <= 55:
        return random.randint(17, 22)
    elif 55 < current_value <= 60:
        return random.randint(15, 21)
    elif 60 < current_value <= 65:
        return random.randint(13, 19)
    elif 65 < current_value <= 70:
        return random.randint(12, 17)
    elif 70 < current_value <= 75:
        return random.randint(7, 12)
    elif 75 < current_value <= 80:
        return random.randint(5, 8)
    elif 80 < current_value <= 85:
        return random.randint(3, 4)
    elif 85 < current_value <= 90:
        return random.randint(2, 3)
    elif 90 < current_value <= 95:
        return 1
    else:
        return 1


def calculate_special_overall_enhancement(overall_rating):
    if overall_rating <= 45:
        return random.randint(20, 30)
    elif 45 < overall_rating <= 50:
        return random.randint(18, 24)
    elif 50 < overall_rating <= 55:
        return random.randint(17, 22)
    elif 55 < overall_rating <= 60:
        return random.randint(15, 21)
    elif 60 < overall_rating <= 65:
        return random.randint(13, 19)
    elif 65 < overall_rating <= 70:
        return random.randint(12, 17)
    elif 70 < overall_rating <= 75:
        return random.randint(11, 12)
    elif 75 < overall_rating <= 80:
        return random.randint(5, 8)
    elif 80 < overall_rating <= 85:
        return random.randint(2, 4)
    elif 85 < overall_rating <= 90:
        return random.randint(1, 3)
    elif 90 < overall_rating <= 95:
        return 1
    else:
        return 1


def calculate_toty_enhancement(overall_rating):
    if overall_rating <= 70:
        return random.randint(88, 93) - overall_rating
    elif 71 <= overall_rating <= 80:
        return random.randint(90, 94) - overall_rating
    elif 81 <= overall_rating <= 85:
        return random.randint(92, 96) - overall_rating
    elif 86 <= overall_rating <= 90:
        return random.randint(94, 99) - overall_rating
    elif 91 <= overall_rating <= 95:
        return random.randint(95, 99) - overall_rating
    else:
        return 99 - overall_rating


def enhance_toty_card(card_type, stats, individual_stats):
    overall_enhancement = calculate_toty_enhancement(stats.get("Overall", 75))
    stats["Overall"] = min(stats.get("Overall", 75) + overall_enhancement, 99)
    for key, value in individual_stats.items():
        if isinstance(value, (int, float)):
            if key in ["FifaSkillMoves", "FifaWeakFoot"]:
                enhancement = random.randint(0, 1)
                individual_stats[key] = min(value + enhancement, 5)
            else:
                enhancement = calculate_toty_individual_stats(value)
                individual_stats[key] = min(value + enhancement, 99)

    for key, value in stats.items():
        if key != "Overall" and isinstance(value, (int, float)):
            enhancement = calculate_toty_individual_stats(value)
            stats[key] = min(value + enhancement, 99)


def enhance_potm_card(card_type, stats, individual_stats):
    overall_enhancement = calculate_potm_enhancement(stats.get("Overall", 75))
    stats["Overall"] = min(stats.get("Overall", 75) + overall_enhancement, 99)
    for key, value in individual_stats.items():
        if isinstance(value, (int, float)):
            if key in ["FifaSkillMoves", "FifaWeakFoot"]:
                enhancement = random.randint(0, 1)
                individual_stats[key] = min(value + enhancement, 5)
            else:
                enhancement = calculate_potm_individual_stats(value)
                individual_stats[key] = min(value + enhancement, 99)

    for key, value in stats.items():
        if key != "Overall" and isinstance(value, (int, float)):
            enhancement = calculate_potm_individual_stats(value)
            stats[key] = min(value + enhancement, 99)


def enhance_hero_card(card_type, stats, individual_stats):
    overall_enhancement = calculate_hero_enhancement(stats.get("Overall", 75))
    stats["Overall"] = min(stats.get("Overall", 75) + overall_enhancement, 99)
    for key, value in individual_stats.items():
        if isinstance(value, (int, float)):
            if key in ["FifaSkillMoves", "FifaWeakFoot"]:
                enhancement = random.randint(0, 1)
                individual_stats[key] = min(value + enhancement, 5)
            else:
                enhancement = calculate_hero_individual_stats(value)
                individual_stats[key] = min(value + enhancement, 99)

    for key, value in stats.items():
        if key != "Overall" and isinstance(value, (int, float)):
            enhancement = calculate_hero_individual_stats(value)
            stats[key] = min(value + enhancement, 99)


def assign_weak_foot(left_foot_rating, right_foot_rating):
    # Define the new scale for foot ratings
    scale = {
        "Very Strong": 5,
        "Strong": 4,
        "Fairly Strong": 3,
        "Reasonable": 3,
        "Fairly Weak": 2,
        "Weak": 2,
        "Very Weak": 1,
    }

    left = scale.get(left_foot_rating, 0)
    right = scale.get(right_foot_rating, 0)

    weak_foot_strength = min(left, right)
    return weak_foot_strength


def assign_skill_moves(fla_rating):
    if fla_rating >= 90:
        return 5
    elif fla_rating >= 80:
        return 4
    elif fla_rating >= 70:
        return 3
    elif fla_rating >= 30:
        return 2
    else:
        return 1


def calculate_overall_score(selected_row, attribute_weights):
    total_weighted_value = 0
    total_weight = sum(attribute_weights.values())

    for attr, weight in attribute_weights.items():
        attr_value = selected_row.get(attr, 0)
        actual_weight = weight / total_weight
        weighted_value = attr_value * actual_weight
        total_weighted_value += weighted_value
    average_value = total_weighted_value
    scaled_score = 5 * (average_value * 1.05)
    overall_score = min(int(round(scaled_score)), 99)
    return overall_score


def calculate_new_attributes(selected_row, position_attribute_weights, use_alt_mapping):
    attribute_mapping = (
        alt_attribute_mapping if use_alt_mapping else new_attribute_mapping
    )
    raw_position = selected_row.get("Position", "").split(",")[0]
    fifa_position = fifa_position_mapping.get(raw_position, raw_position)
    attribute_weights = position_attribute_weights.get(fifa_position, {})

    category_scores = {}
    individual_attributes = {}
    fla_rating = selected_row.get("Fla", 0) * 5
    fifa_skill_moves = assign_skill_moves(fla_rating)
    left_foot_rating = selected_row.get("Left Foot", "Unknown")
    right_foot_rating = selected_row.get("Right Foot", "Unknown")
    weak_foot = assign_weak_foot(left_foot_rating, right_foot_rating)
    for category, attrs in attribute_mapping.items():
        total_weighted_value = 0
        for attr, weight in attrs:
            attr_value = selected_row.get(attr, 0)
            total_weighted_value += attr_value * weight

            weighted_attr_value = attr_value * 5.025
            scaled_attr_value = round(weighted_attr_value)
            individual_attributes[attr] = scaled_attr_value
        max_possible_weighted_value = sum(weight * 100 for _, weight in attrs)
        category_score = (total_weighted_value / max_possible_weighted_value) * 100
        category_scores[category] = min(int(round(category_score * 5.075)), 99)

    overall_score = calculate_overall_score(selected_row, attribute_weights)

    individual_attributes["FifaWeakFoot"] = weak_foot
    individual_attributes["FifaSkillMoves"] = fifa_skill_moves
    category_scores_df = pd.DataFrame(
        [{**{"Overall": overall_score}, **category_scores}]
    )
    individual_attributes_df = pd.DataFrame(
        [{**{"Overall": overall_score}, **individual_attributes}]
    )

    return category_scores_df, individual_attributes_df


def register_fifa_card_callbacks(app):
    @app.callback(
        Output("fifa-card-modal", "is_open"),
        [Input("open-fifa-card-modal-button", "n_clicks")],
        [State("fifa-card-modal", "is_open")],
    )
    def toggle_fifa_modal(n_clicks, is_open):
        if n_clicks:
            return not is_open
        return is_open

    @app.callback(
        Output("club-cards-output", "src"),
        [Input("club-image-url-store", "data")],
    )
    def update_image_src(image_url):
        if image_url:
            return image_url
        else:
            return "itnowork.png"

    @app.callback(
        [
            Output("club-cards-button", "children"),
            Output("interval-component", "disabled"),
        ],
        [
            Input("club-cards-button", "n_clicks"),
            Input("interval-component", "n_intervals"),
        ],
        [State("club-render-modal", "is_open")],
    )
    def update_button_text(n_clicks, n_intervals, is_modal_open):
        ctx = callback_context
        trigger_id = ctx.triggered[0]["prop_id"].split(".")[0]

        if trigger_id == "club-cards-button" and n_clicks and not is_modal_open:
            return "Loading...", False

        elif trigger_id == "interval-component" and n_intervals > 0:
            return "Generate Club Cards", True

        else:
            return no_update, True

    @app.callback(
        [
            Output("download-fifa-card", "href"),
            Output("download-fifa-card", "download"),
        ],
        [Input("stored-filename", "data")],
        prevent_initial_call=True,
    )
    def generate_download_link(stored_filename):
        if stored_filename:
            return f"/assets/{stored_filename}", stored_filename
        return "", "FIFA_Card_Image.png"

    @app.callback(
        [
            Output("club-image-url-store", "data"),
            Output("club-render-modal", "is_open"),
            Output("club-cards-error", "children"),
        ],
        [
            Input("club-cards-button", "n_clicks"),
        ],
        [
            State("selected-club-store", "data"),
            State("render-images-checkbox", "value"),
            State("render-club-images-checkbox", "value"),
            State("render-subs-checkbox", "value"),
            State("alt-mapping-checkbox", "value"),
            State("formation-dropdown", "value"),
            State("card-type-dropdown", "value"),
            State("num-upgrades", "value"),
            State("session-store", "data"),
            State("table-sorting-filtering", "active_cell"),
            State("table-sorting-filtering", "page_current"),
            State("table-sorting-filtering", "page_size"),
            State("table-sorting-filtering", "derived_virtual_indices"),
            State("table-sorting-filtering", "derived_virtual_data"),
            State("current-index-store", "data"),
        ],
    )
    def generate_club_cards(
        n_clicks,
        selected_club,
        render_images_values,
        render_club_images_values,
        render_subs_values,
        use_alt_mapping,
        formation_value,
        card_type,
        num_upgrades,
        session_data,
        active_cell,
        page_current,
        page_size,
        derived_virtual_indices,
        derived_virtual_data,
        current_index_data,
    ):
        print(f"selected club is {selected_club}")
        error_message = ""
        render_images_bool = (
            1 in render_images_values
            if isinstance(render_images_values, list)
            else render_images_values
        )
        render_club_images_bool = (
            1 in render_club_images_values
            if isinstance(render_club_images_values, list)
            else render_club_images_values
        )
        render_subs_bool = (
            1 in render_subs_values
            if isinstance(render_subs_values, list)
            else render_subs_values
        )
        if not n_clicks or not session_data:
            return "", False, ""

        try:
            selected_formation = FORMATIONS.get(formation_value, FORMATIONS["4-3-3(2)"])
            full_df = pd.DataFrame(session_data)
            full_df = full_df[full_df["Position"] != "GK"]
            club_players = full_df[full_df["Club"] == selected_club]
            if club_players.empty:
                return "", False, f"No players found for the club: {selected_club}"

            def calculate_fifa_details(row):
                positions = row["Position"].split(",") if row["Position"] else []
                max_overall_score = 0
                primary_position = None
                primary_position_fifa = None
                for pos in positions:
                    pos = pos.strip()
                    if not pos:
                        continue

                    row["Position"] = pos
                    fifa_position = fifa_position_mapping.get(pos, pos)
                    attribute_weights = position_attribute_weights.get(
                        fifa_position, {}
                    )
                    overall_score = calculate_overall_score(row, attribute_weights)

                    if overall_score > max_overall_score:
                        max_overall_score = overall_score
                        primary_position = pos
                        primary_position_fifa = fifa_position

                alternate_positions = [
                    fifa_position_mapping.get(p.strip(), p.strip())
                    for p in positions
                    if p.strip() != primary_position
                ]
                alternate_positions = alternate_positions[:3]

                return primary_position_fifa, alternate_positions, max_overall_score

            try:
                club_players[
                    [
                        "fifa_primary_position",
                        "fifa_alternate_positions",
                        "fifa_overall",
                    ]
                ] = club_players.apply(
                    lambda row: calculate_fifa_details(row),
                    axis=1,
                    result_type="expand",
                )

                grid_image_url = create_image_grid(
                    club_players,
                    selected_formation,
                    card_type,
                    render_images_bool,
                    render_club_images_bool,
                    render_subs_bool,
                    num_upgrades,
                    use_alt_mapping,
                )

                if grid_image_url:
                    if grid_image_url.startswith("./"):
                        grid_image_url = f"{grid_image_url}?t={time.time()}"
                    return grid_image_url, True, ""

                else:
                    return "", False, ""
            except ValueError as e:
                return "", False, str(e)

            except IndexError as e:
                return (
                    "",
                    False,
                    "Not enough players at the club in the given dataset.",
                )
        except Exception as e:
            return "", False, f"An unexpected error occurred: {str(e)}"


base_image_cache = {}


def get_cached_base_image(path):
    if path not in base_image_cache:
        base_image_cache[path] = Image.open(path)
    return base_image_cache[path]


def create_fifa_card(
    player_row,
    card_type,
    render_images_values,
    render_club_images_values,
    num_upgrades,
    use_alt_mapping,
):
    player_name = player_row.get("Name", "Unknown")
    positions = player_row.get("Position", "")

    if not player_name or not positions:
        return None

    player_name = player_name.replace(" ", "_")
    positions = positions.split(",")

    max_overall_score = 0
    primary_position = None

    for pos in positions:
        pos = pos.strip()
        if not pos:
            continue

        player_row["Position"] = pos
        fifa_position = fifa_position_mapping.get(pos, pos)

        overall_score = calculate_overall_score(
            player_row, position_attribute_weights.get(fifa_position, {})
        )

        if overall_score > max_overall_score:
            max_overall_score = overall_score
            primary_position = fifa_position

    if primary_position is None:
        return None

    base_image_path = get_base_image_path(max_overall_score, card_type)
    base_image = get_cached_base_image(base_image_path)

    all_positions = [fifa_position_mapping.get(p.strip(), p.strip()) for p in positions]
    secondary_positions = [pos for pos in all_positions if pos != primary_position][:3]

    new_attributes_df, individual_attributes_df = calculate_new_attributes(
        player_row, position_attribute_weights, use_alt_mapping
    )
    stats = new_attributes_df.iloc[0].to_dict()
    stats["Overall"] = max_overall_score
    if (
        card_type == "normal"
        and max_overall_score >= 75
        and all(value < 80 for key, value in stats.items() if key != "Overall")
    ):
        base_image_path = "./fifa_card_assets/cards/0_gold.png"
    elif (
        card_type == "normal"
        and 64 < max_overall_score <= 75
        and all(value < 70 for key, value in stats.items() if key != "Overall")
    ):
        base_image_path = "./fifa_card_assets/cards/0_silver.png"
    elif (
        card_type == "normal"
        and max_overall_score <= 64
        and all(value < 60 for key, value in stats.items() if key != "Overall")
    ):
        base_image_path = "./fifa_card_assets/cards/0_bronze.png"
    else:
        base_image_path = get_base_image_path(max_overall_score, card_type)

    base_image = Image.open(base_image_path)
    uploaded_player_image = None
    individual_stats = individual_attributes_df.iloc[0].to_dict()
    modified_image = overlay_fifa_card_stats(
        base_image,
        player_row,
        stats,
        [primary_position] + secondary_positions,
        individual_stats,
        card_type,
        render_images_values,
        render_club_images_values,
        num_upgrades,
        uploaded_player_image,
    )

    unique_id = uuid.uuid4()
    unique_filename = f"fifa_card_{player_name}_{primary_position}_{unique_id}.png"
    output_image_path = f"./assets/{unique_filename}"
    modified_image.save(output_image_path)

    image_url = f"./assets/{unique_filename}"
    return image_url


def overlay_fifa_card_stats(
    image,
    selected_row,
    stats,
    positions,
    individual_stats,
    card_type,
    render_images,
    render_club_images,
    num_upgrades,
    uploaded_player_image,
):
    player_image_path = None
    player_name = selected_row.get("Name")
    nationality = selected_row.get("Nationality", "").upper()
    render_images_values = render_images
    render_club_images_values = render_club_images

    if image.mode != "RGBA":
        image = image.convert("RGBA")

    base_color = image.getpixel((200, 200))
    draw = ImageDraw.Draw(image)
    if card_type == "normal" and stats.get("Overall", 75) >= 75:
        color_set = colors.get("default", colors["default"])
    elif card_type == "normal" and 64 < stats.get("Overall", 75) <= 75:
        color_set = colors.get("silver", colors["default"])
    elif card_type == "normal" and stats.get("Overall", 75) <= 64:
        color_set = colors.get("bronze", colors["default"])
    else:
        color_set = colors.get(card_type, colors["default"])
    text_color = color_set["text_color"]
    circle_fill_color = color_set["circle_fill_color"]
    circle_outline_color = color_set["circle_outline_color"]
    rect_fill_color = color_set["rect_fill_color"]
    rect_outline_color = color_set["rect_outline_color"]
    rect_text_color = color_set["rect_text_color"]

    if card_type == "icon":
        overall_enhancement = calculate_icon_enhancement(stats.get("Overall", 75))
        stats["Overall"] = min(stats.get("Overall", 75) + overall_enhancement, 99)
        for key, value in individual_stats.items():
            if isinstance(value, (int, float)):
                if key == "FifaSkillMoves" or key == "FifaWeakFoot":
                    enhancement = random.randint(0, 1)
                    individual_stats[key] = min(value + enhancement, 5)
                else:
                    enhancement = calculate_icon_individual_stats(value)
                    individual_stats[key] = min(value + enhancement, 99)

        for key, value in stats.items():
            if key != "Overall" and isinstance(value, (int, float)):
                enhancement = calculate_icon_individual_stats(value)
                stats[key] = min(value + enhancement, 99)

    font_overall = get_font("bold", 72)
    font_primary_position = get_font("bold", 48)
    font_secondary_position = get_font("medium", 24)
    font_bold = get_font("bold", 48)
    font_med = get_font("medium", 34)
    font_rect = get_font("medium", 30)
    font_name = get_font("bold", 60)

    original_overall = stats["Overall"]
    weak_foot = individual_stats.get("FifaWeakFoot", 0)
    if weak_foot == 5:
        stats["Overall"] = min(stats.get("Overall", 0) + 1, 99)
    elif weak_foot == 1:
        stats["Overall"] = max(stats.get("Overall", 0) - 1, 1)
    skill_moves = individual_stats.get("FifaSkillMoves", 0)
    if skill_moves == 5:
        stats["Overall"] = min(stats.get("Overall", 0) + 1, 99)
    hero_card_types = ["hero", "ucl_hero"]
    if card_type in hero_card_types:
        enhance_hero_card(card_type, stats, individual_stats)
    toty_card_types = ["toty"]
    if card_type in toty_card_types:
        enhance_toty_card(card_type, stats, individual_stats)
    for _ in range(num_upgrades):
        potm_card_types = [
            "potm_epl",
            "potm_bundesliga",
            "potm_laliga",
            "potm_ligue1",
            "potm_serie_a",
            "potm_mls",
            "potm_eredivisie",
            "potm_laliga",
        ]
        if card_type in potm_card_types:
            enhance_potm_card(card_type, stats, individual_stats)
    for _ in range(num_upgrades):
        if card_type in [
            "totw",
            "fut_champs",
            "motm",
            "totgs_ucl",
            "totgs_uel",
            "totgs_uecl",
            "rttk_ucl",
            "rttk_uel",
            "rttk_uecl",
        ]:
            overall_enhancement = calculate_special_overall_enhancement(
                stats.get("Overall", 75)
            )
            stats["Overall"] = min(stats.get("Overall", 75) + overall_enhancement, 99)
            for key, value in individual_stats.items():
                if isinstance(value, (int, float)):
                    if key == "FifaSkillMoves" or key == "FifaWeakFoot":
                        enhancement = random.randint(0, 1)
                        individual_stats[key] = min(value + enhancement, 5)
                    else:
                        enhancement = calculate_special_individual_stats(value)
                        enhancement_upgrade = random.randint(
                            enhancement, enhancement + 3
                        )
                        individual_stats[key] = min(value + enhancement_upgrade, 99)
            for key, value in stats.items():
                if key != "Overall" and isinstance(value, (int, float)):
                    enhancement = calculate_special_individual_stats(value)
                    stats[key] = min(value + enhancement, 99)

    best_match = None
    max_criteria_matched = 0
    max_attributes_required = 0
    criteria_met = False
    for image_name, attributes in criteria_map.items():
        criteria_matched = sum(
            individual_stats.get(attr, 0) >= 90 for attr in attributes
        )
        if criteria_matched == len(attributes):
            if criteria_matched > max_criteria_matched or (
                criteria_matched == max_criteria_matched
                and len(attributes) > max_attributes_required
            ):
                if stats["Overall"] >= 75:
                    best_match = image_name
                    max_criteria_matched = criteria_matched
                    max_attributes_required = len(attributes)
                    criteria_met = True
                else:
                    best_match = None
                    criteria_met = False
                    max_criteria_matched = 0
                    max_attributes_required = 0

    if criteria_met and card_type != "totw":
        stats["Overall"] = min(stats.get("Overall", 0) + 1, 99)
    new_overall = stats["Overall"]
    old_image_tier = get_base_image_path(original_overall, card_type)
    new_image_tier = get_base_image_path(new_overall, card_type)
    if card_type != "normal" and 64 < stats.get("Overall", 75) < 75:
        color_set = colors.get("if_silver", colors["icon"])
    elif card_type != "normal" and stats.get("Overall", 75) <= 64:
        color_set = colors.get("if_bronze", colors["default"])
    if new_image_tier == "./fifa_card_assets/cards/if_silver.png":
        color_set = colors.get("if_silver", colors["default"])
    if new_image_tier == "./fifa_card_assets/cards/if_bronze.png":
        color_set = colors.get("if_bronze", colors["default"])

    if old_image_tier != new_image_tier:
        image = Image.open(new_image_tier).convert("RGBA")
        draw = ImageDraw.Draw(image)
    text_color = color_set["text_color"]
    circle_fill_color = color_set["circle_fill_color"]
    circle_outline_color = color_set["circle_outline_color"]
    rect_fill_color = color_set["rect_fill_color"]
    rect_outline_color = color_set["rect_outline_color"]
    rect_text_color = color_set["rect_text_color"]
    full_name = selected_row.get("Name", "Unknown")
    last_name = full_name.split()[-1] if full_name else "Unknown"
    text_width, _ = draw.textsize(last_name, font=font_name)
    name_x = (image.width - text_width) // 2
    name_y = image.height * 2 // 2.95 - 50
    overall_x, overall_y = 110, 110
    overall_text = f"{stats.pop('Overall')}"
    primary_position = fifa_position_mapping.get(
        positions[0].strip(), positions[0].strip()
    )
    draw.text((overall_x, overall_y), overall_text, font=font_overall, fill=text_color)

    position_y = overall_y + 75
    overall_text_width = draw.textsize(overall_text, font=font_overall)[0]
    position_text_width = draw.textsize(primary_position, font=font_primary_position)[0]
    adjusted_position_x = overall_x + (overall_text_width - position_text_width) / 2

    # Drawing the Primary Position
    draw.text(
        (adjusted_position_x, position_y),
        primary_position,
        font=font_primary_position,
        fill=text_color,
    )

    # Drawing the player's last name
    draw.text((name_x, name_y), last_name, font=font_name, fill=text_color)

    if best_match:
        special_image = Image.open(
            f"./fifa_card_assets/playstyles/{card_type}_playstyles/{best_match}"
        )
        special_image = special_image.resize(
            (int(special_image.width * 2 / 3), int(special_image.height * 2 / 3)),
            Image.ANTIALIAS,
        )
        paste_x = int(image.width * 0.035)
        paste_y = int(image.height * 0.4)
        image.paste(special_image, (paste_x, paste_y), special_image)

    # Secondary Positions in Circles
    circle_radius = 30
    circle_y = position_y + 110
    circle_scale_factor = 4
    circle_outline_thickness = 2

    for pos in positions[1:4]:
        fifa_position = fifa_position_mapping.get(pos.strip(), pos.strip())
        circle_diameter = 2 * circle_radius
        circle_canvas_size = (
            circle_diameter * circle_scale_factor,
            circle_diameter * circle_scale_factor,
        )
        circle_canvas = Image.new("RGBA", circle_canvas_size, (0, 0, 0, 0))
        circle_draw = ImageDraw.Draw(circle_canvas)

        hr_circle_radius = circle_radius * circle_scale_factor
        hr_outline_thickness = circle_outline_thickness * circle_scale_factor
        circle_draw.ellipse(
            [
                (hr_outline_thickness, hr_outline_thickness),
                (
                    2 * hr_circle_radius - hr_outline_thickness,
                    2 * hr_circle_radius - hr_outline_thickness,
                ),
            ],
            fill=circle_fill_color,
            outline=circle_outline_color,
            width=hr_outline_thickness,
        )
        scaled_circle = circle_canvas.resize(
            (circle_diameter, circle_diameter), Image.ANTIALIAS
        )
        circle_x = overall_x + circle_radius - 78
        paste_position = (int(circle_x - circle_radius), int(circle_y - circle_radius))
        image.paste(scaled_circle, paste_position, scaled_circle)

        pos_text_width, pos_text_height = draw.textsize(
            fifa_position, font=font_secondary_position
        )
        text_x = circle_x - pos_text_width / 2
        text_y = circle_y - pos_text_height / 2 - 3
        draw.text(
            (text_x, text_y),
            fifa_position,
            font=font_secondary_position,
            fill=text_color,
        )

        circle_y -= 2 * circle_radius + 10

    horizontal_spacing = 20
    total_width = (
        sum(
            max(
                draw.textsize(stat_name, font=font_med)[0],
                draw.textsize(str(stat_value), font=font_bold)[0],
            )
            + horizontal_spacing
            for stat_name, stat_value in stats.items()
        )
        - horizontal_spacing
    )
    start_x = (image.width - total_width) // 2

    attribute_name_y = image.height * 2 // 2.8
    attribute_value_y = attribute_name_y + 30
    for stat_name, stat_value in stats.items():
        name_width, _ = draw.textsize(stat_name, font=font_med)
        value_width, _ = draw.textsize(str(stat_value), font=font_bold)
        max_width = max(name_width, value_width)

        text_x = start_x + (max_width - name_width) // 2
        draw.text((text_x, attribute_name_y), stat_name, font=font_med, fill=text_color)

        text_x = start_x + (max_width - value_width) // 2
        draw.text(
            (text_x, attribute_value_y),
            str(stat_value),
            font=font_bold,
            fill=text_color,
        )

        start_x += max_width + horizontal_spacing

    image_loaded = False

    if uploaded_player_image is not None:
        render_images_values = True

    if render_images_values:
        player_image = None

        if uploaded_player_image is not None:
            player_image = uploaded_player_image
            image_loaded = True
        else:
            player_image = player_image_cache.get(player_name)

            if not player_image:
                player_image_path = get_or_download_player_image(player_name)
                new_height = 384

                if player_image_path:
                    try:
                        player_img = Image.open(player_image_path).convert("RGBA")
                        orig_width, orig_height = player_img.size
                        aspect_ratio = orig_width / orig_height
                        new_width = int(new_height * aspect_ratio)
                        player_img = player_img.resize(
                            (new_width, new_height), Image.ANTIALIAS
                        )
                        player_x = (image.width - new_width) // 2 + 15
                        player_y = (image.height - new_height) // 2 - 70
                        image.paste(player_img, (player_x, player_y), player_img)
                        image_loaded = True
                        player_image = player_img
                    except FileNotFoundError:
                        print(
                            f"Player image not found at {player_image_path}, attempting to use fallback..."
                        )
                    except Exception as e:
                        print(f"Unexpected error occurred: {e}")

                if not image_loaded:
                    print("Attempting to use fallback player image...")
                    nationality = selected_row.get("Nationality")
                    regen_directory = (
                        f"fifa_card_assets/player_img/regen/{nationality}/"
                    )
                    try:
                        if os.path.exists(regen_directory):
                            regen_images = [
                                f
                                for f in os.listdir(regen_directory)
                                if f.endswith(".png")
                            ]
                            if regen_images:
                                random_regen_image = random.choice(regen_images)
                                regen_image_path = os.path.join(
                                    regen_directory, random_regen_image
                                )
                                regen_img = Image.open(regen_image_path).convert("RGBA")
                                image_loaded = True
                            else:
                                print(
                                    f"No local regen images found for {nationality}, trying API..."
                                )
                        nationality_directory_url = (
                            f"{fallback_image_url}{nationality}/"
                        )
                        directory_response = requests.get(
                            nationality_directory_url, timeout=1
                        )
                        if directory_response.status_code == 200:
                            directory_soup = BeautifulSoup(
                                directory_response.content, "html.parser"
                            )
                            image_links = [
                                link.get("href")
                                for link in directory_soup.find_all("a")
                                if link.get("href").endswith(".png")
                            ]
                            if image_links:
                                random_image_url = (
                                    nationality_directory_url
                                    + random.choice(image_links)
                                )
                                image_response = requests.get(
                                    random_image_url, timeout=1
                                )
                                if image_response.status_code == 200:
                                    regen_img = Image.open(
                                        BytesIO(image_response.content)
                                    ).convert("RGBA")
                                    image_loaded = True
                                else:
                                    print("Failed to load image from API")
                            else:
                                print(f"No images found at {nationality_directory_url}")
                        else:
                            print(
                                f"Failed to access directory at {nationality_directory_url}: {directory_response.status_code}"
                            )

                        if image_loaded:
                            orig_width, orig_height = regen_img.size
                            aspect_ratio = orig_width / orig_height
                            new_width = int(new_height * aspect_ratio)
                            regen_img = regen_img.resize(
                                (new_width, new_height), Image.ANTIALIAS
                            )
                            player_x = (image.width - new_width) // 2 + 15
                            player_y = (image.height - new_height) // 2 - 70
                            image.paste(regen_img, (player_x, player_y), regen_img)
                            player_image = regen_img

                    except Exception as e:
                        print(f"Failed to load fallback image: {e}")
                if player_image and not uploaded_player_image:
                    player_image_cache[player_name] = player_image

        if player_image:
            new_height = 384
            orig_width, orig_height = player_image.size
            aspect_ratio = orig_width / orig_height
            new_width = int(new_height * aspect_ratio)
            player_image = player_image.resize((new_width, new_height), Image.ANTIALIAS)

            if player_image.mode != "RGBA":
                player_image = player_image.convert("RGBA")

            if image.mode != "RGBA":
                image = image.convert("RGBA")

            player_x = (image.width - new_width) // 2 + 15
            player_y = (image.height - new_height) // 2 - 70

            image.paste(player_image, (player_x, player_y), player_image)

    box_color = base_color
    box_width, box_height = 100, 30
    box_gap = 10
    box_x = image.width - box_width - box_gap - 10
    wf_y, sm_y = (
        image.height // 2 - box_height - 10,
        image.height // 2 + 10,
    )

    wf_text = f"WF - {individual_stats.get('FifaWeakFoot', '')}"
    wf_text_size = draw.textsize(wf_text, font=font_med)
    wf_text_x = box_x + (box_width - wf_text_size[0]) // 2
    wf_text_y = wf_y + (box_height - wf_text_size[1]) // 2

    sm_text = f"SM - {individual_stats.get('FifaSkillMoves', '')}"
    sm_text_size = draw.textsize(sm_text, font=font_med)
    sm_text_x = box_x + (box_width - sm_text_size[0]) // 2
    sm_text_y = sm_y + (box_height - sm_text_size[1]) // 2

    rect_scale_factor = 4
    rect_outline_thickness = 2

    def create_scaled_rectangle(
        box_width, box_height, fill, outline, rect_scale_factor, rect_outline_thickness
    ):
        hr_canvas_size = (box_width * rect_scale_factor, box_height * rect_scale_factor)
        hr_rectangle = Image.new("RGBA", hr_canvas_size, (0, 0, 0, 0))
        hr_draw = ImageDraw.Draw(hr_rectangle)
        hr_draw.rectangle(
            [
                (rect_outline_thickness, rect_outline_thickness),
                (
                    box_width * rect_scale_factor - rect_outline_thickness,
                    box_height * rect_scale_factor - rect_outline_thickness,
                ),
            ],
            fill=fill,
            outline=outline,
            width=rect_outline_thickness * rect_scale_factor,
        )

        scaled_rectangle = hr_rectangle.resize((box_width, box_height), Image.ANTIALIAS)
        return scaled_rectangle

    wf_rectangle = create_scaled_rectangle(
        box_width,
        box_height,
        rect_fill_color,
        rect_outline_color,
        rect_scale_factor,
        rect_outline_thickness,
    )
    sm_rectangle = create_scaled_rectangle(
        box_width,
        box_height,
        rect_fill_color,
        rect_outline_color,
        rect_scale_factor,
        rect_outline_thickness,
    )

    flag_path = None
    if render_club_images_values == True:
        flag_path = f"fifa_card_assets/nation_images/{nationality}.png"
        try:
            flag_img = Image.open(flag_path).convert("RGBA")
            aspect_ratio = flag_img.width / flag_img.height
            new_width = 56
            new_height = int(new_width / aspect_ratio)

            flag_img = flag_img.resize((new_width, new_height), Image.ANTIALIAS)

            flag_x = image.width - new_width - 360
            flag_y = image.height - new_height - 120

            image.paste(flag_img, (flag_x, flag_y), flag_img)
            club = selected_row.get("Club")
            if render_club_images_values and club and club != "0":
                image = integrate_league_image(image, selected_row)
            if not club or club == "0":
                placeholder_img_path = "fifa_card_assets/free_agent.png"
                placeholder_img = Image.open(placeholder_img_path).convert("RGBA")
                aspect_ratio = placeholder_img.width / placeholder_img.height
                new_height = 48
                new_width = int(new_height * aspect_ratio)
                placeholder_img = placeholder_img.resize(
                    (new_width, new_height), Image.ANTIALIAS
                )
                club_x = (image.width - new_width) // 2 + 60
                club_y = image.height - new_height - 110

                image.paste(placeholder_img, (club_x, club_y), placeholder_img)
                print(f"Using placeholder image for free agent {player_name}")
            else:
                image = integrate_scraping(
                    image, selected_row, render_club_images_values
                )
        except FileNotFoundError:
            print(f"Flag for {nationality} not found at {flag_path}")
    image.paste(wf_rectangle, (box_x, wf_y), wf_rectangle)
    image.paste(sm_rectangle, (box_x, sm_y), sm_rectangle)
    draw.text((wf_text_x, wf_text_y), wf_text, font=font_rect, fill=rect_text_color)
    draw.text((sm_text_x, sm_text_y), sm_text, font=font_rect, fill=rect_text_color)

    return image


def create_image_grid(
    club_players,
    formation,
    card_type,
    render_images,
    render_club_images,
    render_subs,
    num_upgrades,
    use_alt_mapping,
):
    try:
        min_players_no_subs = 10
        min_players_with_subs = 17
        min_players_required = (
            min_players_with_subs if render_subs else min_players_no_subs
        )
        if len(club_players) < min_players_required:
            raise ValueError(
                f"Not enough players in the club to form a team. At least {min_players_required} players are required."
            )

        club_players = preprocess_player_data(club_players)
        position_player_map = map_players_to_positions(club_players)
        selected_players = determine_selected_players(
            position_player_map, formation, club_players
        )
        reference_image = Image.open("./fifa_card_assets/cards/1_gold.png")
        img_width, img_height = reference_image.size
        additional_spacing = 20
        total_player_space = img_width + additional_spacing
        grid_width = max(len(row.split(",")) for row in formation) * total_player_space
        grid_image = Image.new("RGBA", (grid_width, 0))
        position_occurrences = {
            pos.strip(): 0 for row in formation for pos in row.split(",")
        }
        y_offset_accumulator = 0

        image_creation_tasks = []
        for row_idx, formation_row in enumerate(formation):
            positions_in_row = formation_row.split(",")
            x_start = (grid_width - len(positions_in_row) * total_player_space) // 2
            row_max_up = max(
                abs(position_offsets.get(pos.strip(), {"vertical": 0})["vertical"])
                for pos in positions_in_row
            )
            row_max_down = max(
                position_offsets.get(pos.strip(), {"vertical": 0})["vertical"]
                for pos in positions_in_row
            )
            row_height = img_height + row_max_up + row_max_down

            new_grid_height = grid_image.size[1] + row_height
            expanded_grid_image = Image.new("RGBA", (grid_width, new_grid_height))
            expanded_grid_image.paste(grid_image, (0, 0))
            grid_image = expanded_grid_image

            for i, position in enumerate(positions_in_row):
                position = position.strip()
                occurrence = position_occurrences[position]
                position_key = f"{position}_{occurrence}"
                offsets = position_offsets.get(
                    position, {"vertical": 0, "horizontal": 0}
                )
                x = x_start + i * total_player_space + offsets["horizontal"]
                y = y_offset_accumulator + row_max_up + offsets["vertical"]
                if position_key in selected_players:
                    player_name = selected_players[position_key]
                    player_row = club_players.loc[
                        club_players["Name"] == player_name
                    ].iloc[0]
                    image_creation_tasks.append((player_row, x, y))
                position_occurrences[position] += 1

            y_offset_accumulator += row_height

        grid_height = y_offset_accumulator
        if render_subs is True:
            all_player_names = set(club_players["Name"])
            selected_player_names = set(selected_players.values())
            unselected_players = list(all_player_names - selected_player_names)
            unselected_players_best = sorted(
                unselected_players,
                key=lambda x: club_players.loc[
                    club_players["Name"] == x, "fifa_overall"
                ].iloc[0],
                reverse=True,
            )[:7]

            scale_factor = 0.60
            scaled_img_width = int(img_width * scale_factor)
            scaled_img_height = int(img_height * scale_factor)
            closer_spacing = -20
            total_unselected_player_space = scaled_img_width + closer_spacing

            grid_height += scaled_img_height

            x_start_unselected = (grid_width - 7 * total_unselected_player_space) // 2
            y_start_unselected = y_offset_accumulator

            for i, player_name in enumerate(unselected_players_best):
                x_unselected = x_start_unselected + i * total_unselected_player_space
                player_row = club_players.loc[club_players["Name"] == player_name].iloc[
                    0
                ]
                image_creation_tasks.append(
                    (player_row, x_unselected, y_start_unselected)
                )

        required_width = grid_width
        if render_subs is True:
            required_width = max(
                x_start_unselected + 7 * total_unselected_player_space, grid_width
            )

        grid_image_extended = Image.new("RGBA", (required_width, grid_height))
        grid_image_extended.paste(grid_image, (0, 0))
        images_info = []
        with ThreadPoolExecutor() as executor:
            future_to_image = {
                executor.submit(
                    create_fifa_card,
                    task[0],
                    card_type,
                    render_images,
                    render_club_images,
                    num_upgrades,
                    use_alt_mapping,
                ): task
                for task in image_creation_tasks
            }
            for future in as_completed(future_to_image):
                player_row, x, y = future_to_image[future]
                image_url = future.result()
                if image_url:
                    is_unselected = render_subs and (
                        player_row["Name"] in unselected_players
                    )
                    images_info.append((image_url, x, y, is_unselected))

        for image_url, x, y, is_unselected in images_info:
            with Image.open(image_url) as img:
                if is_unselected:
                    img = img.resize(
                        (scaled_img_width, scaled_img_height), Image.Resampling.BICUBIC
                    )
                if (
                    0 <= x <= grid_image_extended.width - img.width
                    and 0 <= y <= grid_image_extended.height - img.height
                ):
                    grid_image_extended.paste(img, (x, y), img)
                else:
                    print(
                        f"Skipping image at ({x},{y}), as it's outside the grid bounds."
                    )
            os.remove(image_url)
        unique_id = uuid.uuid4()
        grid_image_path = f"./assets/club_{unique_id}.webp"
        grid_image_extended.save(grid_image_path, format="WEBP", quality=90)

        return grid_image_path

    except Exception as e:
        print(f"An error occurred in create_image_grid: {e}")
        raise


def determine_selected_players(position_player_map, formation, club_players):
    selected_players = {}
    used_players = set()

    flat_formation = [pos.strip() for row in formation for pos in row.split(",")]
    position_assignment_count = {position: 0 for position in set(flat_formation)}

    for position in flat_formation:
        if position in position_player_map:
            sorted_players = sorted(
                position_player_map[position], key=lambda x: x[1], reverse=True
            )

            for player_idx, _ in sorted_players:
                if player_idx not in used_players:
                    player_name = club_players.at[player_idx, "Name"]
                    position_key = f"{position}_{position_assignment_count[position]}"
                    selected_players[position_key] = player_name
                    used_players.add(player_idx)
                    position_assignment_count[position] += 1
                    break
            else:
                position_key = f"{position}_{position_assignment_count[position]}"
                selected_players[position_key] = None
                position_assignment_count[position] += 1
        else:
            position_key = f"{position}_{position_assignment_count[position]}"
            selected_players[position_key] = None
            position_assignment_count[position] += 1

    return selected_players


def preprocess_player_data(club_players):
    club_players["fifa_primary_position"] = (
        club_players["fifa_primary_position"].fillna("").astype(str)
    )
    club_players["fifa_alternate_positions"] = club_players[
        "fifa_alternate_positions"
    ].apply(lambda x: x if isinstance(x, list) else [])

    club_players["eligible_positions"] = club_players.apply(
        lambda row: [row["fifa_primary_position"].strip()]
        + [pos.strip() for pos in row["fifa_alternate_positions"]],
        axis=1,
    )
    return club_players


def map_players_to_positions(club_players):
    position_player_map = {}
    for row in club_players.itertuples():
        unique_positions = set()
        for fm_position in row.eligible_positions:
            fifa_position = fifa_position_mapping.get(fm_position, fm_position)
            if fifa_position not in unique_positions:
                unique_positions.add(fifa_position)
                position_player_map.setdefault(fifa_position, []).append(
                    (row.Index, row.fifa_overall)
                )

    for position in position_player_map:
        position_player_map[position] = sorted(
            position_player_map[position], key=lambda x: x[1], reverse=True
        )

    return position_player_map
