"""
This module provides functions for calculating and formatting value scores for players.
"""

import re
import numpy as np
from typing import Tuple
import time
import logging


clean_value_regex = re.compile(r"[^\d.]")

multipliers = {"M": 1_000_000, "K": 1_000, "": 1}

periodicity_multipliers = {
    "p/w": 1,  # weekly to weekly
    "p/m": 1 / 4,  # monthly to weekly
    "p/a": 1 / 52,  # annual to weekly
}


def calculate_value_score_vectorized(overall_scores, upper_value_ranges, alpha=1):
    start_time = time.time()
    overall_scores = np.asarray(overall_scores).flatten()
    upper_value_ranges = np.asarray(upper_value_ranges).flatten()

    np.maximum(overall_scores, 0, out=overall_scores)
    np.maximum(upper_value_ranges, 0, out=upper_value_ranges)

    value_scores = np.zeros_like(overall_scores)

    valid_range_mask = upper_value_ranges > 0

    corrected_upper_value_ranges = upper_value_ranges[valid_range_mask] / 1_000_000
    normalized_overall_scores = overall_scores[valid_range_mask] / np.max(
        overall_scores[valid_range_mask]
    )
    normalized_upper_value_ranges = corrected_upper_value_ranges / np.max(
        corrected_upper_value_ranges
    )

    log_overall_scores = np.log1p(normalized_overall_scores)
    log_upper_values = np.log1p(normalized_upper_value_ranges)

    value_scores[valid_range_mask] = (
        log_overall_scores - alpha * log_upper_values
    ) * 100

    end_time = time.time()
    total_time = end_time - start_time
    logging.debug(
        "Time taken to calculate value score vectorized: %.4f seconds", total_time
    )

    return value_scores


def convert_transfer_fee_to_number(currency_string: str) -> Tuple[float, float]:
    """
    Converts a currency string to a numeric value. Handles None input by returning (0.0, 0.0).
    """

    if currency_string is None:
        return 0.0, 0.0

    if " - " in currency_string:
        lower_str, upper_str = currency_string.split(" - ")
    else:
        lower_str = upper_str = currency_string

    def convert_single_value(value_str):
        cleaned_value_str = clean_value_regex.sub("", value_str)
        if not cleaned_value_str:
            return 0.0

        suffix = value_str[-1] if value_str[-1] in multipliers else ""
        multiplier = multipliers[suffix]

        return float(cleaned_value_str) * multiplier

    lower_value = convert_single_value(lower_str)
    upper_value = convert_single_value(upper_str)

    return lower_value, upper_value


def convert_wage_to_weekly(wage_string: str) -> float:
    """
    Converts a wage string to a weekly numeric value. Handles None input by returning 0.0.
    """

    if wage_string is None:
        return 0.0

    cleaned_value_str = clean_value_regex.sub("", wage_string)
    if not cleaned_value_str:
        return 0.0

    # Determine periodicity and get the corresponding multiplier
    periodicity = re.findall(r"p/w|p/m|p/a", wage_string)
    multiplier = periodicity_multipliers[periodicity[0]] if periodicity else 1

    return (float(cleaned_value_str) * multiplier) / 1_000
