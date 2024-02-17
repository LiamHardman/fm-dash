import unittest
import sys
import pandas as pd
from pandas import DataFrame
from pandas.testing import assert_series_equal
from typing import Dict

sys.path.append("../")
from calc import (
    convert_to_float,
    calculate_category_score,
    convert_to_float,
    calculate_player_scores,
    assign_weak_foot,
)


class TestConvertToFloat(unittest.TestCase):
    def test_convert_percentage(self):
        self.assertEqual(convert_to_float("50%"), 50.0)
        self.assertEqual(convert_to_float("75.5%"), 75.5)
        self.assertEqual(convert_to_float("0%"), 0.0)
        self.assertEqual(convert_to_float("100%"), 100.0)

    def test_convert_number(self):
        self.assertEqual(convert_to_float("10"), 10.0)
        self.assertEqual(convert_to_float("3.14"), 3.14)
        self.assertEqual(convert_to_float("-5"), -5.0)
        self.assertEqual(convert_to_float("0"), 0.0)

    def test_convert_invalid_input(self):
        self.assertEqual(convert_to_float("abc"), 0.0)
        self.assertEqual(convert_to_float(""), 0.0)
        self.assertEqual(convert_to_float(None), 0.0)


class TestCalculateCategoryScore(unittest.TestCase):
    def setUp(self):
        self.dataframe = pd.DataFrame(
            {"attr1": [1, 2, 3], "attr2": [4, 5, 6], "attr3": [7, 8, 9]}
        )
        self.weights = {"attr1": 0.5, "attr2": 0.3, "attr3": 0.2}

    def test_calculate_category_score(self):
        expected_result = pd.Series([3.5, 4.9, 6.3])
        result = calculate_category_score(self.dataframe, self.weights)
        assert_series_equal(result, expected_result)


class TestConvertToFloat(unittest.TestCase):
    def test_convert_percentage(self):
        self.assertEqual(convert_to_float("50%"), 50.0)
        self.assertEqual(convert_to_float("75.5%"), 75.5)
        self.assertEqual(convert_to_float("0%"), 0.0)
        self.assertEqual(convert_to_float("100%"), 100.0)

    def test_convert_number(self):
        self.assertEqual(convert_to_float("10"), 10.0)
        self.assertEqual(convert_to_float("3.14"), 3.14)
        self.assertEqual(convert_to_float("-5"), -5.0)
        self.assertEqual(convert_to_float("0"), 0.0)

    def test_convert_invalid_input(self):
        self.assertEqual(convert_to_float("abc"), 0.0)
        self.assertEqual(convert_to_float(""), 0.0)
        self.assertEqual(convert_to_float(None), 0.0)


class TestCalculatePlayerScores(unittest.TestCase):
    def test_calculate_player_scores(self):
        # Create a sample dataframe for testing
        data = {
            "player_id": [1, 2, 3],
            "attributes_physical": [80, 90, 70],
            "attributes_mental": [75, 85, 80],
            "attributes_technical": [85, 80, 90],
            "upper_value_range": [100, 100, 100],
        }
        df = DataFrame(data)

        # Define the expected output dataframe
        expected_output = DataFrame(
            {
                "player_id": [1, 2, 3],
                "attributes_physical": [80, 90, 70],
                "attributes_mental": [75, 85, 80],
                "attributes_technical": [85, 80, 90],
                "upper_value_range": [100, 100, 100],
                "physical_score": [80.0, 90.0, 70.0],
                "mental_score": [75.0, 85.0, 80.0],
                "technical_score": [85.0, 80.0, 90.0],
                "overall_score": [81.67, 88.33, 80.0],
                "value_score": [81.67, 88.33, 80.0],
            }
        )

        # Define the selected weights
        selected_weights = {
            "attributes_physical": 1.0,
            "attributes_mental": 1.0,
            "attributes_technical": 1.0,
        }

        # Call the function to calculate player scores
        result = calculate_player_scores(df, selected_weights)

        # Assert that the result matches the expected output
        self.assertTrue(result.equals(expected_output))


class TestAssignWeakFoot(unittest.TestCase):
    def test_assign_weak_foot(self):
        self.assertEqual(assign_weak_foot("Very Strong", "Strong"), "Strong")
        self.assertEqual(assign_weak_foot("Very Strong", "Weak"), "Weak")
        self.assertEqual(assign_weak_foot("Very Weak", "Very Strong"), "Very Weak")
        self.assertEqual(assign_weak_foot("Unknown", "Unknown"), "Unknown")


if __name__ == "__main__":
    unittest.main()
