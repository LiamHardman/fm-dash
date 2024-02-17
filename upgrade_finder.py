from dash import Output, Input, State, html, dash_table
from dash.exceptions import PreventUpdate
import load_data

weights_data = load_data.load_weights_data("weights_data.json")
import logging
import pandas as pd
import calc


def register_upgrade_finder_callbacks(app):
    @app.callback(
        Output("upgrade-finder-modal", "is_open"),
        [Input("open-upgrade-finder-button", "n_clicks")],
        [State("upgrade-finder-modal", "is_open")],
    )
    def toggle_upgrade_finder_modal(n_clicks, is_open):
        if n_clicks:
            return not is_open
        return is_open

    @app.callback(
        Output("club-dropdown", "options"),
        [Input("open-upgrade-finder-button", "n_clicks")],  # Triggered by button click
        [State("session-store", "data")],  # Using session-store data as state
    )
    def set_upgrade_finder_club_options(n_clicks, session_data):
        # Check if the button has been clicked at least once and session data exists
        if n_clicks is None or not session_data:
            return []

        df = pd.DataFrame(session_data)
        clubs = df["Club"].unique()
        return [{"label": club, "value": club} for club in clubs if pd.notna(club)]

    def role_matches_position(role, position):
        # Split the role prefix by space and then by slash
        role_prefix = role.split(" ")[0]  # Get the prefix part of the role
        role_parts = role_prefix.split("/")

        # Direct match for non-compound roles (e.g., "DR" matches "DR")
        if role_prefix == position:
            return True

        # Compound match (e.g., "DL" matches "DR/L")
        if len(role_parts) > 1:
            # Handle both left and right positions
            for part in role_parts:
                if part == position or (
                    part[:-1] == position[:-1] and position[-1] in ["R", "L"]
                ):
                    return True

        return False

    # Below code commented out in favour of using static list of nationalities in defaults.py due to the fact that this either broke debug=True (if analysis-done was used as the input) or lowered performance substantially (if using session-store as the input)

    # @app.callback(
    #     Output("nationality-dropdown", "options"),
    #     [Input("analysis-done", "data")],
    #     [State("session-store", "data")],
    # )
    # def set_nationality_options(analysis_done, session_data):
    #     print(
    #         f"Callback triggered: analysis_done: {analysis_done}, session_data: {session_data}"
    #     )

    #     # Only proceed if 'analysis_done' is a dictionary with 'done' equals True and session_data is present
    #     if (
    #         analysis_done
    #         and isinstance(analysis_done, dict)
    #         and analysis_done.get("done")
    #         and session_data
    #     ):
    #         df = pd.DataFrame(session_data)
    #         nationalities = df["Nationality"].unique()
    #         options = [
    #             {"label": nationality, "value": nationality}
    #             for nationality in nationalities
    #             if pd.notna(nationality)
    #         ]
    #         print(f"Options set for nationality-dropdown: {options}")
    #         return options

    #     print("Returning empty list due to unmet conditions")
    #     return []

    def get_roles_for_position(position):
        # Get roles where the position matches the role's prefix
        return [role for role in weights_data if role_matches_position(role, position)]

    @app.callback(
        Output("role-dropdown", "options"), [Input("position-dropdown", "value")]
    )
    def update_role_dropdown(selected_position):
        if not selected_position:
            return []  # Return empty options if no position is selected

        # Get roles for the selected position
        roles = get_roles_for_position(selected_position)

        # Format for dropdown options
        role_options = [{"label": role, "value": role} for role in roles]

        return role_options

    @app.callback(
        [
            Output(
                "upgrade-finder-table-container", "children"
            ),  # Ensure this matches your layout
            Output("club-player-table-container", "children"),
            Output("accordion-club-players", "style"),
        ],
        [Input("find-upgrades-button", "n_clicks")],
        [
            State("club-dropdown", "value"),
            State("position-dropdown", "value"),
            State("role-dropdown", "value"),
            State("min-overall-upgrade", "value"),
            State("max-transfer-value", "value"),
            State("nationality-dropdown", "value"),
            State("age-input", "value"),  # Add the age input state
            State("session-store", "data"),
        ],
        prevent_initial_call=True,
    )
    def find_upgrade_players(
        n_clicks,
        club,
        position,
        role,
        min_overall_upgrade,
        max_transfer_value,
        selected_nationalities,
        age,
        session_data,
    ):
        if n_clicks is None:
            raise PreventUpdate

        df = pd.DataFrame(session_data)
        if min_overall_upgrade is None:
            min_overall_upgrade = 3

        # Filter for the selected position
        position_df = df[df["Position"].str.contains(position, na=False)].copy()
        if position_df.empty:
            logging.warning("No data after filtering by Position")
            return html.Div("No players found for the selected position.")

        # Calculate scores for all players in the selected position
        selected_weights = weights_data.get(role, {})
        position_df = calc.calculate_player_scores(position_df, selected_weights)

        # Define club_players only filtering by position, not by nationality or age
        club_players_df = position_df[
            position_df["Club"].str.strip().str.lower() == club.strip().lower()
        ]

        # Filter the rest of the players by selected nationalities (if any) and age
        other_players_df = position_df[
            position_df["Club"].str.strip().str.lower() != club.strip().lower()
        ]
        if selected_nationalities:
            other_players_df = other_players_df[
                other_players_df["Nationality"].isin(selected_nationalities)
            ]
        if age is not None:
            other_players_df = other_players_df[other_players_df["Age"] < age]

        # Find upgrade candidates among other players
        club_max_overall = club_players_df["overall_score"].max()
        upgrade_candidates = other_players_df[
            other_players_df["overall_score"] >= club_max_overall + min_overall_upgrade
        ]

        # Apply max transfer value filter if specified
        if max_transfer_value is not None:
            upgrade_candidates = upgrade_candidates[
                upgrade_candidates["formatted_upper_value_range"] <= max_transfer_value
            ]

        def create_player_table(dataframe, table_id):
            # Define the table columns
            columns_to_display = [
                {"name": "Name", "id": "Name"},
                {"name": "Position", "id": "Position"},
                {"name": "Age", "id": "Age"},
                {"name": "Nation", "id": "Nationality"},
                {"name": "Overall Score", "id": "overall_score"},
                {"name": "Technical Score", "id": "technical_score"},
                {"name": "Mental Score", "id": "mental_score"},
                {"name": "Physical Score", "id": "physical_score"},
                {"name": "Value Score", "id": "value_score"},
                {"name": "Wage (K p/w)", "id": "Wage"},
                {"name": "Value (M)", "id": "formatted_upper_value_range"},
            ]

            # Create the DataTable with the specified styling
            return dash_table.DataTable(
                id="upgrade-finder-table",
                columns=columns_to_display,
                data=dataframe.to_dict("records"),
                sort_action="native",
                filter_action="native",
                sort_mode="multi",
                sort_by=[{"column_id": "value_score", "direction": "desc"}],
                page_action="native",
                page_size=10,
                style_table={
                    "width": "100%",
                    "minWidth": "100%",
                    "maxHeight": "75vh",
                    "overflowY": "auto",
                },
                style_cell={
                    "minWidth": "30px",
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
                    "whiteSpace": "normal",
                },
                style_data_conditional=[
                    {"if": {"row_index": "odd"}, "backgroundColor": "#1E293B"},
                    {"if": {"row_index": "even"}, "backgroundColor": "#233044"},
                ],
                style_cell_conditional=[
                    {"if": {"column_id": "Name"}, "width": "200px"},
                    {"if": {"column_id": "Age"}, "width": "45px"},
                    {"if": {"column_id": "Wage"}, "width": "150px"},
                    {"if": {"column_id": "Position"}, "width": "200px"},
                    {"if": {"column_id": "overall_score"}, "width": "100px"},
                    {"if": {"column_id": "technical_score"}, "width": "100px"},
                    {"if": {"column_id": "mental_score"}, "width": "100px"},
                    {"if": {"column_id": "physical_score"}, "width": "100px"},
                    {"if": {"column_id": "value_score"}, "width": "100px"},
                    {
                        "if": {"column_id": "formatted_upper_value_range"},
                        "width": "150px",
                    },
                ],
            )

        upgrade_table = create_player_table(upgrade_candidates, "upgrade-finder-table")
        club_players_table = create_player_table(club_players_df, "club-players-table")

        # Show accordion only when button is clicked
        accordion_style = {} if n_clicks and n_clicks > 0 else {"display": "none"}

        return upgrade_table, club_players_table, accordion_style
