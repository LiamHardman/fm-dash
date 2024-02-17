import plotly.graph_objects as go
import os
from dash.dependencies import Input, Output, State
from dash.exceptions import PreventUpdate
import dash
from flask import session
import dash_core_components as dcc
import time
import logging
import orjson
import yaml
import pandas as pd
from dash import html, dash_table
from datetime import datetime
from typing import Union
from calc import convert_to_float
from scipy import stats
from opentelemetry import context
from tracer_provider import tracer


def load_config():
    config_path = os.getenv("FMD_CONF_LOCATION", "config.yml")
    with open(config_path) as f:
        return yaml.safe_load(f)


yaml_config = load_config()

axes_map = yaml_config["axes_map"]


attributes = yaml_config["attributes"]
attributes_physical = list(attributes["physical"].values())
attributes_mental = list(attributes["mental"].values())
attributes_technical = list(attributes["technical"].values())


def create_attribute_table(
    selected_row, attributes_physical, attributes_mental, attributes_technical
):
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
            return "#7ec4cf"
        elif score >= 85:
            return "#83fca0"
        elif score >= 80:
            return "#28A228"
        elif score >= 75:
            return "#2F6030"
        elif score >= 70:
            return "#fcdd86"
        elif score >= 65:
            return "#f5a05f"
        else:
            return "#e35d5d"

    def get_value_score_color(score):
        if score >= 60:
            return "#7ec4cf"
        elif score >= 50:
            return "#83fca0"
        elif score >= 45:
            return "#28A228"
        elif score >= 40:
            return "#2F6030"
        elif score >= 30:
            return "#fcdd86"
        elif score >= 20:
            return "#f5a05f"
        else:
            return "#e35d5d"

    def get_average_rating_color(score):
        if score >= 7.6:
            return "#7ec4cf"
        elif score >= 7.3:
            return "#83fca0"
        elif score >= 7.1:
            return "#28A228"
        elif score >= 6.9:
            return "#2F6030"
        elif score >= 6.8:
            return "#fcdd86"
        elif score >= 6.6:
            return "#f5a05f"
        else:
            return "#e35d5d"

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

    table_headers = [
        html.Th("Technical", style=header_style, colSpan="2"),
        html.Th("Mental", style=header_style, colSpan="2"),
        html.Th("Physical", style=header_style, colSpan="2"),
        html.Th("Additional Attributes", style=header_style, colSpan="2"),
    ]

    rows = []
    max_length = max(
        len(attributes_physical),
        len(attributes_mental),
        len(attributes_technical),
        len(additional_attributes),
    )

    for i in range(max_length):
        row = []
        for category in [attributes_technical, attributes_mental, attributes_physical]:
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
                            str(attr_value), style={**cell_style_value, "color": color}
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
            if additional_attr_key in [
                "overall_score",
                "physical_score",
                "mental_score",
                "technical_score",
            ]:
                color = get_score_color(float(additional_attr_value))
            elif additional_attr_key == "value_score":
                color = get_value_score_color(float(additional_attr_value))
            elif additional_attr_key == "Av Rat":
                color = get_average_rating_color(float(additional_attr_value))

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


def register_graph_callbacks(app):
    def generate_hover_text(row):
        """
        Generates the hover text for a row in the table.
        """
        return (
            f"Name: {row['Name']}<br>"
            f"Position: {row['Position']}<br>"
            f"Age: {row['Age']}<br>"
            f"Value Score: {row['value_score']}<br>"
            f"Overall Score: {row['overall_score']}"
        )

    @app.callback(
        [Output("scatter-plot", "figure"), Output("scatter-plot", "style")],
        [
            Input("x-axis-dropdown", "value"),
            Input("y-axis-dropdown", "value"),
            Input("graph-type-dropdown", "value"),
            Input("top-results-slider", "value"),
            Input("table-data-store", "data"),
            Input("analysis-done", "data"),
            Input("slider_height", "value"),
        ],
    )
    def update_graph_style_and_visibility(
        x_axis,
        y_axis,
        graph_type,
        top_results,
        table_data_json,
        analysis_done,
        slider_height,
    ):
        """
        Updates the graph style and visibility based on user inputs and analysis completion.

        This function generates and styles various types of graphs
        (scatter, line, bar, histogram, box violin, area)
        based on selected axes and graph type.
        It also adjusts the graph's visibility and style based on whether the analysis is complete.
        The function is memoized with a session-specific ID.

        Parameters:
        - x_axis (str): The selected x-axis parameter.
        - y_axis (str): The selected y-axis parameter.
        - graph_type (str): The type of graph to display.
        - top_results (int): The number of top results to display.
        - table_data_json (json or dict): The table data in JSON format or as a Python dictionary.
        - analysis_done (dict): A flag indicating whether the analysis is completed.
        - slider_value (int): The value of a slider controlling the graph's height.

        Returns:
        - tuple: A tuple containing the graph figure and its styling information.
        """
        current_ctx = context.get_current()
        with tracer.start_as_current_span(
            "update_graph_style_and_visibility", context=current_ctx
        ) as current_span:
            start_time = time.time()
            style = {"height": "40vh", "display": "none"}

            if analysis_done and analysis_done.get("done", False) and table_data_json:
                table_data = (
                    orjson.loads(table_data_json)
                    if isinstance(table_data_json, str)
                    else table_data_json
                )  # pylint: disable=no-member
                df = pd.DataFrame(table_data)

                top_results = min(top_results, len(df))

                df = df.nlargest(top_results, x_axis)
                current_span.add_event("Updating graph style and visibility")
                if graph_type == "scatter":
                    fig = go.Figure(
                        data=go.Scatter(
                            x=df[x_axis],
                            y=df[y_axis],
                            mode="markers",
                            marker=dict(size=10, opacity=0.5, color="turquoise"),
                            hoverinfo="text",
                            hovertext=df.apply(generate_hover_text, axis=1),
                        )
                    )
                elif graph_type == "line":
                    fig = go.Figure(
                        data=go.Scatter(
                            x=df[x_axis],
                            y=df[y_axis],
                            mode="lines",
                            line=dict(color="turquoise", width=2),
                            hoverinfo="text",
                            hovertext=df.apply(generate_hover_text, axis=1),
                        )
                    )
                elif graph_type == "bar":
                    fig = go.Figure(
                        data=go.Bar(
                            x=df[x_axis],
                            y=df[y_axis],
                            marker_color="turquoise",
                            text=df["Name"],
                            hoverinfo="text",
                            hovertext=df.apply(generate_hover_text, axis=1),
                        )
                    )
                elif graph_type == "histogram":
                    fig = go.Figure(
                        data=go.Histogram(
                            x=df[x_axis],
                            marker_color="turquoise",
                            hoverinfo="text",
                            hovertext=df.apply(generate_hover_text, axis=1),
                        )
                    )
                elif graph_type == "box":
                    fig = go.Figure(
                        data=go.Box(
                            x=df[x_axis],
                            y=df[y_axis],
                            marker_color="turquoise",
                            hoverinfo="text",
                            hovertext=df.apply(generate_hover_text, axis=1),
                        )
                    )
                elif graph_type == "violin":
                    fig = go.Figure(
                        data=go.Violin(
                            x=df[x_axis],
                            y=df[y_axis],
                            line=dict(color="turquoise"),
                            hoverinfo="text",
                            hovertext=df.apply(generate_hover_text, axis=1),
                        )
                    )
                elif graph_type == "area":
                    fig = go.Figure(
                        data=go.Scatter(
                            x=df[x_axis],
                            y=df[y_axis],
                            fill="tozeroy",
                            line=dict(color="turquoise"),
                            hoverinfo="text",
                            hovertext=df.apply(generate_hover_text, axis=1),
                        )
                    )

                readable_x_axis = axes_map.get(x_axis, x_axis)
                readable_y_axis = axes_map.get(y_axis, y_axis)

                fig.update_layout(
                    title=f"{readable_x_axis} vs {readable_y_axis}",
                    xaxis_title=readable_x_axis,
                    yaxis_title=readable_y_axis,
                    paper_bgcolor="#1E293B",
                    font=dict(color="#FFFFFF"),
                    plot_bgcolor="#1E293B",
                )

            if analysis_done and analysis_done.get("done", False):
                height_in_vh = slider_height if slider_height is not None else 40
                style = {
                    "width": "calc(100vw - 20px)",
                    "height": f"{height_in_vh}vh",
                    "display": "block",
                }
            else:
                fig = go.Figure()

            end_time = time.time()
            total_time = end_time - start_time
            logging.debug(
                "Time taken to update graph style and visibility: %.4f seconds",
                total_time,
            )
            current_span.set_attribute("time_taken", total_time)
            current_span.add_event("Updated graph style and visibility")

            return fig, style

    return update_graph_style_and_visibility


def register_update_output_div_callback(app):
    @app.callback(
        Output("results-container", "children"),
        [Input("table-data-store", "data"), Input("analysis-done", "data")],
    )
    def update_output_div(
        table_data_json: Union[str, dict], analysis_done: dict
    ) -> Union[html.Div, dash_table.DataTable]:
        if not table_data_json:
            return html.Div("")

        try:
            table_data = (
                orjson.loads(table_data_json)
                if isinstance(table_data_json, str)
                else table_data_json
            )
        except orjson.JSONDecodeError:
            return html.Div("There was an error processing the table data.")
        if not isinstance(table_data, list) or not all(
            isinstance(record, dict) for record in table_data
        ):
            return html.Div("Invalid table data format.")

        run_all_weights = (
            analysis_done.get("run_all_weights", False) if analysis_done else False
        )

        selected_weights = (
            analysis_done.get("selected_weights", False) if analysis_done else False
        )

        if run_all_weights and selected_weights == run_all_weights:
            columns_to_display = [
                {"name": "Name", "id": "Name"},
                {"name": "Position", "id": "Position"},
                {"name": "Age", "id": "Age"},
                {"name": "Nation", "id": "Nationality"},
                {"name": "Best Role Name", "id": "best_role_name"},
                {"name": "Role Overall", "id": "best_role_overall_score"},
                {"name": "Physical", "id": "best_role_physical_score"},
                {"name": "Technical", "id": "best_role_technical_score"},
                {"name": "Mental", "id": "best_role_mental_score"},
                {"name": "Value Score", "id": "value_score"},
                {"name": "Wage (K p/w)", "id": "Wage"},
                {"name": "Value (M)", "id": "formatted_upper_value_range"},
            ]
            default_sort_column = "Role Overall Score"
        else:
            columns_to_display = [
                {"name": "Name", "id": "Name"},
                {"name": "Position", "id": "Position"},
                {"name": "Age", "id": "Age"},
                {"name": "Nation", "id": "Nationality"},
                {"name": "Overall", "id": "overall_score"},
                {"name": "Physical", "id": "physical_score"},
                {"name": "Mental", "id": "mental_score"},
                {"name": "Technical", "id": "technical_score"},
                {"name": "Value Score", "id": "value_score"},
                {"name": "Wage (K p/w)", "id": "Wage"},
                {"name": "Value (M)", "id": "formatted_upper_value_range"},
            ]
            default_sort_column = "overall_score"
        return dash_table.DataTable(
            id="table-sorting-filtering",
            columns=columns_to_display,
            data=table_data,
            sort_action="native",
            filter_action="native",
            sort_mode="multi",
            sort_by=[{"column_id": default_sort_column, "direction": "desc"}],
            page_action="native",
            page_size=10,
            page_current=0,
            hidden_columns=["find_similar"],
            fixed_rows={"headers": True},
            style_table={
                "width": "100%",
                "minWidth": "100%",
                "maxHeight": "75vh",
                "overflowY": "auto",
            },
            style_cell={
                "minWidth": "30px",
                "width": "auto",
                "maxWidth": "70px",
                "textAlign": "left",
                "color": "#FFFFFF",
                "backgroundColor": "#1E293B",
                "overflow": "hidden",
                "textOverflow": "ellipsis",
            },
            style_header={
                "minWidth": "30px",
                "width": "auto",
                "maxWidth": "70px",
                "backgroundColor": "#233044",
                "color": "#FFFFFF",
                "fontWeight": "bold",
                "overflow": "auto",
                "textOverflow": "ellipsis",
            },
            style_data_conditional=[
                {"if": {"row_index": "odd"}, "backgroundColor": "#1E293B"},
                {"if": {"row_index": "even"}, "backgroundColor": "#233044"},
                {
                    "if": {"column_id": "find_similar"},
                    "color": "transparent",
                    "backgroundColor": "transparent",
                    "cursor": "default",
                },
            ],
            style_cell_conditional=[
                {"if": {"column_id": "Nationality"}, "width": "100px"},
                {"if": {"column_id": "Age"}, "width": "45px"},
                {"if": {"column_id": "Age"}, "width": "150px"},
                {
                    "if": {"column_id": "overall_score"},
                    "width": "auto",
                    "max-width": "180px",
                    "overflow": "ellipsis",
                },
            ],
        )

    return update_output_div


def register_display_similar_players_callback(app):
    @app.callback(
        Output("similar-players-output", "children"),
        [
            Input("table-sorting-filtering", "active_cell"),
            Input("table-sorting-filtering", "page_current"),
            Input("table-sorting-filtering", "page_size"),
            Input("table-sorting-filtering", "derived_virtual_data"),
            Input("table-sorting-filtering", "derived_virtual_indices"),
        ],
        [State("table-sorting-filtering", "data")],
    )
    def display_similar_players(
        active_cell,
        page_current,
        page_size,
        filtered_rows,
        filtered_indices,
        all_rows,
    ):
        """
        Displays similar players based on the selected player in a paginated, filtered table.

        The function is memoized with a session ID for 300 seconds. It calculates the correct index of
        the selected player in a filtered and paginated dataset,
        and displays data for similar players.
        The update is prevented if the required conditions,
        ike an active cell or filtered rows, are not met.

        Parameters:
        - session_id (str): The session ID for memoization.
        - active_cell (dict): The actively selected cell in the table.
        - page_current (int): The current page number in the table.
        - page_size (int): The number of rows per page in the table.
        - filtered_rows (list): The list of rows after filtering.
        - filtered_indices (list): The indices of rows after filtering.
        - all_rows (list): All rows in the table.

        Returns:
        - html.Component or PreventUpdate: The component displaying similar players,
        or a PreventUpdate exception.
        """
        columns_to_display = [
            {"name": "Name", "id": "Name"},
            {"name": "Age", "id": "Age"},
            {"name": "Position", "id": "Position"},
            {"name": "Value (M)", "id": "formatted_upper_value_range"},
            {"name": "Wage (K p/w)", "id": "Wage"},
            {"name": "Overall Score", "id": "overall_score"},
            {"name": "Value Score", "id": "value_score"},
            {"name": "Similarity Score", "id": "similarity_score"},
        ]
        if not active_cell or not filtered_rows or not filtered_indices:
            raise PreventUpdate

        filtered_index_on_page = active_cell["row"]
        page_start = page_current * page_size
        absolute_index = page_start + filtered_index_on_page

        try:
            selected_player_index = filtered_indices[absolute_index]
            selected_player_series = pd.Series(all_rows[selected_player_index])
        except (IndexError, TypeError):
            raise PreventUpdate from None

        df = pd.DataFrame(all_rows)

        all_attributes = attributes_physical + attributes_mental + attributes_technical
        attribute_differences = df[all_attributes].sub(
            selected_player_series[all_attributes], axis="columns"
        )
        similarity_scores = attribute_differences.abs().le(2).sum(axis="columns")

        df["similarity_score"] = similarity_scores

        similarity_threshold = (
            len(all_attributes) * 0.75
        )  # pylint: disable=unused-variable
        value_range_threshold = (  # pylint: disable=unused-variable
            2 * selected_player_series["upper_value_range"]
        )
        similar_players_df = df.query(
            "similarity_score >= @similarity_threshold and upper_value_range <= @value_range_threshold"
        ).nlargest(10, "similarity_score")

        return dash_table.DataTable(
            data=similar_players_df.to_dict("records"),
            columns=columns_to_display,
            style_table={"maxHeight": "300px", "overflowY": "auto"},
            style_cell={
                "minWidth": "30px",
                "width": "70px",
                "maxWidth": "70px",
                "textAlign": "left",
                "color": "#FFFFFF",
                "backgroundColor": "#1E293B",
                "overflow": "auto",
                "textOverflow": "ellipsis",
            },
            style_header={
                "minWidth": "30px",
                "width": "70px",
                "maxWidth": "70px",
                "backgroundColor": "#233044",
                "color": "#FFFFFF",
                "fontWeight": "bold",
                "overflow": "ellipsis",
                "textOverflow": "ellipsis",
            },
            style_data_conditional=[
                {"if": {"row_index": "odd"}, "backgroundColor": "#1E293B"},
                {"if": {"row_index": "even"}, "backgroundColor": "#233044"},
            ],
        )
        return display_similar_players
