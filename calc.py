"""
This module provides functionality for calculating player scores based on various attributes.
"""

from typing import List, Dict
import pandas as pd
from pandas import DataFrame
from tracer_provider import tracer
from opentelemetry import context
import os
import time
import logging
import yaml

import values


def load_config():
    config_path = os.getenv("FMD_CONF_LOCATION", "config.yml")
    with open(config_path) as f:
        return yaml.safe_load(f)


yaml_config = load_config()
attributes = yaml_config["attributes"]
attributes_physical = list(attributes["physical"].values())
attributes_mental = list(attributes["mental"].values())
attributes_technical = list(attributes["technical"].values())
results_to_return = yaml_config["results_to_return"]


def convert_to_float(val):
    if isinstance(val, str) and "%" in val:
        try:
            return float(val.replace("%", ""))
        except ValueError:
            return 0

    try:
        return float(val)
    except ValueError:
        return 0


def calculate_category_score(
    dataframe: DataFrame, selected_weights: Dict[str, float]
) -> pd.Series:
    """
    Calculate the category score based on the provided dataframe and selected weights.

    Parameters:
        dataframe (DataFrame): The input dataframe containing the data.
        selected_weights (Dict[str, float]): The weights for each attribute.

    Returns:
        pd.Series: The calculated category score.

    """
    valid_attrs = selected_weights.keys() & dataframe.columns
    weights = pd.Series(selected_weights)[list(valid_attrs)]
    weighted_sums = dataframe[list(valid_attrs)].mul(weights).sum(axis=1)
    return weighted_sums.round(2)


def calculate_player_scores(
    dataframe: DataFrame, selected_weights: Dict[str, float]
) -> DataFrame:
    """
    Calculate player scores based on the provided dataframe and selected weights.

    Args:
        dataframe (DataFrame): The input dataframe containing player data.
        selected_weights (Dict[str, float]): A dictionary of attribute weights for scoring.

    Returns:
        DataFrame: The dataframe with calculated player scores.

    Raises:
        Exception: If an error occurs during the calculation.

    """
    current_ctx = context.get_current()
    with tracer.start_as_current_span(
        "calculate_player_scores", context=current_ctx
    ) as current_span:
        start_time = time.time()
        categories = ["physical_score", "mental_score", "technical_score"]
        attributes = [attributes_physical, attributes_mental, attributes_technical]

        try:
            for category, attrs in zip(categories, attributes):
                valid_attrs = [attr for attr in attrs if attr in dataframe.columns]
                weights = pd.Series(selected_weights, index=valid_attrs)
                dataframe[category] = dataframe[valid_attrs].mul(weights).sum(axis=1)

            for category in categories:
                max_score = dataframe[category].max()
                if max_score > 0:
                    dataframe[category] = (dataframe[category] / max_score) ** 2 * 100

            dataframe["overall_score"] = dataframe[categories].mean(axis=1)
            dataframe["value_score"] = values.calculate_value_score_vectorized(
                dataframe["overall_score"].to_numpy(),
                dataframe["upper_value_range"].to_numpy(),
                alpha=1,
            )

            min_value_score = dataframe["value_score"].min()
            max_value_score = dataframe["value_score"].max()

            if max_value_score > min_value_score:
                dataframe["value_score"] = (
                    (dataframe["value_score"] - min_value_score)
                    / (max_value_score - min_value_score)
                    * 100
                )
            else:
                dataframe["value_score"] = 100.0

            current_span.add_event("Player score calculation completed")

        except Exception as e:
            current_span.record_exception(e)
            logging.error(f"Error in calculate_player_scores: {e}")
            raise

        finally:
            end_time = time.time()
            total_time = end_time - start_time
            current_span.set_attribute("calculation_duration", total_time)
            current_span.set_attribute("number of players analyzed:", len(dataframe))
            logging.debug(
                "Time taken to calculate player scores: %.4f seconds", total_time
            )

        return dataframe.round(2)


def assign_weak_foot(left_foot_rating, right_foot_rating):
    """
    Assigns a description to the weak foot strength based on the ratings of the left and right foot.

    Args:
        left_foot_rating (str): The rating of the left foot.
        right_foot_rating (str): The rating of the right foot.

    Returns:
        str: The description of the weak foot strength. Possible values are:
            - "Very Strong"
            - "Strong"
            - "Fairly Strong"
            - "Reasonable"
            - "Fairly Weak"
            - "Weak"
            - "Very Weak"
            - "Unknown" if the ratings are not found in the scale.
    """
    scale = {
        "Very Strong": 7,
        "Strong": 6,
        "Fairly Strong": 5,
        "Reasonable": 4,
        "Fairly Weak": 3,
        "Weak": 2,
        "Very Weak": 1,
    }

    left = scale.get(left_foot_rating, 0)
    right = scale.get(right_foot_rating, 0)
    weak_foot_strength = min(left, right)
    for desc, val in scale.items():
        if val == weak_foot_strength:
            return desc

    return "Unknown"
