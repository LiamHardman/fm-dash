# TODO - Move below code to its own module

import app_config
import os
import json
import uuid
import os
import dash
from dash.dependencies import Input, Output, State
from dash.exceptions import PreventUpdate
from flask import session

import app_config

yaml_config = app_config.load_config()
attributes = yaml_config["attributes"]
attributes_physical = list(attributes["physical"].values())
attributes_mental = list(attributes["mental"].values())
attributes_technical = list(attributes["technical"].values())

# Ensure tmp directory exists
tmp_dir = "tmp"
os.makedirs(tmp_dir, exist_ok=True)


def register_weight_modal_callbacks(app):
    @app.callback(
        Output("weights-modal", "is_open"),
        [
            Input("open-weight-modal-button", "n_clicks"),
            Input("close-weight-modal-button", "n_clicks"),
            Input("save-weights-button", "n_clicks"),
        ],
        [State("weights-modal", "is_open")],
    )
    def toggle_weights_modal(open_clicks, close_clicks, save_clicks, is_open):
        """
        Toggles the visibility of the weights modal based on button clicks.

        Parameters:
        - open_clicks (int): Number of clicks on the "open-weight-modal-button".
        - close_clicks (int): Number of clicks on the "close-weight-modal-button".
        - save_clicks (int): Number of clicks on the "save-weights-button".
        - is_open (bool): Current visibility state of the weights modal.

        Returns:
        - bool: Updated visibility state of the weights modal.
        """
        ctx = dash.callback_context

        if not ctx.triggered:
            return is_open

        button_id = ctx.triggered[0]["prop_id"].split(".")[0]

        if button_id == "open-weight-modal-button" or button_id == "close-modal-button":
            return not is_open

        return is_open

    @app.callback(
        [
            Output("weights-store", "data"),
            Output("save-weights-alert", "children"),
            Output("save-weights-alert", "is_open"),
        ],
        [Input("save-weights-button", "n_clicks")],
        [State("weight-name-input", "value")]
        + [
            State(f"input-{attr}", "value")
            for attr in attributes_physical + attributes_mental + attributes_technical
        ],
    )
    def update_weights_store(n_clicks, weight_name, *values):
        """
        Update the weights store with the provided weight values.

        Parameters:
        - n_clicks (int): The number of times the save-weights-button has been clicked.
        - weight_name (str): The name of the weight.
        - values (tuple): The values of the weight attributes.

        Returns:
        - list: A list containing the updated weights store data, the save-weights-alert message, and the save-weights-alert visibility.
        """
        if not n_clicks:
            raise PreventUpdate
        weights_attributes = dict(
            zip(attributes_physical + attributes_mental + attributes_technical, values)
        )
        formatted_weight_name = f"CUSTOM - {weight_name}"
        session_id = session.get("session_id", str(uuid.uuid4()))
        file_path = os.path.join(tmp_dir, f"{session_id}.json")
        existing_weights = {}
        if os.path.exists(file_path):
            with open(file_path, "r") as file:
                existing_weights = json.load(file)
        existing_weights[formatted_weight_name] = weights_attributes
        try:
            with open(file_path, "w") as file:
                json.dump(existing_weights, file)
            return [f"Data saved to {file_path}", "Weights saved successfully!", True]
        except Exception as e:
            print(f"Error saving file: {e}")
            return [f"Error saving file: {e}", "Error saving weights.", True]
