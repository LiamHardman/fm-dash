import dash
from dash import Output, Input, State, ALL
import load_data
import logging
import app_config
import pandas as pd
from flask import session
import dash_bootstrap_components as dbc
import os
import json

yaml_config = app_config.load_config()

weights_data = load_data.load_weights_data("weights_data.json")
good_personalities = yaml_config["good_personalities"]
good_media_handling = yaml_config["good_media_handling"]
good_weak_foot = yaml_config["good_weak_foot"]

attributes = yaml_config["attributes"]
attributes_physical = list(attributes["physical"].values())
attributes_mental = list(attributes["mental"].values())
attributes_technical = list(attributes["technical"].values())


def create_list_group_items(weight_names):
    return [
        dbc.ListGroupItem(
            weight_name,
            id={"type": "weight", "index": weight_name},
            n_clicks=0,
            action=True,
        )
        for weight_name in weight_names
    ]


def apply_filters(df, filter_values):
    filters = []

    if "personality" in filter_values and filter_values["personality"]:
        filters.append(df["Personality"].isin(filter_values["personality"]))

    if "media_handling" in filter_values and filter_values["media_handling"]:
        media_filter = filter_values["media_handling"]
        filters.append(
            df["Media Handling"].apply(
                lambda x: any(
                    media.strip() in [style.strip() for style in x.split(",")]
                    for media in media_filter
                )
            )
        )

    if "weak_foot" in filter_values and filter_values["weak_foot"]:
        filters.append(df["Weak Foot"].isin(filter_values["weak_foot"]))

    if "max_age" in filter_values and filter_values["max_age"] is not None:
        filters.append(df["Age"] <= filter_values["max_age"])

    if "nationality" in filter_values and filter_values["nationality"]:
        filters.append(df["Nationality"].isin(filter_values["nationality"]))

    if "position" in filter_values and filter_values["position"]:
        position_filter = filter_values["position"]
        filters.append(
            df["Position"].apply(
                lambda x: any(pos in str(x).split(", ") for pos in position_filter)
            )
        )

    if (
        "max_transfer_fee" in filter_values
        and filter_values["max_transfer_fee"] is not None
    ):
        max_transfer_fee = filter_values["max_transfer_fee"]
        filters.append(df["formatted_upper_value_range"] <= max_transfer_fee)

    if "max_wage" in filter_values and filter_values["max_wage"] is not None:
        max_wage = filter_values["max_wage"]
        filters.append(df["Wage"] <= max_wage)

    if filters:
        combined_filter = pd.concat(filters, axis=1).all(axis=1)
        return df[combined_filter]
    else:
        return df


def create_button_group(weight_names):
    # Use a list comprehension to create a list of dbc.Button components for each weight
    return dbc.ButtonGroup(
        [
            dbc.Button(
                weight_name,
                id={"type": "weight", "index": weight_name},
                n_clicks=0,
                className="me-1",
            )
            for weight_name in weight_names
        ],
        vertical=True,
        style={"width": "110%"},
    )


def register_filters_callbacks(app):
    @app.callback(
        [
            Output("weight-set-dropdown", "options"),
            Output("button-group-container", "children"),
        ],
        [
            Input("open-filters-button", "n_clicks"),
            Input("open-weight-modal-button", "n_clicks"),
        ],
        prevent_initial_call=True,
    )
    def populate_weight_dropdown(n_clicks1, n_clicks2):
        session_id = session.get("session_id")
        print(f"Session ID: {session_id}")

        if not session_id:
            return [], []
        tmp_dir = "tmp"
        file_path = os.path.join(tmp_dir, f"{session_id}.json")
        print(f"Looking for file: {file_path}")

        custom_weight_names = []
        if os.path.exists(file_path):
            with open(file_path, "r") as file:
                data = json.load(file)
                custom_weight_names.extend(data.keys())
                print(f"Custom weights found: {custom_weight_names}")
            print(f"Custom weights data: {data}")
        else:
            print("Custom weights file not found")

        print(f"Preset weight names: {list(weights_data.keys())}")

        combined_weight_names = custom_weight_names + list(weights_data.keys())
        print(f"Combined weights: {combined_weight_names}")

        options = [{"label": key, "value": key} for key in combined_weight_names]
        button_group = create_button_group(combined_weight_names)
        return options, button_group

    @app.callback(
        Output("filters-modal", "is_open"),
        [
            Input("open-filters-button", "n_clicks"),
            Input("apply-filters-button", "n_clicks"),
        ],
        [State("filters-modal", "is_open")],
    )
    def toggle_filter_modal(open_clicks, apply_clicks, is_open):
        if open_clicks or apply_clicks:
            return not is_open
        return is_open

    @app.callback(
        Output("filter-values-store", "data"),
        [Input("apply-filters-button", "n_clicks")],
        [
            State("personality-dropdown", "value"),
            State("media-handling-dropdown", "value"),
            State("weak-foot-dropdown", "value"),
            State("max-age-input", "value"),
            State("nationality-filter-dropdown", "value"),
            State("position-filter-dropdown", "value"),
            State("max-transfer-fee-input", "value"),
            State("max-wage-input", "value"),
        ],
    )
    def store_filter_values(
        n_clicks,
        personality_values,
        media_values,
        weak_foot_values,
        max_age,
        nationality_values,
        position_values,
        max_transfer_fee,
        max_wage,
    ):
        return {
            "personality": personality_values,
            "media_handling": media_values,
            "weak_foot": weak_foot_values,
            "max_age": max_age,
            "nationality": nationality_values,
            "position": position_values,
            "max_transfer_fee": max_transfer_fee,
            "max_wage": max_wage,
        }

    @app.callback(
        [
            Output("personality-dropdown", "options"),
            Output("media-handling-dropdown", "options"),
            Output("weak-foot-dropdown", "options"),
        ],
        [Input("open-filters-button", "n_clicks")],
        [State("session-store", "data")],
    )
    def update_filter_dropdown_options(n_clicks, session_data):
        if n_clicks is None or not session_data:
            return (
                [],
                [],
                [],
            )

        df = pd.DataFrame(session_data)

        personality_options = [
            {"label": personality, "value": personality}
            for personality in df["Personality"].dropna().unique()
        ]
        media_handling_unique = set()
        for item in df["Media Handling"].dropna().unique():
            media_handling_unique.update([style.strip() for style in item.split(",")])

        media_handling_options = [
            {"label": media, "value": media} for media in media_handling_unique
        ]
        weak_foot_options = [
            {"label": foot, "value": foot} for foot in df["Weak Foot"].dropna().unique()
        ]
        return (
            personality_options,
            media_handling_options,
            weak_foot_options,
        )

    @app.callback(
        [
            Output("personality-dropdown", "value"),
            Output("media-handling-dropdown", "value"),
            Output("weak-foot-dropdown", "value"),
        ],
        [
            Input("good-personalities-button", "n_clicks"),
            Input("good-media-button", "n_clicks"),
            Input("good-weak-foot-button", "n_clicks"),
        ],
        prevent_initial_call=True,
    )
    def set_preset_dropdown_values(click_p, click_m, click_w):
        ctx = dash.callback_context

        if not ctx.triggered:
            return dash.no_update

        button_id = ctx.triggered[0]["prop_id"].split(".")[0]

        if button_id == "good-personalities-button":
            return good_personalities, dash.no_update, dash.no_update
        elif button_id == "good-media-button":
            return dash.no_update, good_media_handling, dash.no_update
        elif button_id == "good-weak-foot-button":
            return dash.no_update, dash.no_update, good_weak_foot

        return dash.no_update, dash.no_update, dash.no_update

    @app.callback(
        [
            Output(f"input-{attr}", "value")
            for attr in attributes_physical + attributes_mental + attributes_technical
        ]
        + [Output("weight-name-input", "value")],
        [
            Input({"type": "weight", "index": ALL}, "n_clicks"),
        ],
        prevent_initial_call=True,
    )
    def update_inputs(data, *args):
        ctx = dash.callback_context

        if not ctx.triggered:
            return dash.no_update
        else:
            input_id = ctx.triggered[0]["prop_id"].split(".")[0]
            n_clicks = ctx.triggered[0]["value"]

        if n_clicks is None:
            return dash.no_update

        print(f"Triggered component ID: {input_id}")

        input_id_dict = json.loads(input_id)

        if (
            "type" in input_id_dict
            and "index" in input_id_dict
            and input_id_dict["type"] == "weight"
        ):
            weight_name = input_id_dict["index"]
            if weight_name in weights_data:
                selected_weights = weights_data[weight_name]
                attributes = (
                    attributes_physical + attributes_mental + attributes_technical
                )
                return [selected_weights[attr] for attr in attributes] + [weight_name]
        else:
            return dash.no_update
