import base64
import io
import logging
import os
import uuid
from datetime import timedelta
from typing import Union


import dash
import dash_bootstrap_components as dbc
import numpy as np
import orjson
import pandas as pd
from dash import dcc, html, Input, Output, State, dash_table
from dash.exceptions import PreventUpdate
from flask import Flask, session
import time
from flask_httpauth import HTTPBasicAuth
from werkzeug.security import generate_password_hash, check_password_hash
from apscheduler.schedulers.background import BackgroundScheduler
import atexit
import json

from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.trace import get_current_span
from opentelemetry.sdk.resources import Resource
from opentelemetry import trace, context
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from prometheus_flask_exporter import PrometheusMetrics
from prometheus_client import Histogram, Counter, generate_latest, CONTENT_TYPE_LATEST
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from tracer_provider import tracer

import app_config
import app_routing
import calc
from graph_callbacks import (
    register_graph_callbacks,
    register_update_output_div_callback,
    register_display_similar_players_callback,
)
from detailled_view import (
    register_detailled_view_callbacks,
    register_position_dropdown_callbacks,
    register_detailled_card_creation,
)
from web_weights_creator import register_weight_modal_callbacks
from fifa_stats import register_fifa_card_callbacks
from app_schedule import (
    delete_old_files,
    delete_old_cache,
    delete_old_cards,
    delete_old_tmp,
)
from data_exporter import generate_download
from control_visibility import register_visibility_callbacks
from upgrade_finder import register_upgrade_finder_callbacks
from filters import register_filters_callbacks, apply_filters
import home_screen
import load_data
import load_markdown


scheduler = BackgroundScheduler()
scheduler.add_job(func=delete_old_files, trigger="interval", hours=3)
scheduler.add_job(func=delete_old_cache, trigger="interval", hours=12)
scheduler.add_job(func=delete_old_cards, trigger="interval", minutes=5)
scheduler.add_job(func=delete_old_tmp, trigger="interval", minutes=30)
scheduler.start()
atexit.register(lambda: scheduler.shutdown())

yaml_config = app_config.load_config()
app_config.setup_logging(yaml_config["log_level"])
upload_max = yaml_config["upload_max"]

attributes = yaml_config["attributes"]
attributes_physical = list(attributes["physical"].values())
attributes_mental = list(attributes["mental"].values())
attributes_technical = list(attributes["technical"].values())

results_to_return = yaml_config["results_to_return"]

app = dash.Dash(
    __name__,
    title="FM-Dash",
    external_stylesheets=[dbc.themes.BOOTSTRAP],
    suppress_callback_exceptions=True,
)
server = app.server
app.server.secret_key = os.urandom(24)

app.server.config["PERMANENT_SESSION"] = True
app.server.config["PERMANENT_SESSION_LIFETIME"] = timedelta(minutes=10)
app.server.config["SESSION_TYPE"] = "filesystem"
callback_duration = Histogram(
    "update_analysis_duration_seconds", "Duration of update analysis callback"
)
players_analyzed_counter = Counter("players_analyzed", "Number of players analyzed")
FlaskInstrumentor().instrument_app(app.server)

auth = HTTPBasicAuth()
users = {"admin": generate_password_hash("secret")}  # TODO - Set this dynamically


@auth.verify_password
def verify_password(username, password):
    if username in users:
        return check_password_hash(users.get(username), password)
    return False


metrics = PrometheusMetrics(app.server, metrics_decorator=auth.login_required)
metrics.info("app_info", "Application info", version="pre-pre-release")

file_size_histogram = Histogram(
    "uploaded_file_size_bytes",
    "Size of uploaded files",
    buckets=[
        10**5,
        5 * 10**5,
        10**6,
        5 * 10**6,
        10**7,
        5 * 10**7,
        10**8,
    ],  # Define suitable buckets
)

# Counter for successful uploads
upload_success_counter = Counter("upload_successes", "Number of successful uploads")

# Counter for failed uploads
upload_failure_counter = Counter(
    "upload_failures", "Number of failed uploads", ["failure_reason"]
)

# Histogram to track processing time of the uploaded files
processing_time_histogram = Histogram(
    "uploaded_file_processing_time_seconds", "Time taken to process uploaded files"
)


LoggingInstrumentor().instrument(set_logging_format=True)


def initialize_session_id():
    if "session_id" not in session:
        session["session_id"] = str(uuid.uuid4())


@app.server.before_request
def before_request():
    initialize_session_id()


register_graph_callbacks(app)
register_detailled_view_callbacks(app)
register_update_output_div_callback(app)
register_display_similar_players_callback(app)
register_upgrade_finder_callbacks(app)
register_filters_callbacks(app)
register_position_dropdown_callbacks(app)
register_fifa_card_callbacks(app)
register_detailled_card_creation(app)
register_visibility_callbacks(app)
register_weight_modal_callbacks(app)


weights_data = load_data.load_weights_data("weights_data.json")
markdown_files = load_markdown.read_markdown_files("./docs")
markdown_layouts = {
    filename: load_markdown.generate_markdown_layout(content)
    for filename, content in markdown_files.items()
}


app.index_string = """
<!DOCTYPE html>
<html>
    <head>
        {%metas%}
        <title>{%title%}</title>
        {%favicon%}
        {%css%}
        <meta name="viewport" content="width=device-width, initial-scale=1">
    </head>
    <body>
        {%app_entry%}
        <footer>
            {%config%}
            {%scripts%}
            {%renderer%}
        </footer>
    </body>
</html>
"""


app.layout = html.Div(
    [
        dcc.Location(id="url", refresh=False),
        html.Div(id="page-content", children=home_screen.home_page_layout()),
    ]
)


@app.callback(Output("page-content", "children"), [Input("url", "pathname")])
def render_page_content(pathname):
    """
    Callback for rendering page content based on the URL pathname.
    """
    return app_routing.get_page_layout(
        pathname, yaml_config, weights_data, markdown_files
    )


@app.callback(
    Output("url", "pathname"),
    [Input("start-button", "n_clicks")],
    prevent_initial_call=True,
)
def route_to_main_app(n_clicks):
    """
    Routes the user to the main application page when the start button is clicked.

    Parameters:
    - n_clicks (int): The number of times the start button has been clicked.

    Returns:
    - str or PreventUpdate: The pathname to redirect to ("/app") or a PreventUpdate exception.
    """
    if n_clicks > 0:
        return "/app"
    else:
        raise PreventUpdate


@app.callback(
    [
        Output("session-store", "data"),
        Output("error-modal", "is_open"),
        Output("error-message", "children"),
    ],
    [
        Input("upload-data", "contents"),
        Input("text-input-filename", "n_submit"),
    ],
    [
        State("upload-data", "filename"),
        State("text-input-filename", "value"),
    ],
)
def parse_upload(contents, n_submit, uploaded_filename, input_filename):
    current_ctx = context.get_current()
    modal_open = False
    error_msg = ""
    df = {}
    ctx = dash.callback_context
    trigger_id = ctx.triggered[0]["prop_id"].split(".")[0]

    with tracer.start_as_current_span("parse_upload", context=current_ctx) as span:
        decoded = None
        if trigger_id == "upload-data" and contents:
            filename = uploaded_filename
            decoded = base64.b64decode(contents.split(",")[1])
        elif trigger_id == "text-input-filename" and input_filename:
            filename = f"{input_filename}.html"
            try:
                with open(f"uploads/{filename}", "rb") as file:
                    decoded = file.read()
            except FileNotFoundError:
                error_msg = f"File named {filename} not found."
                logging.error(error_msg)
                return {}, True, error_msg
        if not decoded:
            span.add_event("No content to process")
            return {}, False, "No file uploaded or filename entered."
        try:
            file_size_mb = len(decoded) / (1024**2)
            file_size_histogram.observe(file_size_mb)
            span.set_attribute("upload.file_size_mb", file_size_mb)
            span.set_attribute("upload.filename", filename)

            df, error_message = load_data.prepare_data(
                io.StringIO(decoded.decode("utf-8"))
            )
            if error_message:
                span.record_exception(error_message)
                upload_failure_counter.labels(
                    failure_reason="data_preparation_error"
                ).inc()
                modal_open = True
                return {}, modal_open, error_message
            df.reset_index(inplace=True, drop=True)
            df["id"] = df.index
            df["find_similar"] = [f"[Find Similar](find:{id})" for id in df.index]
            if df is not None:
                df["Weak Foot"] = df.apply(
                    lambda x: calc.assign_weak_foot(x["Left Foot"], x["Right Foot"]),
                    axis=1,
                )

            span.add_event("File processed successfully")
            return df.to_dict("records"), modal_open, error_msg

        except (ValueError, IOError, pd.errors.ParserError) as e:
            logging.error(f"Failed to process file {filename}: {e}")
            span.record_exception(e)
            upload_failure_counter.labels(failure_reason="processing_error").inc()
            modal_open = True
            error_msg = f"Failed to process file {filename}. Error: {e}"
            return df, modal_open, error_msg


@app.callback(
    [
        Output("save-file-alert", "children"),
        Output("save-file-alert", "is_open"),
        Output("save-button", "disabled"),
        Output("file-save-modal", "is_open"),
        Output("share-id-display", "children"),
        Output("clipboard-container", "style"),
    ],
    [
        Input("save-button", "n_clicks"),
        Input("close-save-modal-button", "n_clicks"),
        Input("open-file-save-modal-button", "n_clicks"),
    ],
    [
        State("upload-data", "contents"),
        State("upload-data", "filename"),
        State("file-save-modal", "is_open"),
    ],
)
def save_upload(
    save_n_clicks, close_n_clicks, open_n_clicks, contents, filename, is_open
):
    current_ctx = context.get_current()
    file_id = str(uuid.uuid4())
    triggered_id = dash.callback_context.triggered[0]["prop_id"].split(".")[0]

    if triggered_id == "save-button" and contents:
        with tracer.start_as_current_span("save_upload", context=current_ctx) as span:
            span.set_attribute("upload.original_filename", filename)
            span.set_attribute("upload.file_id", file_id)

            content_type, content_string = contents.split(",")
            decoded = base64.b64decode(content_string)
            new_filename = f"{file_id}.html"
            file_path = f"uploads/{new_filename}"

            try:
                with open(file_path, "wb") as file:
                    file.write(decoded)
                span.add_event("File saved successfully")
                return (
                    f"File saved! Copy the Share ID below to load this in the future.",
                    True,
                    True,
                    is_open,
                    file_id,
                    {"display": "flex"},
                )
            except Exception as e:
                span.record_exception(e)
                logging.error(f"Failed to save file {new_filename}: {e}")
                return (
                    f"Failed to save file: {str(e)}",
                    True,
                    False,
                    is_open,
                    dash.no_update,
                    {"display": "none"},
                )

    if triggered_id in ["close-save-modal-button", "open-file-save-modal-button"]:
        return (
            dash.no_update,
            dash.no_update,
            False,
            not is_open,
            dash.no_update,
            {"display": "none"} if not is_open else dash.no_update,
        )

    return (
        dash.no_update,
        dash.no_update,
        False,
        is_open,
        dash.no_update,
        {"display": "none"} if not is_open else dash.no_update,
    )


@app.callback(
    Output("player-weights-store", "data"),
    [Input("attribute-search-button", "n_clicks")],
    [
        State("table-sorting-filtering", "page_current"),
        State("table-sorting-filtering", "page_size"),
        State("table-sorting-filtering", "derived_virtual_data"),
        State("table-sorting-filtering", "derived_virtual_indices"),
        State("table-sorting-filtering", "active_cell"),
    ],
)
def update_player_weights(
    n_clicks, page_current, page_size, virtual_data, virtual_indices, active_cell
):
    current_ctx = context.get_current()
    with tracer.start_as_current_span(
        "update_player_weights", context=current_ctx
    ) as current_span:
        start_time = time.time()

        if not n_clicks:
            current_span.add_event("No clicks registered, preventing update")
            raise PreventUpdate

        if not virtual_data or not active_cell:
            current_span.add_event("No data or active cell, preventing update")
            raise PreventUpdate

        try:
            filtered_index_on_page = active_cell["row"]
            page_start = page_current * page_size
            absolute_index = page_start + filtered_index_on_page

            current_span.add_event(f"Calculating absolute index: {absolute_index}")

            try:
                selected_player_index = virtual_indices[absolute_index]
                selected_player_data = virtual_data[selected_player_index]
            except (IndexError, TypeError) as exc:
                current_span.add_event(f"Error accessing player data: {exc}")
                raise PreventUpdate from exc

            player_weights = {
                attr: selected_player_data[attr] * 5
                for attr in attributes_physical
                + attributes_mental
                + attributes_technical
            }

            current_span.add_event("Player weights calculated successfully")
            return player_weights

        except Exception as e:
            logging.error(f"Error during weight calculation: {e}")
            current_span.add_event(f"Error during weight calculation: {e}")
            raise PreventUpdate from e
        finally:
            duration = time.time() - start_time
            current_span.add_event(f"Callback duration: {duration:.2f} seconds")
            logging.info(f"Callback duration: {duration:.2f} seconds")


@app.callback(
    Output("weight-change-trigger", "data"),
    [Input("weight-set-dropdown", "value"), Input("player-weights-store", "data")],
    prevent_initial_call=True,
)
def update_trigger_on_weight_change(selected_weight_set, player_weights):
    """
    Update the trigger data when the weight set dropdown or player weights change.

    Parameters:
    - selected_weight_set (any): The selected weight set from the dropdown.
    - player_weights (any): The data of player weights.

    Returns:
    - dict: A dictionary containing the updated trigger data.
      - "triggered" (bool): Indicates if the trigger was activated.
      - "trigger_id" (str): The ID of the trigger.
      - "selected_weight_set" (any): The selected weight set.
      - "player_weights" (any): The player weights data.
    """
    ctx = dash.callback_context
    trigger_id = ctx.triggered[0]["prop_id"].split(".")[0]
    return {
        "triggered": True,
        "trigger_id": trigger_id,
        "selected_weight_set": selected_weight_set,
        "player_weights": player_weights,
    }


@app.callback(
    [Output("table-data-store", "data"), Output("analysis-done", "data")],
    [
        Input("session-store", "data"),
        Input("weight-change-trigger", "data"),
        Input("filter-values-store", "data"),
    ],
    [State("weight-set-dropdown", "value"), State("player-weights-store", "data")],
)
def update_analysis(
    session_data, weight_trigger, filter_values, dropdown_value, player_weights
):
    """
    Update the analysis based on the given inputs.

    Args:
        session_data (any): The session data.
        weight_trigger (any): The weight trigger data.
        filter_values (any): The filter values data.
        dropdown_value (any): The dropdown value.
        player_weights (any): The player weights data.

    Returns:
        tuple: A tuple containing the updated table data and analysis done data.
    """
    current_ctx = context.get_current()
    with tracer.start_as_current_span(
        "update_analysis", context=current_ctx
    ) as current_span:
        start_time = time.time()

        if not session_data:
            logging.info("No session data available, preventing update.")
            current_span.add_event("No session data available, preventing update")
            raise PreventUpdate

        try:
            df = pd.DataFrame(session_data)

            if filter_values:
                df = apply_filters(df, filter_values)

            selected_weights = None
            run_all_weights = True

            if weight_trigger and weight_trigger.get("player_weights"):
                selected_weights = weight_trigger["player_weights"]
                run_all_weights = False
            elif dropdown_value:
                if dropdown_value.startswith("CUSTOM - "):
                    tmp_dir = "tmp"
                    session_id = session.get("session_id")
                    file_path = os.path.join(tmp_dir, f"{session_id}.json")

                    if os.path.exists(file_path):
                        with open(file_path, "r") as file:
                            custom_weights_data = json.load(file)
                            selected_weights = custom_weights_data.get(dropdown_value)

                    run_all_weights = False
                elif dropdown_value in weights_data:
                    selected_weights = weights_data[dropdown_value]
                    run_all_weights = False
            run_all_weights = True
            # TODO - Fix logic around flipping run_all_weights to True or False, the above run_all_weights = True shouldn't be necessary!

            if not selected_weights:
                selected_weights = run_all_weights

            if run_all_weights:
                roles_data = {
                    weight_name: calc.calculate_player_scores(df, weights)
                    for weight_name, weights in weights_data.items()
                }

                # Precompute the max overall score and the corresponding role for each index
                max_scores = pd.DataFrame(
                    {role: data["overall_score"] for role, data in roles_data.items()}
                )
                df["best_role_name"] = max_scores.idxmax(axis=1)
                best_roles = max_scores.max(axis=1)

                # Retrieve best role scores for each category
                for score_type in [
                    "overall_score",
                    "physical_score",
                    "technical_score",
                    "mental_score",
                ]:
                    df[f"best_role_{score_type}"] = np.nan
                    for role in roles_data:
                        mask = df["best_role_name"] == role
                        df.loc[mask, f"best_role_{score_type}"] = roles_data[role].loc[
                            mask, score_type
                        ]

                # Top 5 roles
                top_5_roles = max_scores.apply(
                    lambda x: list(x.nlargest(5).index), axis=1
                )
                top_5_scores = max_scores.apply(lambda x: list(x.nlargest(5)), axis=1)
                df["top_5_role_names"] = top_5_roles
                df["top_5_role_scores"] = top_5_scores

            else:
                df = calc.calculate_player_scores(df, selected_weights)

            # Outputting column names and their data types to a text file
            column_info = df.dtypes

            with open("column_info.txt", "w") as file:
                for column_name, data_type in column_info.items():
                    file.write(f"{column_name}: {data_type}\n")

            df = calc.calculate_player_scores(df, selected_weights)
            df = df.nlargest(results_to_return, "overall_score")
            duration = time.time() - start_time
            num_rows = len(df)

            callback_duration.observe(duration)
            players_analyzed_counter.inc(num_rows)

            logging.info(f"Callback latency: {duration:.2f} seconds")
            logging.info(f"Number of rows analyzed: {num_rows}")

            current_span.add_event(f"Callback latency: {duration:.2f} seconds")
            current_span.add_event(f"Number of rows analyzed: {num_rows}")

            return df.to_dict("records"), {
                "done": True,
                "run_all_weights": run_all_weights,
                "selected_weights": selected_weights,
            }

        except Exception as e:
            logging.error(f"Error during analysis: {e}")
            current_span.add_event(f"Error during analysis: {e}")
            raise PreventUpdate from e
        finally:
            duration = time.time() - start_time
            callback_duration.observe(duration)


@app.callback(
    Output("table-sorting-filtering", "page_size"),
    [Input("page-size-dropdown", "value")],
)
def update_table_page_size(
    page_size_value,
):  # pylint: disable=missing-function-docstring
    return page_size_value


@app.callback(
    [Output("top-results-slider", "max"), Output("top-results-slider", "marks")],
    [Input("table-data-store", "data")],
)
def set_slider_max_and_marks(table_data_json):
    """
    Sets the maximum value and marks for a slider based on the size of the table data.

    This callback function reads the table data and changes the table height slider's max,
    and its corresponding marks. The maximum value is set to the length of the table data,
    and marks are generated for easy navigation. If no data is available, default values are set.

    Parameters:
    - table_data_json (json or dict): The table data in JSON format or as a Python dictionary.

    Returns:
    - tuple: A tuple containing the maximum value for the slider and a dictionary of marks.
    """
    if table_data_json:
        table_data = (
            orjson.loads(table_data_json)  # pylint: disable=maybe-no-member
            if isinstance(table_data_json, str)
            else table_data_json
        )
        max_value = len(table_data)
        marks = {
            i: {"label": str(i)}
            for i in range(1, max_value + 1, max(1, max_value // 10))
        }
        return max_value, marks
    else:
        return 10, {i: {"label": str(i)} for i in range(1, 11)}


@app.callback(
    Output("download", "data"),
    [
        Input("export-csv-btn", "n_clicks"),
        Input("export-xlsx-btn", "n_clicks"),
        Input("export-md-btn", "n_clicks"),
        Input("export-html-btn", "n_clicks"),
    ],
    [State("table-data-store", "data")],
    prevent_initial_call=True,
)
def callback_wrapper(*args, **kwargs):
    return generate_download(*args, **kwargs)


if __name__ == "__main__":
    host = os.getenv("FLASK_RUN_HOST", "127.0.0.1")
    port = int(os.getenv("FLASK_RUN_PORT", "5000"))
    app.run_server(host=host, port=port, debug=False)
