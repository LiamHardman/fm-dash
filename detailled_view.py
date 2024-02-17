"""
This module contains all of the callbacks that trigger as a result of pressing "View Player Info" in the UI.
This includes the modal that pops up with the player's stats, as well as the card preview that is generated.

Vars:
position_groups = Used to logically group positions that are expected to have similar performance stats (etc. Tackles/90, xG/90)
"""


from scipy import stats
import dash
from dash import html
from dash.exceptions import PreventUpdate
from dash.dependencies import Input, Output, State, ALL
import plotly.graph_objects as go
from PIL import Image
import base64
import dash_bootstrap_components as dbc
import json
import numpy as np
import io
import time
from flask import session
import uuid
import os
from calc import convert_to_float
from fifa_stats import (
    calculate_overall_score,
    calculate_new_attributes,
    generate_formation_options,
    position_attribute_weights,
    fifa_position_mapping,
    get_base_image_path,
    overlay_fifa_card_stats,
)
import app_config
from opentelemetry import trace, context
from prometheus_client import Counter, Histogram
from opentelemetry.trace.status import Status, StatusCode
from tracer_provider import tracer
import load_data


weights_data = load_data.load_weights_data("weights_data.json")
# Counter for the total number of times the detailed view card creation callback is called
card_preview_creation_calls = Counter(
    "detailed_card_creation_calls",
    "Total number of calls to the detailed card creation callback",
)

# Counter for the number of errors encountered during the card creation process
card_preview_creation_errors = Counter(
    "detailed_card_creation_errors",
    "Total number of errors in the detailed card creation callback",
)
image_file_size_histogram = Histogram(
    "image_file_size_histogram",
    "Histogram of the file sizes for generated images in KB",
    buckets=[10, 50, 100, 500, 1000, 5000, 10000],
)

player_image_path = None
yaml_config = app_config.load_config()
defense_stats = yaml_config["defense_stats"]
midfield_stats = yaml_config["midfield_stats"]
attack_stats = yaml_config["attack_stats"]
attributes = yaml_config["attributes"]
attributes_physical = list(attributes["physical"].values())
attributes_mental = list(attributes["mental"].values())
attributes_technical = list(attributes["technical"].values())
dropdown_options = generate_formation_options()
position_groups = {
    "All": [
        "DC",
        "DL",
        "DR",
        "WBR",
        "WBL",
        "DM",
        "MC",
        "AMC",
        "MR",
        "ML",
        "AMR",
        "AML",
        "ST",
    ],
    "Centre-backs": ["DC"],
    "Full-backs": ["DL", "DR"],
    "Wing-backs": ["WBR", "WBL"],
    "DM's": ["DM"],
    "Centre Mids": ["MC"],
    "Attacking mids": ["AMC"],
    "All central midfielders": ["CM", "MC", "AMC"],
    "Wingers": ["MR", "ML", "AMR", "AML"],
    "Attackers": ["AMR", "AML", "ST"],
    "Strikers": ["STC"],
}

# Ensure tmp directory exists
tmp_dir = "tmp"
os.makedirs(tmp_dir, exist_ok=True)


def register_position_dropdown_callbacks(app):
    @app.callback(
        Output("position-button-group", "children"),
        [
            Input("current-index-store", "modified_timestamp"),
        ],
        [
            State("table-sorting-filtering", "data"),
            State("current-index-store", "data"),
        ],
    )
    def update_button_group(ts, data, current_index_data):
        """
        Update the button group based on the selected player's position.

        Parameters:
        - ts (int): The modified timestamp of the current index store.
        - data (list): The data from the table sorting and filtering.
        - current_index_data (int): The current index data.

        Returns:
        - list: The updated button group.
        """
        if data is None or current_index_data is None:
            return []
        current_index = int(current_index_data) if current_index_data else 0

        try:
            selected_player = data[current_index]
            player_positions = selected_player["Position"].split(", ")
        except (IndexError, TypeError, ValueError):
            return []

        buttons = []
        for group, group_positions in position_groups.items():
            if (
                any(player_pos in group_positions for player_pos in player_positions)
                or group == "All"
            ):
                buttons.append(
                    dbc.Button(
                        group,
                        id={"type": "position-button", "index": group},
                        n_clicks=0,
                        className="me-1",
                    )
                )

        return buttons

    return update_button_group


def create_attribute_table(
    selected_row, attributes_physical, attributes_mental, attributes_technical
):
    """
    Create the attribute table for the selected player.

    Args:
    selected_row = The row of data for the selected player
    attributes_physical = The physical attributes to display
    attributes_mental = The mental attributes to display
    attributes_technical = The technical attributes to display

    Returns:
    table = The attribute table as a Dash HTML Table
    """

    def safe_float_conversion(value):
        try:
            return float(value)
        except ValueError:
            return None

    def get_attribute_color(value):
        if value >= 17:
            return "#7ec4cf"
        elif value >= 14:
            return "#83fca0"
        elif value >= 10:
            return "#fcdd86"
        elif value >= 7:
            return "#f5a05f"
        else:
            return "#e35d5d"

    def get_score_color(score):
        if score >= 90:
            return "#7ec4cf"  # Blue-like
        elif score >= 85:
            return "#83fca0"  # Light Green
        elif score >= 80:
            return "#28A228"  # Green
        elif score >= 75:
            return "#2F6030"  # Dark Green
        elif score >= 70:
            return "#fcdd86"  # Yellow
        elif score >= 65:
            return "#f5a05f"  # Orange
        else:  # <= 60
            return "#e35d5d"  # Red

    def get_value_score_color(score):
        if score >= 80:
            return "#7ec4cf"  # Blue-like
        elif score >= 70:
            return "#83fca0"  # Light Green
        elif score >= 60:
            return "#28A228"  # Green
        elif score >= 50:
            return "#2F6030"  # Dark Green
        elif score >= 40:
            return "#fcdd86"  # Yellow
        elif score >= 30:
            return "#f5a05f"  # Orange
        else:  # <= 30
            return "#e35d5d"  # Red

    def get_average_rating_color(score):
        if score >= 7.6:
            return "#7ec4cf"  # Blue-like
        elif score >= 7.3:
            return "#83fca0"  # Light Green
        elif score >= 7.1:
            return "#28A228"  # Green
        elif score >= 6.9:
            return "#2F6030"  # Dark Green
        elif score >= 6.8:
            return "#fcdd86"  # Yellow
        elif score >= 6.6:
            return "#f5a05f"  # Orange
        else:  # <= 6.4
            return "#e35d5d"  # Red

    additional_attributes = {
        "Name": "Name",
        "Age": "Age",
        "Nationality": "Nationality",
        "Personality": "Personality",
        "Media Handling": "Media Handling",
        "Position": "Position",
        "Club": "Club",
        "Value (M)": "formatted_upper_value_range",
        "Overall Score": "overall_score",
        "Physical Score": "physical_score",
        "Mental Score": "mental_score",
        "Technical Score": "technical_score",
        "Value Score": "value_score",
        "Average Rating": "Av Rat",
    }

    header_style = {
        "backgroundColor": "#2B3A52",
        "color": "white",
        "fontSize": "32px",
        "width": "33.34%",
        "textAlign": "center",
    }

    cell_style_name = {
        "fontSize": "24px",
        "width": "15%",
        "textAlign": "center",
        "overflow": "hidden",
        "whiteSpace": "nowrap",
        "textOverflow": "ellipsis",
    }

    cell_style_value = {
        "fontSize": "24px",
        "width": "10%",
        "textAlign": "center",
        "overflow": "hidden",
        "whiteSpace": "nowrap",
        "textOverflow": "ellipsis",
    }

    section_title_style = {
        "backgroundColor": "#2B3A52",
        "color": "white",
        "fontSize": "32px",
        "textAlign": "center",
        "fontWeight": "bold",
    }
    table_headers = [
        html.Th("Technical", style=header_style, colSpan="2"),
        html.Th("Mental", style=header_style, colSpan="2"),
        html.Th("Physical", style=header_style, colSpan="2"),
        html.Th("Additional Attributes", style=header_style, colSpan="2"),
    ]

    rows = []
    max_length = (
        max(
            len(attributes_physical),
            len(attributes_mental),
            len(attributes_technical),
            len(additional_attributes),
        )
        + 6
    )

    for i in range(max_length):
        row = []
        is_top_roles_title_row = i == len(attributes_physical)
        for category in [attributes_technical, attributes_mental, attributes_physical]:
            if category == attributes_physical:
                if i < len(attributes_physical):
                    attr_long_name = list(category.keys())[i]
                    attr_short_name = category[attr_long_name]
                    attr_value = selected_row.get(attr_short_name, 0)
                    color = get_attribute_color(attr_value)
                    row.extend(
                        [
                            html.Td(attr_long_name, style=cell_style_name),
                            html.Td(
                                str(attr_value),
                                style={**cell_style_value, "color": color},
                            ),
                        ]
                    )
                else:
                    if is_top_roles_title_row:
                        row.extend(
                            [
                                html.Td(
                                    "Top 5 Roles",
                                    style=section_title_style,
                                    colSpan="2",
                                )
                            ]
                        )
                        is_top_roles_title_row = False
                    elif (
                        len(selected_row.get("top_5_role_names", []))
                        == len(selected_row.get("top_5_role_scores", []))
                        and i < len(attributes_physical) + 6
                    ):
                        role_index = i - len(attributes_physical) - 1
                        role_name = selected_row["top_5_role_names"][role_index]
                        role_score = selected_row["top_5_role_scores"][role_index]
                        color = get_attribute_color(role_score)
                        row.extend(
                            [
                                html.Td(role_name, style=cell_style_name),
                                html.Td(
                                    str(role_score),
                                    style={**cell_style_value, "color": color},
                                ),
                            ]
                        )
                    else:
                        row.extend([html.Td(""), html.Td("")])
            else:
                attr_keys = list(category.keys())
                if i < len(attr_keys):
                    attr_long_name = attr_keys[i]
                    attr_short_name = category[attr_long_name]
                    attr_value = selected_row.get(attr_short_name, 0)
                    color = get_attribute_color(attr_value)
                    row.extend(
                        [
                            html.Td(attr_long_name, style=cell_style_name),
                            html.Td(
                                str(attr_value),
                                style={**cell_style_value, "color": color},
                            ),
                        ]
                    )
                else:
                    row.extend([html.Td(""), html.Td("")])

        if i < len(additional_attributes):
            additional_attr_name, additional_attr_key = list(
                additional_attributes.items()
            )[i]
            additional_attr_value = selected_row.get(additional_attr_key, "")
            color = "#FFFFFF"

            float_value = safe_float_conversion(additional_attr_value)
            if float_value is not None:
                if additional_attr_key in [
                    "overall_score",
                    "physical_score",
                    "mental_score",
                    "technical_score",
                ]:
                    color = get_score_color(float_value)
                elif additional_attr_key == "value_score":
                    color = get_value_score_color(float_value)
                elif additional_attr_key == "Av Rat":
                    color = get_average_rating_color(float_value)

            row.extend(
                [
                    html.Td(additional_attr_name, style=cell_style_name),
                    html.Td(
                        str(additional_attr_value),
                        style={**cell_style_value, "color": color},
                    ),
                ]
            )
        else:
            row.extend([html.Td(""), html.Td("")])

        rows.append(html.Tr(row))

    table = html.Table(
        [html.Tr(table_headers)] + rows,
        style={"margin": "auto", "width": "100%", "textAlign": "center"},
    )

    return table


def get_index_from_active_cell(
    active_cell, page_current, page_size, derived_virtual_indices
):
    """
    Calculate the absolute index based on the current page, page size,
    and derived_virtual_indices from the active cell.
    """
    if derived_virtual_indices is not None:
        page_start = page_current * page_size
        absolute_filtered_index = page_start + active_cell["row"]
        try:
            return derived_virtual_indices[absolute_filtered_index]
        except IndexError:
            return 0
    else:
        return active_cell["row"] + (page_current * page_size)


def adjust_index_for_navigation(
    current_index, button_id, data, derived_virtual_indices
):
    """
    Adjust the current index for navigation considering the derived_virtual_indices.
    """
    if derived_virtual_indices:
        sorted_filtered_indices = list(derived_virtual_indices)
        if current_index in sorted_filtered_indices:
            sorted_index = sorted_filtered_indices.index(current_index)
        else:
            sorted_index = 0

        if button_id == "back-button" and sorted_index > 0:
            return sorted_filtered_indices[sorted_index - 1]
        elif (
            button_id == "forward-button"
            and sorted_index < len(sorted_filtered_indices) - 1
        ):
            return sorted_filtered_indices[sorted_index + 1]
        else:
            return current_index

    else:
        if button_id == "back-button" and current_index > 0:
            return current_index - 1
        elif button_id == "forward-button" and current_index < len(data) - 1:
            return current_index + 1
        else:
            return current_index


def get_bar_color(percentile):
    if percentile == 50:
        return "rgb(128,128,128)"
    elif percentile < 50:
        red = 128 + int((50 - percentile) * 2.56)
        green = 128 - int((50 - percentile) * 2.56)
        return f"rgb({red},{green},{green})"
    else:
        green = 128 + int((percentile - 50) * 2.56)
        red = 128 - int((percentile - 50) * 2.56)
        return f"rgb({red},{green},{red})"


def calculate_percentile(data, stat, selected_row):
    player_value = convert_to_float(selected_row.get(stat, np.nan))

    if np.isnan(player_value):
        return np.nan
    non_zero_values = [
        convert_to_float(row.get(stat, 0))
        for row in data
        if convert_to_float(row.get(stat, 0)) > 0
    ]
    if not non_zero_values:
        return np.nan
    percentile = stats.percentileofscore(non_zero_values, player_value, kind="rank")
    return percentile


def create_percentile_bar_chart(stats, percentiles, values, title):
    colors = []
    hover_text = []
    for stat, value, percentile in zip(stats.values(), values, percentiles):
        if np.isnan(percentile):
            continue
        else:
            colors.append(get_bar_color(percentile))
            hover_text.append(f"{stat}: {value}<br>Percentile: {percentile:.1f}%")

    fig = go.Figure(
        data=[
            go.Bar(
                y=list(stats.values()),
                x=percentiles,
                marker_color=colors,
                orientation="h",
                hoverinfo="text",
                hovertext=hover_text,
                text=values,
                textposition="inside",
            )
        ]
    )
    fig.update_layout(
        title=title,
        xaxis_title="Percentile",
        xaxis_range=[0, 100],
        paper_bgcolor="rgba(30, 41, 59, 0.8)",
        font=dict(color="#FFFFFF"),
        plot_bgcolor="rgba(30, 41, 59, 0.8)",
        margin=dict(l=20, r=20, t=40, b=20),
    )
    return fig


def update_graphs(selected_row, data, selected_group):
    """
    Update the graphs based on the selected player's position and the selected position group.

    Args:
    selected_row = The row of data for the selected player
    data = The full dataset
    selected_group = The selected position group

    Returns:
    defense_fig = The defense stats bar chart
    midfield_fig = The midfield stats bar chart
    attack_fig = The attack stats bar chart
    """
    group_positions = position_groups[selected_group]

    data = [
        row
        for row in data
        if "Position" in row
        and any(pos in str(row["Position"]).split(", ") for pos in group_positions)
    ]

    defense_values = [selected_row.get(stat, 0) for stat in defense_stats]
    defense_percentiles = [
        calculate_percentile(data, stat, selected_row) for stat in defense_stats
    ]
    defense_fig = create_percentile_bar_chart(
        defense_stats, defense_percentiles, defense_values, "Defense Stats"
    )

    midfield_values = [selected_row.get(stat, 0) for stat in midfield_stats]
    midfield_percentiles = [
        calculate_percentile(data, stat, selected_row) for stat in midfield_stats
    ]
    midfield_fig = create_percentile_bar_chart(
        midfield_stats, midfield_percentiles, midfield_values, "Midfield Stats"
    )

    attack_values = [selected_row.get(stat, 0) for stat in attack_stats]
    attack_percentiles = [
        calculate_percentile(data, stat, selected_row) for stat in attack_stats
    ]
    attack_fig = create_percentile_bar_chart(
        attack_stats, attack_percentiles, attack_values, "Attack Stats"
    )

    return defense_fig, midfield_fig, attack_fig


def register_detailled_view_callbacks(app):
    @app.callback(
        [
            Output("detail-modal", "is_open"),
            Output("bar-graph-1", "figure"),
            Output("bar-graph-2", "figure"),
            Output("bar-graph-3", "figure"),
            Output("modal-body-content", "children"),
            Output("current-index-store", "data"),
            Output("selected-club-store", "data"),
        ],
        [
            Input("detail-view-button", "n_clicks"),
            Input("close-modal-button", "n_clicks"),
            Input({"type": "position-button", "index": ALL}, "n_clicks"),
            Input("back-button", "n_clicks"),
            Input("forward-button", "n_clicks"),
        ],
        [
            State("table-sorting-filtering", "active_cell"),
            State("table-sorting-filtering", "data"),
            State("table-sorting-filtering", "page_current"),
            State("table-sorting-filtering", "page_size"),
            State("table-sorting-filtering", "derived_virtual_indices"),
            State("current-index-store", "data"),
        ],
    )
    def toggle_detail_modal(
        n1,
        n2,
        button_clicks,
        back_n,
        forward_n,
        active_cell,
        data,
        page_current,
        page_size,
        derived_virtual_indices,
        current_index_data,
    ):
        """
        Toggle the detail modal and update the graphs and attribute table based on the selected player.

        Args:
        n1 = The number of times the detail view button was clicked
        n2 = The number of times the close modal button was clicked
        button_clicks = The number of times any position button was clicked
        back_n = The number of times the back button was clicked
        forward_n = The number of times the forward button was clicked
        active_cell = The active cell in the table
        data = The full dataset
        page_current = The current page of the table
        page_size = The page size of the table
        derived_virtual_indices = The derived virtual indices of the table
        current_index_data = The current index of the table

        Returns:
        is_open = Whether or not the modal is open
        defense_fig = The defense stats bar chart
        midfield_fig = The midfield stats bar chart
        attack_fig = The attack stats bar chart
        attribute_table = The attribute table as a Dash HTML Table
        current_index = The current index of the table

        """

        ctx = dash.callback_context
        attributes_physical = yaml_config["attributes"]["physical"]
        attributes_mental = yaml_config["attributes"]["mental"]
        attributes_technical = yaml_config["attributes"]["technical"]
        selected_group = "All"
        if not ctx.triggered:
            button_id = "No clicks yet"
        else:
            button_id = ctx.triggered[0]["prop_id"].split(".")[0]

        if button_id == "detail-view-button" and active_cell:
            current_index = get_index_from_active_cell(
                active_cell, page_current, page_size, derived_virtual_indices
            )
        elif button_id in ["back-button", "forward-button"]:
            current_index = adjust_index_for_navigation(
                current_index_data, button_id, data, derived_virtual_indices
            )
        else:
            current_index = current_index_data or 0

        try:
            selected_row = data[current_index]
            selected_club = selected_row["Club"]
        except (IndexError, TypeError):
            return (
                False,
                go.Figure(),
                go.Figure(),
                go.Figure(),
                None,
                current_index,
                None,
            )
        defense_fig, midfield_fig, attack_fig = update_graphs(
            selected_row, data, selected_group
        )
        attribute_table = create_attribute_table(
            selected_row, attributes_physical, attributes_mental, attributes_technical
        )

        if button_id in ["detail-view-button", "back-button", "forward-button"]:
            return (
                True,
                defense_fig,
                midfield_fig,
                attack_fig,
                html.Div([attribute_table]),
                current_index,
                selected_club,
            )
        elif button_id == "close-modal-button":
            return (
                False,
                dash.no_update,
                dash.no_update,
                dash.no_update,
                dash.no_update,
                current_index,
                selected_club,
            )
        if "position-button" in button_id:
            button_info = json.loads(button_id)
            selected_group = button_info.get("index")

            defense_fig, midfield_fig, attack_fig = update_graphs(
                selected_row, data, selected_group
            )
            attribute_table = create_attribute_table(
                selected_row,
                attributes_physical,
                attributes_mental,
                attributes_technical,
            )
            return (
                dash.no_update,
                defense_fig,
                midfield_fig,
                attack_fig,
                html.Div([attribute_table]),
                current_index,
                selected_club,
            )

        return (
            dash.no_update,
            dash.no_update,
            dash.no_update,
            dash.no_update,
            dash.no_update,
            current_index,
            selected_club,
        )

    return toggle_detail_modal


def register_detailled_card_creation(app):
    @app.callback(
        [Output("fifa-player-image", "src"), Output("stored-filename", "data")],
        [
            Input("detail-view-button", "n_clicks"),
            Input("back-button", "n_clicks"),
            Input("forward-button", "n_clicks"),
            Input("card-type-dropdown", "value"),
            Input("render-club-images-checkbox", "value"),
            Input("render-images-checkbox", "value"),
            Input("num-upgrades", "value"),
            Input("alt-mapping-checkbox", "value"),
            Input("player-image-upload", "contents"),
        ],
        [
            State("table-sorting-filtering", "active_cell"),
            State("table-sorting-filtering", "data"),
            State("table-sorting-filtering", "page_current"),
            State("table-sorting-filtering", "page_size"),
            State("table-sorting-filtering", "derived_virtual_indices"),
            State("current-index-store", "data"),
        ],
    )
    def detailled_view_card_preview(
        n1,
        back_n,
        forward_n,
        card_type,
        render_club_images,
        render_images,
        num_upgrades,
        use_alt_mapping,
        uploaded_image_contents,
        active_cell,
        data,
        page_current,
        page_size,
        derived_virtual_indices,
        current_index_data,
    ):
        current_ctx = context.get_current()
        with tracer.start_as_current_span(
            "detailled_view_card_preview", context=current_ctx
        ) as span:
            render_images_bool = 1 in render_images
            render_club_images_bool = 1 in render_club_images
            use_alt_mapping_bool = 1 in use_alt_mapping
            uploaded_player_image = None
            if uploaded_image_contents:
                content_type, content_string = uploaded_image_contents.split(",")
                decoded = base64.b64decode(content_string)
                uploaded_player_image = Image.open(io.BytesIO(decoded))
            ctx = dash.callback_context
            trigger_id = (
                ctx.triggered[0]["prop_id"].split(".")[0] if ctx.triggered else ""
            )

            current_index = current_index_data or 0

            if trigger_id in ["back-button", "forward-button"]:
                current_index = adjust_index_for_navigation(
                    current_index, trigger_id, data, derived_virtual_indices
                )
            elif trigger_id == "detail-view-button" and active_cell:
                current_index = get_index_from_active_cell(
                    active_cell, page_current, page_size, derived_virtual_indices
                )

            try:
                card_preview_creation_calls.inc()
                selected_row = data[current_index]

                if "Position" not in selected_row or not selected_row["Position"]:
                    print("Position field is missing or empty in selected_row")
                    return None, None

                player_name = selected_row.get("Name", "Unknown").replace(" ", "_")
                positions = selected_row.get("Position", "").split(",")
                max_overall_score = 0
                primary_position = None

                for pos in positions:
                    selected_row["Position"] = pos.strip()
                    fifa_position = fifa_position_mapping.get(pos.strip(), pos.strip())

                    attribute_weights = position_attribute_weights.get(
                        fifa_position, {}
                    )
                    overall_score = calculate_overall_score(
                        selected_row, attribute_weights
                    )

                    if overall_score > max_overall_score:
                        max_overall_score = overall_score
                        primary_position = fifa_position

                secondary_positions = [
                    p
                    for p in positions
                    if fifa_position_mapping.get(p.strip(), p.strip())
                    != primary_position
                ]

                new_attributes_df, individual_attributes_df = calculate_new_attributes(
                    selected_row, position_attribute_weights, use_alt_mapping_bool
                )
                stats = new_attributes_df.iloc[0].to_dict()
                stats["Overall"] = max_overall_score

                if (
                    card_type == "normal"
                    and max_overall_score >= 75
                    and all(
                        value < 80 for key, value in stats.items() if key != "Overall"
                    )
                ):
                    base_image_path = "./fifa_card_assets/cards/0_gold.png"
                elif (
                    card_type == "normal"
                    and 64 < max_overall_score <= 75
                    and all(
                        value < 70 for key, value in stats.items() if key != "Overall"
                    )
                ):
                    base_image_path = "./fifa_card_assets/cards/0_silver.png"
                elif (
                    card_type == "normal"
                    and max_overall_score <= 64
                    and all(
                        value < 60 for key, value in stats.items() if key != "Overall"
                    )
                ):
                    base_image_path = "./fifa_card_assets/cards/0_bronze.png"
                else:
                    base_image_path = get_base_image_path(max_overall_score, card_type)

                base_image = Image.open(base_image_path)
                individual_stats = individual_attributes_df.iloc[0].to_dict()
                span.set_attribute("card_type", card_type)
                span.set_attribute("render_images", render_images_bool)
                span.set_attribute("render_club_images", render_club_images_bool)
                span.set_attribute("player_name", player_name)

                modified_image = overlay_fifa_card_stats(
                    base_image,
                    selected_row,
                    stats,
                    [primary_position] + secondary_positions,
                    individual_stats,
                    card_type,
                    render_images_bool,
                    render_club_images_bool,
                    num_upgrades,
                    uploaded_player_image,
                )
                unique_id = uuid.uuid4()
                unique_filename = (
                    f"fifa_card_{player_name}_{primary_position}_{unique_id}.webp"
                )
                output_image_path = f"./assets/{unique_filename}"
                modified_image.save(output_image_path, format="WEBP", quality=90)

                timestamp = int(time.time())
                image_url = f"/assets/{unique_filename}?t={timestamp}"

                file_size = os.path.getsize(output_image_path) / 1024  # File size in KB
                file_size_kb = (
                    os.path.getsize(output_image_path) / 1024
                )  # Convert bytes to kilobytes
                image_file_size_histogram.observe(file_size_kb)
                span.set_attribute("image_file_size_kb", file_size)

                return image_url, unique_filename

            except Exception as e:
                card_preview_creation_errors.inc()
                print(f"Error occurred: {e}")
                span.record_exception(e)
                span.set_status(Status(StatusCode.ERROR, "Error occurred in callback"))
                return None, None
