from dash import Output, Input, State


def register_visibility_callbacks(app):
    @app.callback(
        Output("pre_upload_text", "style"),
        [
            Input("upload-data", "filename"),
            Input("text-input-filename", "n_submit"),
        ],
    )
    def update_pre_upload_text_visibility(uploaded_filename, n_submit):
        """
        Adjusts the visibility of pre-upload text based on whether a file has been uploaded
        or a filename has been submitted.

        Parameters:
        - uploaded_filename (str): The name of the uploaded file.
        - n_submit (int): The number of times enter has been pressed in the text input.

        Returns:
        - dict: A CSS style dict that's applied to pre_upload_text.
        """
        if uploaded_filename or n_submit:
            return {"display": "none"}
        return {
            "display": "block",
            "position": "absolute",
            "top": "40%",
            "left": "50%",
            "transform": "translate(-50%, -50%)",
            "textAlign": "center",
            "fontSize": "2em",
            "width": "80%",
        }

    @app.callback(
        Output("player-image-upload", "contents"),
        Input("clear-upload-button", "n_clicks"),
        prevent_initial_call=True,
    )
    def clear_upload(n_clicks):
        return None

    @app.callback(
        Output("text-input-filename", "style"),
        [
            Input("text-input-filename", "n_submit"),
            Input("upload-data", "filename"),
        ],
    )
    def hide_text_input(n_submit, uploaded_filename):
        if n_submit or uploaded_filename:
            return {"display": "none"}
        return {
            "position": "absolute",
            "top": "60%",
            "left": "50%",
            "transform": "translate(-50%, -50%)",
            "width": "35%",
            "fontSize": "2em",
            "textAlign": "center",
        }

    @app.callback(
        [Output("upload-button", "className"), Output("upload-button", "style")],
        [Input("analysis-done", "data")],
    )
    def update_button_class_and_style(analysis_state):
        if analysis_state and analysis_state.get("done", False):
            return "upload-button", {"display": "block"}
        else:
            return "upload-button-large", {"display": "block"}

    @app.callback(
        Output("graph-selector", "style"),
        [Input("session-store", "data"), Input("analysis-done", "data")],
    )
    def show_graph_dropdown(session_data, analysis_state):
        """
        Shows the graph dropdown if the analysis is done and the session data exists.
        Inputs:
        - session_data (dict): The session data stored in a dcc.Store component.
        - analysis_state (dict): A dictionary containing the analysis state.
        Returns:
        - dict: A CSS style dict that's applied to the graph dropdown.
        """
        return (
            {"color": "#000000", "display": "block"}
            if session_data and analysis_state.get("done", False)
            else {"display": "none"}
        )

    @app.callback(
        [
            Output("attribute-search-button", "style"),
            Output("detail-view-button", "style"),
            Output("open-filters-button", "style"),
            Output("export_buttons_container", "style"),
            Output("open-upgrade-finder-button", "style"),
            Output("page-size-container", "style"),
            Output("controls-row", "style"),
        ],
        [Input("analysis-done", "data")],
    )
    def toggle_visibility(analysis_state):
        """
        Toggle the visibility of various components based on the analysis done state.

        Parameters:
        - analysis_state (dict): The analysis state data.

        Returns:
        - list: A list of styles for each component.
        """
        is_done = analysis_state and analysis_state.get("done", False)

        visible_style = {"display": "block", "padding": "10px"}

        # Apply visibility settings based on analysis done state
        if is_done:
            return [
                visible_style,  # attribute-search-button
                visible_style,  # detail-view-button
                visible_style,  # open-filters-button
                {"display": "block"},  # export_buttons_container
                visible_style,  # open-upgrade-finder-button
                visible_style,  # page-size-container
                {"display": "block", "width": "100%"},  # controls-row
            ]
        else:
            invisible_style = {"display": "none"}
            # Return invisible style for all components when analysis is not done
            return [invisible_style] * 7
