"""
This module provides functionality for loading and processing player data.
"""
import logging
import json
from typing import Dict, Any
import io
from pandas import DataFrame
from values import convert_transfer_fee_to_number, convert_wage_to_weekly
import time
import logging
import pandas as pd
import numpy as np
import yaml
import os
import multiprocessing
from concurrent.futures import ThreadPoolExecutor
from prometheus_client import Counter
from tracer_provider import tracer
from opentelemetry import context


rows_processed_counter = Counter(
    "rows_processed", "Total number of rows processed by prepare_data function"
)


def load_config():
    config_path = os.getenv("FMD_CONF_LOCATION", "config.yml")
    with open(config_path) as f:
        return yaml.safe_load(f)


yaml_config = load_config()
rows_per_chunk = yaml_config["rows_per_chunk"]
threads_count = multiprocessing.cpu_count()
string_columns = ["Personality", "Media Handling", "Weak Foot"]


def rename_first_nat(df, original, new):
    col_list = df.columns.tolist()
    for i, col_name in enumerate(col_list):
        if col_name == original:
            col_list[i] = new
            break
    df.columns = col_list


def replace_int_with_string(value):
    if isinstance(value, int):
        return ""  # Replace integer with empty string
    return value


def custom_html_parser(html_chunk, headers=None):
    data = []
    rows = html_chunk.split("</tr>")
    column_count = len(headers) if headers else None
    headers_processed = False if headers is None else True
    for row in rows:
        cells = [cell.split(">")[1] for cell in row.split("</td>") if "<td" in cell]
        if not headers_processed:
            headers = [
                cell.split(">")[1] for cell in row.split("</th>") if "<th" in cell
            ]
            if headers:
                column_count = len(headers)
                headers_processed = True
                continue
        if cells and len(cells) == column_count:
            data.append(cells)

    if headers_processed and data:
        df = pd.DataFrame(data, columns=headers)
        logging.info(
            f"Parsed DataFrame with {len(df)} rows and {len(df.columns)} columns."
        )
        return df
    else:
        logging.warning(
            f"Failed to parse DataFrame. Headers processed: {headers_processed}, Data length: {len(data)}"
        )
        return pd.DataFrame()


def split_into_chunks(file_content, rows_per_chunk):
    rows = file_content.split('<tr bgcolor="#EEEEEE">')[1:]

    chunks = [
        '<tr bgcolor="#EEEEEE">'
        + '<tr bgcolor="#EEEEEE">'.join(rows[i : i + rows_per_chunk])
        for i in range(0, len(rows), rows_per_chunk)
    ]

    return chunks


def extract_headers(first_chunk):
    start = first_chunk.find("<tr")
    end = first_chunk.find("</tr>", start) + 5
    header_row = first_chunk[start:end]

    return header_row


def process_chunk(html_chunk, headers):
    current_ctx = context.get_current()
    with tracer.start_as_current_span("process_chunk", current_ctx) as span:
        span.set_attribute("process_chunk.size", len(html_chunk))
    logging.info(f"Processing chunk with size: {len(html_chunk)}")
    parsed_df = custom_html_parser(html_chunk, headers)
    if parsed_df.empty:
        logging.warning("Parsed DataFrame is empty.")
    return parsed_df


def prepare_data(file_path, rows_per_chunk=rows_per_chunk):
    with tracer.start_as_current_span("prepare_data") as span:
        span.add_event("Start processing new file")
        logging.info("Processing a new file.")
        start_time = time.time()

        file_content = ""
        if isinstance(file_path, io.StringIO):
            file_content = file_path.getvalue()
            span.add_event("Read content from StringIO")
        else:
            with open(file_path, "rb", encoding="utf8") as file:
                file_content = file.read()
                span.add_event("Read content from file")
        chunks = split_into_chunks(file_content, rows_per_chunk)
        span.add_event("File content split into chunks")

        if not chunks:
            logging.error("No chunks were created from the file content.")
            span.add_event("No chunks created")
            return pd.DataFrame(), "No chunks created"

        first_chunk_data = custom_html_parser(chunks[0])
        span.add_event("First chunk processed")

        headers = (
            first_chunk_data.columns.tolist() if not first_chunk_data.empty else None
        )
        if headers is None:
            logging.error("Failed to extract headers from the first chunk.")
            span.add_event("Headers extraction failed")
            return pd.DataFrame(), "Headers extraction failed"

        span.add_event("Headers extracted")

        with ThreadPoolExecutor(max_workers=threads_count) as executor:
            span.add_event("Starting multithreaded chunk processing")
            futures = [
                executor.submit(process_chunk, chunk, headers) for chunk in chunks[1:]
            ]
            chunked_dfs = [first_chunk_data] + [future.result() for future in futures]
            span.add_event("Multithreaded chunk processing completed")

        try:
            combined_df = pd.concat(chunked_dfs, ignore_index=True)
            span.add_event("DataFrames concatenated")
        except Exception as e:
            logging.error(f"Error during concatenation: {e}")
            span.record_exception(e)
            raise

        rename_first_nat(combined_df, "Nat", "Nationality")
        numeric_columns = [
            "Ch C/90",
            "Mins/Gl",
            "Av Rat",
            "Asts/90",
            "Clr/90",
            "Blk/90",
            "Cr C/90",
            "Drb/90",
            "xA/90",
            "xG/90",
            "Gls/90",
            "Hdrs W/90",
            "Int/90",
            "K Ps/90",
            "NP-xG/90",
            "Ps C/90",
            "Poss Won/90",
            "Shot/90",
            "Tck/90",
        ]
        for col in numeric_columns:
            if col in combined_df.columns:
                combined_df[col] = pd.to_numeric(combined_df[col], errors="coerce")
        rows_processed_counter.inc(len(combined_df))

        int8_columns = [
            "Acc",
            "Agi",
            "Bal",
            "Jum",
            "Nat",
            "Pac",
            "Sta",
            "Str",
            "Agg",
            "Ant",
            "Bra",
            "Cmp",
            "Cnt",
            "Dec",
            "Det",
            "Fla",
            "OtB",
            "Ldr",
            "Pos",
            "Tea",
            "Vis",
            "Wor",
            "Cor",
            "Cro",
            "Dri",
            "Fin",
            "Fir",
            "Fre",
            "Hea",
            "Lon",
            "L Th",
            "Mar",
            "Pas",
            "Pen",
            "Tck",
            "Tec",
            "Age",
        ]

        def calculate_midpoint(value):
            try:
                if pd.isna(value) or value == "":
                    return np.nan
                if "-" in str(value):
                    numbers = str(value).split("-")
                    return (float(numbers[0]) + float(numbers[1])) / 2
                else:
                    return float(value)
            except ValueError:
                return np.nan

        for col in int8_columns:
            if col in combined_df.columns:
                combined_df[col] = combined_df[col].apply(calculate_midpoint)
                combined_df[col] = pd.to_numeric(combined_df[col], errors="coerce")
                combined_df[col] = combined_df[col].fillna(0)
                combined_df[col] = combined_df[col].astype(np.int8)

        try:
            combined_df["lower_value_range"], combined_df["upper_value_range"] = zip(
                *combined_df["Transfer Value"].apply(convert_transfer_fee_to_number)
            )
            combined_df["Wage"] = combined_df["Wage"].apply(convert_wage_to_weekly)
        except KeyError:
            column_name = "Transfer Value"
            error_message = f"Column '{column_name}' not found in the data."
            logging.error(error_message)
            return pd.DataFrame(), error_message

        combined_df["Position"] = combined_df["Position"].apply(process_position_field)
        combined_df["formatted_upper_value_range"] = (
            combined_df["upper_value_range"] / 1_000_000
        )

        def replace_empty_or_hyphen(value):
            if pd.isna(value) or value == "" or value == "-":
                return 0
            else:
                return value

        for col in combined_df.columns:
            combined_df[col] = combined_df[col].apply(replace_empty_or_hyphen)

        for col in string_columns:
            if col in combined_df.columns:
                combined_df[col] = combined_df[col].apply(replace_int_with_string)
        combined_df = combined_df[combined_df["Name"] != "0"]
        num_rows = len(combined_df)
        end_time = time.time()
        total_time = end_time - start_time
        rows_per_second = num_rows / total_time if total_time > 0 else 0
        logging.debug(
            f"Total time taken: {total_time:.2f} seconds, Rows per second: {rows_per_second:.2f}"
        )
        span.set_attribute("prepare_data.time_taken", total_time)
        span.add_event(f"Number of rows processed: {num_rows}")
        span.add_event(f"File processing completed in {total_time:.2f} seconds")

        return combined_df, ""


def process_position_field(position: str) -> str:
    """
    Processes the 'Position' field to expand and normalize position abbreviations.

    Args:
        position (str): The position field to be processed.

    Returns:
        str: The processed position field with expanded and normalized abbreviations.
    """
    start_time = time.time()

    positions = position.split(",")

    expanded_positions = set()
    for pos in positions:
        pos = pos.strip()
        parts = pos.split("/")
        bracket_content = ""
        if "(" in parts[-1]:
            parts[-1], bracket_content = parts[-1].split("(")
            bracket_content = bracket_content.replace(")", "")

        if bracket_content:
            expanded_positions.update(
                [
                    part.strip() + sub_position
                    for part in parts
                    for sub_position in bracket_content
                ]
            )
        else:
            expanded_positions.update(parts)

    end_time = time.time()
    total_time = end_time - start_time

    return ", ".join(sorted(expanded_positions))


def load_weights_data(filepath: str) -> Dict[str, Any]:
    """
    Loads weight data from a JSON file.
    """
    with open(filepath, encoding="utf-8") as file:
        return json.load(file)
