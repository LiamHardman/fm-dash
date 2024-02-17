"""
This module defines the layout for the FM Dash app.

The layout includes various components such as header,
footer, upload buttons, export buttons, dropdowns, sliders, and graphs.

Functions:
- get_layout: Returns the layout for the app.

"""

# pylint: skip-file
from dash import html, dcc
from dash_core_components import Download
import dash_bootstrap_components as dbc
import dash_table
import header_footer
from graph_callbacks import create_attribute_table
from fifa_stats import generate_formation_options
from defaults import nations
import app_config

yaml_config = app_config.load_config()
attributes = yaml_config["attributes"]
attributes_physical = list(attributes["physical"].values())
attributes_mental = list(attributes["mental"].values())
attributes_technical = list(attributes["technical"].values())
dropdown_options = generate_formation_options()
nation_options = [
    {"label": f"{country} ({code})", "value": code} for country, code in nations.items()
]
position_options = (
    [
        {"label": "GK", "value": "GK"},
        {"label": "DR", "value": "DR"},
        {"label": "DC", "value": "DC"},
        {"label": "DL", "value": "DL"},
        {"label": "DM", "value": "DM"},
        {"label": "WBR", "value": "WBR"},
        {"label": "WBL", "value": "WBL"},
        {"label": "MC", "value": "MC"},
        {"label": "MR", "value": "MR"},
        {"label": "ML", "value": "ML"},
        {"label": "AMC", "value": "AMC"},
        {"label": "AMR", "value": "AMR"},
        {"label": "AML", "value": "AML"},
        {"label": "STC", "value": "STC"},
    ],
)


def create_attribute_inputs(attributes, style):
    return [
        html.Div(
            [
                html.Label(
                    attr,
                    style={
                        "width": "100px",
                        "marginRight": "5px",
                        "textAlign": "center",
                        "fontSize": "24px",
                        "fontWeight": "bold",
                    },
                ),
                dcc.Input(
                    type="number",
                    id=f"input-{attr}",
                    min=0,
                    max=100,
                    step=1,
                    style={
                        **style,
                        "width": "70px",
                    },
                ),
            ],
            style={"display": "flex", "alignItems": "center", "marginBottom": "10px"},
        )
        for attr in attributes
    ]


def create_weights_modal():
    input_style = {"flex": "2", "margin": "0 20px"}  # Adjust the margin as needed
    column_style = {
        "display": "flex",
        "flexDirection": "column",
        "marginRight": "20px",
        "margin": "0 10px",  # Add this line
    }
    modal_container_style = {
        "padding": "20px",
        "backgroundColor": "#1E293B",
        "color": "#FFFFFF",
        "overflowY": "auto",
        "maxHeight": "90vh",
        "display": "flex",
        "flexDirection": "column",
        "alignItems": "center",
    }

    return dbc.Modal(
        id="weights-modal",
        className="weights-modal",
        children=[
            html.Div(
                [
                    html.H3(
                        "Create New Weight",
                        style={"textAlign": "center", "color": "#FFFFFF"},
                    ),
                    html.P(
                        "Enter weight values from 0 (Least important) to 100 (Most important). "
                        'You can then use these by selecting the Weight in "Open Filters" at the main menu',
                        style={
                            "textAlign": "center",
                            "color": "#FFFFFF",
                            "marginBottom": "20px",
                        },
                    ),
                    dbc.Alert(
                        "No file saved yet",
                        id="save-weights-alert",
                        is_open=False,
                        duration=3000,
                    ),
                    dcc.Input(
                        id="weight-name-input",
                        type="text",
                        placeholder="Enter weight name",
                        style={
                            "textAlign": "center",
                            "marginBottom": "20px",
                            "height": "40px",
                            "fontSize": "24px",
                            "padding": "5px",
                        },
                    ),
                    html.Div(
                        [
                            html.Div(
                                id="button-group-container",
                                style={
                                    "textAlign": "center",
                                    "marginBottom": "10px",
                                    "height": "60vh",
                                    "width": "15vw",
                                    "minWidth": "8vw",
                                    "fontSize": "12px",
                                    "padding": "5px",
                                    "overflowY": "auto",
                                },
                            ),
                            html.Div(
                                [
                                    html.Div(
                                        create_attribute_inputs(
                                            attributes_technical, input_style
                                        ),
                                        style=column_style,
                                    ),
                                    html.Div(
                                        create_attribute_inputs(
                                            attributes_mental, input_style
                                        ),
                                        style=column_style,
                                    ),
                                    html.Div(
                                        create_attribute_inputs(
                                            attributes_physical, input_style
                                        ),
                                        style=column_style,
                                    ),
                                ],
                                style={
                                    "display": "flex",
                                    "justifyContent": "center",
                                    "width": "100%",
                                },
                            ),
                        ],
                        style={
                            "display": "flex",
                            "justifyContent": "center",
                            "width": "100%",
                        },
                    ),
                    html.Div(
                        [
                            html.Button(
                                "Save",
                                id="save-weights-button",
                                style={"marginTop": "20px", "marginRight": "10px"},
                                className="btn btn-secondary",
                            ),
                            html.Button(
                                "Close",
                                id="close-weight-modal-button",
                                style={"marginTop": "20px"},
                                className="btn btn-secondary",
                            ),
                        ],
                        style={
                            "display": "flex",
                            "justifyContent": "center",
                            "width": "100%",
                        },
                    ),
                ],
                style=modal_container_style,
            )
        ],
        is_open=False,
    )


def get_layout(axes_map):
    """
    Returns the layout for the app.
    """
    header = header_footer.create_header()
    footer = header_footer.create_footer()
    pre_upload_text = html.Div(
        "Upload your FM export data to get started. Check out the quick start guide if you haven't already!",  # pylint: disable=line-too-long
        id="pre_upload_text",
        style={"display": "none"},
    )

    error_modal = dbc.Modal(
        [
            dbc.ModalHeader(dbc.ModalTitle("Error")),
            dbc.ModalBody(html.Div(id="error-message")),
            dbc.ModalFooter(dbc.Button("Close", id="close-modal", className="ml-auto")),
        ],
        id="error-modal",
        is_open=False,
    )
    club_error_modal = dbc.Modal(
        [
            dbc.ModalHeader(dbc.ModalTitle("Error")),
            dbc.ModalBody(html.Div(id="club-error-message")),
            dbc.ModalFooter(dbc.Button("Close", id="close-modal", className="ml-auto")),
        ],
        id="club-error-modal",
        is_open=False,
    )

    file_save_modal = html.Div(
        [
            dbc.Modal(
                id="file-save-modal",
                children=[
                    dbc.ModalHeader(
                        dbc.ModalTitle("Save/Export Data"),
                        style={
                            "backgroundColor": "rgba(30, 41, 59, 0.8)",
                            "color": "#FFFFFF",
                        },
                    ),
                    dbc.ModalBody(
                        [
                            dbc.Row(
                                [
                                    dbc.Col(
                                        dbc.DropdownMenu(
                                            label="Export Data",
                                            children=[
                                                dbc.DropdownMenuItem(
                                                    "Export as CSV", id="export-csv-btn"
                                                ),
                                                dbc.DropdownMenuItem(
                                                    "Export as XLSX",
                                                    id="export-xlsx-btn",
                                                ),
                                                dbc.DropdownMenuItem(
                                                    "Export as Markdown",
                                                    id="export-md-btn",
                                                ),
                                                dbc.DropdownMenuItem(
                                                    "Export as HTML",
                                                    id="export-html-btn",
                                                ),
                                            ],
                                            color="primary",
                                        ),
                                        width=3,
                                        style={"paddingRight": "10px"},
                                    ),
                                    dbc.Col(
                                        dbc.Button(
                                            "Save Data",
                                            id="save-button",
                                            className="ms-auto",
                                            n_clicks=0,
                                        ),
                                        width=3,
                                        style={"paddingLeft": "10px"},
                                    ),
                                ],
                                justify="center",
                                align="center",
                                className="mb-3 g-1",
                            ),
                            dbc.Alert(
                                "No file saved yet",
                                id="save-file-alert",
                                is_open=False,
                            ),
                            html.Div(
                                [
                                    html.Span(id="share-id-display"),
                                    dcc.Clipboard(
                                        target_id="share-id-display",
                                        title="copy",
                                        style={
                                            "margin-left": "10px",
                                        },
                                    ),
                                ],
                                style={
                                    "display": "none",
                                    "alignItems": "center",
                                    "justifyContent": "center",
                                    "paddingTop": "10px",
                                },
                                id="clipboard-container",
                            ),
                        ]
                    ),
                    dbc.ModalFooter(
                        [
                            dbc.Alert(
                                "All saved files get deleted after 3 days!",
                                id="file-duration-alert",
                                is_open=True,
                                color="danger",
                            ),
                            dbc.Button(
                                "Close",
                                id="close-save-modal-button",
                                className="ms-auto",
                            ),
                        ],
                        style={
                            "display": "flex",
                            "backgroundColor": "rgba(30, 41, 59, 0.8)",
                        },
                    ),
                ],
                is_open=False,
                className="file-save-modal",
            ),
        ]
    )

    upgrade_finder_modal = html.Div(
        [
            dbc.Modal(
                id="upgrade-finder-modal",
                children=[
                    dbc.ModalHeader(
                        dbc.ModalTitle("Upgrade Finder"),
                        style={
                            "backgroundColor": "rgba(30, 41, 59, 0.8)",
                            "color": "#FFFFFF",
                        },
                    ),
                    dbc.ModalBody(
                        html.Div(
                            [
                                dbc.Row(
                                    [
                                        dbc.Col(
                                            dcc.Dropdown(
                                                id="club-dropdown",
                                                className="filter-dropdown",
                                                placeholder="Select Club",
                                            ),
                                            width=6,
                                        ),
                                        dbc.Col(
                                            dcc.Dropdown(
                                                id="nationality-dropdown",
                                                options=nation_options,
                                                className="filter-dropdown",
                                                placeholder="Select Nationality (Optional)",
                                                multi=True,
                                            ),
                                            width=6,
                                        ),
                                    ],
                                    className="mb-3",
                                ),
                                dbc.Row(
                                    [
                                        dbc.Col(
                                            dcc.Dropdown(
                                                id="position-dropdown",
                                                className="filter-dropdown",
                                                options=[
                                                    {"label": pos, "value": pos}
                                                    for pos in [
                                                        "GK",
                                                        "DR",
                                                        "DC",
                                                        "DL",
                                                        "DM",
                                                        "WBR",
                                                        "WBL",
                                                        "MC",
                                                        "MR",
                                                        "ML",
                                                        "AMC",
                                                        "AMR",
                                                        "AML",
                                                        "ST",
                                                    ]
                                                ],
                                                placeholder="Select Position",
                                            ),
                                            width=6,
                                        ),
                                        dbc.Col(
                                            dcc.Dropdown(
                                                id="role-dropdown",
                                                className="filter-dropdown",
                                                placeholder="Select Role",
                                            ),
                                            width=6,
                                        ),
                                    ],
                                    className="mb-3",
                                ),
                                dbc.Row(
                                    [
                                        dbc.Col(
                                            dcc.Input(
                                                id="age-input",
                                                type="number",
                                                min=14,
                                                max=45,
                                                placeholder="Enter Age",
                                                className="custom-input",
                                            ),
                                            width=6,
                                        ),
                                        dbc.Col(
                                            dcc.Input(
                                                id="max-transfer-value",
                                                type="number",
                                                min=0,
                                                max=300,
                                                step=1,
                                                placeholder="Enter Max Transfer Value (M)",
                                                className="custom-input",
                                            ),
                                            width=6,
                                        ),
                                    ],
                                    className="mb-3",
                                ),
                                dbc.Accordion(
                                    [
                                        dbc.AccordionItem(
                                            html.Div(
                                                id="club-player-table-container",
                                                style={"margin-top": "20px"},
                                            ),
                                            title="View Current Club Players",
                                            item_id="accordion-club-players",
                                        )
                                    ],
                                    start_collapsed=True,
                                    id="accordion-club-players",
                                    style={
                                        "backgroundColor": "rgba(30, 41, 59, 0.8)",
                                        "color": "#FFFFFF",
                                        "display": "none",
                                    },
                                ),
                                html.Div(
                                    id="upgrade-finder-table-container",
                                    style={"margin-top": "20px"},
                                ),
                            ]
                        ),
                    ),
                    dbc.ModalFooter(
                        dbc.Row(
                            [
                                dbc.Col(
                                    html.Button(
                                        "Find Upgrades",
                                        id="find-upgrades-button",
                                        n_clicks=0,
                                        className="upgrade-button",
                                    ),
                                    width=6,
                                ),
                                dbc.Col(
                                    dcc.Input(
                                        id="min-overall-upgrade",
                                        type="number",
                                        min=0,
                                        placeholder="Min. Overall Upgrade (Optional)",
                                        className="custom-input",
                                    ),
                                    width=6,
                                ),
                            ],
                            className="gx-2",
                        ),
                        style={
                            "display": "flex",
                            "backgroundColor": "rgba(30, 41, 59, 0.8)",
                        },
                    ),
                ],
                is_open=False,
                className="finder-modal",
            ),
        ]
    )
    filters_modal = html.Div(
        [
            dbc.Modal(
                id="filters-modal",
                children=[
                    dbc.ModalHeader(
                        dbc.ModalTitle("Filters"),
                        style={
                            "backgroundColor": "rgba(30, 41, 59, 0.8)",
                            "color": "#FFFFFF",
                        },
                    ),
                    dbc.ModalBody(
                        [
                            dbc.Row(
                                [
                                    dbc.Col(
                                        dcc.Dropdown(
                                            id="personality-dropdown",
                                            className="modal-filter-dropdown",
                                            multi=True,
                                            placeholder="Select Personality",
                                        ),
                                        width=6,
                                    ),
                                    dbc.Col(
                                        dcc.Dropdown(
                                            id="media-handling-dropdown",
                                            className="modal-filter-dropdown",
                                            multi=True,
                                            placeholder="Select Media Handling",
                                        ),
                                        width=6,
                                    ),
                                ],
                            ),
                            dbc.Row(
                                [
                                    dbc.Col(
                                        dcc.Dropdown(
                                            id="weak-foot-dropdown",
                                            className="modal-filter-dropdown",
                                            multi=True,
                                            placeholder="Select Weak Foot",
                                        ),
                                        width=6,
                                    ),
                                    dbc.Col(
                                        dcc.Dropdown(
                                            id="position-filter-dropdown",
                                            className="modal-filter-dropdown",
                                            options=[
                                                {"label": "GK", "value": "GK"},
                                                {"label": "DR", "value": "DR"},
                                                {"label": "DC", "value": "DC"},
                                                {"label": "DL", "value": "DL"},
                                                {"label": "DM", "value": "DM"},
                                                {"label": "WBR", "value": "WBR"},
                                                {"label": "WBL", "value": "WBL"},
                                                {"label": "MC", "value": "MC"},
                                                {"label": "MR", "value": "MR"},
                                                {"label": "ML", "value": "ML"},
                                                {"label": "AMC", "value": "AMC"},
                                                {"label": "AMR", "value": "AMR"},
                                                {"label": "AML", "value": "AML"},
                                                {"label": "STC", "value": "STC"},
                                            ],
                                            placeholder="Select Position",
                                            multi=True,
                                        ),
                                        width=6,
                                    ),
                                ],
                            ),
                            dbc.Row(
                                [
                                    dbc.Col(
                                        dcc.Dropdown(
                                            id="weight-set-dropdown",
                                            className="modal-filter-dropdown",
                                            placeholder="Select Role Weighting",
                                        ),
                                        width=6,
                                    ),
                                    dbc.Col(
                                        dcc.Dropdown(
                                            id="nationality-filter-dropdown",
                                            className="modal-filter-dropdown",
                                            options=nation_options,  # Ensure this variable is defined with the options you need
                                            placeholder="Select Nationality (Optional)",
                                            multi=True,
                                        ),
                                        width=6,
                                    ),
                                ],
                            ),
                            dbc.Row(
                                [
                                    dbc.Col(
                                        dcc.Input(
                                            id="max-transfer-fee-input",
                                            className="filter-input",
                                            type="number",
                                            min=0,
                                            max=5000,
                                            step=1,
                                            placeholder="Enter Max Transfer Fee (M)",
                                        ),
                                        width=6,
                                    ),
                                    dbc.Col(
                                        dcc.Input(
                                            id="max-wage-input",
                                            className="filter-input",
                                            type="number",
                                            min=0,
                                            max=30000,
                                            step=1,
                                            placeholder="Enter Max Wage (k p/w)",
                                        ),
                                        width=6,
                                    ),
                                ],
                            ),
                            dbc.Row(
                                [
                                    dbc.Col(
                                        dcc.Input(
                                            id="max-age-input",
                                            className="filter-input",
                                            type="number",
                                            min=0,
                                            max=100,
                                            step=1,
                                            placeholder="Enter Max Age",
                                        ),
                                        width=6,
                                    ),
                                ],
                            ),
                        ],
                    ),
                    dbc.ModalFooter(
                        [
                            html.Button(
                                "Include Good Weak Foot",
                                id="good-weak-foot-button",
                                n_clicks=0,
                                className="upgrade-button",
                            ),
                            html.Button(
                                "Only Good Personalities",
                                id="good-personalities-button",
                                n_clicks=0,
                                className="upgrade-button",
                            ),
                            html.Button(
                                "Only Good Media Handling",
                                id="good-media-button",
                                n_clicks=0,
                                className="upgrade-button",
                            ),
                            html.Button(
                                "Apply Filters",
                                id="apply-filters-button",
                                n_clicks=0,
                                className="upgrade-button",
                            ),
                        ],
                        style={"backgroundColor": "rgba(30, 41, 59, 0.8)"},
                    ),
                ],
                is_open=False,
                className="filters-modal",
            ),
        ]
    )

    export_buttons_layout = html.Div(
        [
            dcc.Store(id="weights-store"),
            html.Button(
                "Create New Weight",
                id="open-weight-modal-button",
                className="export-button",
            ),
            html.Button(
                "Save / Export Data",
                id="open-file-save-modal-button",
                className="export-button",
            ),
            Download(id="download"),
        ],
        className="export-buttons",
    )

    export_buttons_container = html.Div(
        id="export_buttons_container",
        children=[export_buttons_layout],
        style={"display": "none"},
    )

    upload_and_weight_selector_container = html.Div(
        [
            html.Div(
                dcc.Upload(
                    id="upload-data",
                    children=html.Div(
                        [
                            html.Button(
                                "Upload File",
                                id="upload-button",
                                className="upload-button-large",
                                style={"margin": "0"},
                            ),
                        ]
                    ),
                    multiple=False,
                    accept=".html",
                ),
                style={"padding": "10px", "position": "relative"},
            ),
            dcc.Input(
                id="text-input-filename",
                type="text",
                placeholder="Or Enter Share ID",
                n_submit=0,
                style={
                    "position": "absolute",
                    "top": "60%",
                    "left": "50%",
                    "transform": "translate(-50%, -50%)",
                    "width": "35%",
                    "fontSize": "2em",
                    "textAlign": "center",
                },
            ),
            html.Div(
                html.Button(
                    "Open Filters",
                    id="open-filters-button",
                    n_clicks=0,
                    className="upload-button",
                ),
                style={"padding": "10px"},
            ),
            html.Div(
                html.Button(
                    "Upgrade Finder",
                    id="open-upgrade-finder-button",
                    n_clicks=0,
                    className="upload-button",
                ),
                style={"display": "block", "padding": "10px"},
            ),
            html.Div(
                html.Button(
                    "Use Player Attributes as Weights",
                    id="attribute-search-button",
                    className="upload-button",
                ),
                style={"padding": "10px"},
            ),
            html.Div(
                html.Button(
                    "View Player Info",
                    id="detail-view-button",
                    n_clicks=0,
                    className="upload-button",
                ),
                style={"padding": "10px"},
            ),
        ],
        style={
            "display": "flex",
        },
    )

    top_controls_container = html.Div(
        [
            upload_and_weight_selector_container,
            export_buttons_container,
        ],
        style={
            "display": "flex",
            "justify-content": "space-between",
            "align-items": "center",
            "flex-wrap": "wrap",
            "margin-bottom": "5px",
        },
    )

    return html.Div(
        [
            header,
            dcc.Interval(
                id="interval-component",
                interval=5 * 1000,
                n_intervals=0,
                disabled=True,
            ),
            dcc.Store(id="weight-change-trigger"),
            dcc.Store(id="player-weights-store"),
            dcc.Loading(
                id="loading-upload",
                type="graph",
                style={
                    "position": "fixed",
                    "top": "35vh",
                    "left": "50vw",
                    "transform": "translate(-50%, -50%)",
                    "zIndex": 9999,
                },
                children=[
                    dcc.Store(id="session-store"),
                    dcc.Store(id="analysis-done", data={"done": False}),
                ],
            ),
            dcc.Store(id="filter-values-store"),
            html.Div(
                [
                    club_error_modal,
                    error_modal,
                    file_save_modal,
                    pre_upload_text,
                    filters_modal,
                    upgrade_finder_modal,
                    create_weights_modal(),
                    top_controls_container,
                    html.Button(
                        "Analyze", id="analyze-button", style={"display": "none"}
                    ),
                    html.Div(
                        [
                            dcc.Loading(
                                id="loading-scatter-plot",
                                type="graph",
                                children=[
                                    dcc.Graph(
                                        id="scatter-plot",
                                        style={
                                            "height": "400px",
                                            "width": "calc(100vw - 20px)",
                                            "display": "none",
                                        },
                                    )
                                ],
                                style={"width": "100%", "padding": "20px"},
                            ),
                            html.Div(
                                id="controls-row",
                                children=[
                                    html.Div(
                                        [
                                            html.Div(
                                                [
                                                    html.Label(
                                                        "Graph height(%)",
                                                        style={"margin-right": "10px"},
                                                    ),
                                                    dcc.Slider(
                                                        id="slider_height",
                                                        min=40,
                                                        max=100,
                                                        step=10,
                                                        value=40,
                                                        marks={
                                                            i: str(i)
                                                            for i in range(40, 100, 10)
                                                        },
                                                        className="custom-slider",
                                                    ),
                                                ],
                                                style={
                                                    "padding": "10px",
                                                    "flex": "1 0 45%",
                                                },
                                            ),
                                            dbc.Modal(
                                                [
                                                    dbc.ModalBody(
                                                        [
                                                            dcc.Store(
                                                                id="club-image-url-store"
                                                            ),
                                                            html.Img(
                                                                id="club-cards-output",
                                                                style={"height": "50%"},
                                                            ),
                                                        ]
                                                    ),
                                                ],
                                                id="club-render-modal",
                                                className="club-render-modal",
                                                is_open=False,
                                                size="lg",  # Adjust size as needed
                                            ),
                                            dbc.Modal(
                                                [
                                                    dbc.ModalHeader(
                                                        dbc.ModalTitle("Detailed View"),
                                                        style={
                                                            "backgroundColor": "rgba(30, 41, 59, 0.8)",
                                                            "color": "#FFFFFF",
                                                        },
                                                    ),
                                                    dbc.ModalBody(
                                                        [
                                                            dbc.Row(
                                                                [
                                                                    dbc.Col(
                                                                        [
                                                                            dbc.Row(
                                                                                [
                                                                                    dbc.Col(
                                                                                        dbc.ButtonGroup(
                                                                                            id="position-button-group",
                                                                                            style={
                                                                                                "width": "100%"
                                                                                            },
                                                                                        ),
                                                                                        width=7,
                                                                                    ),
                                                                                    dbc.Col(
                                                                                        dcc.Dropdown(
                                                                                            id="formation-dropdown",
                                                                                            options=dropdown_options,
                                                                                            value="4-3-3(2)",
                                                                                            className="fifa-dropdown",
                                                                                        ),
                                                                                        width=2,
                                                                                    ),
                                                                                    dbc.Col(
                                                                                        dcc.Dropdown(
                                                                                            id="card-type-dropdown",
                                                                                            options=[
                                                                                                {
                                                                                                    "label": "Normal Card",
                                                                                                    "value": "normal",
                                                                                                },
                                                                                                {
                                                                                                    "label": "TOTW",
                                                                                                    "value": "totw",
                                                                                                },
                                                                                                {
                                                                                                    "label": "MOTM",
                                                                                                    "value": "motm",
                                                                                                },
                                                                                                {
                                                                                                    "label": "Icon",
                                                                                                    "value": "icon",
                                                                                                },
                                                                                                {
                                                                                                    "label": "Hero",
                                                                                                    "value": "hero",
                                                                                                },
                                                                                                {
                                                                                                    "label": "UCL Hero",
                                                                                                    "value": "ucl_hero",
                                                                                                },
                                                                                                {
                                                                                                    "label": "TOTY",
                                                                                                    "value": "toty",
                                                                                                },
                                                                                                {
                                                                                                    "label": "POTM EPL",
                                                                                                    "value": "potm_epl",
                                                                                                },
                                                                                                {
                                                                                                    "label": "POTM Bundesliga",
                                                                                                    "value": "potm_bundesliga",
                                                                                                },
                                                                                                {
                                                                                                    "label": "POTM Ligue 1",
                                                                                                    "value": "potm_ligue1",
                                                                                                },
                                                                                                {
                                                                                                    "label": "POTM MLS",
                                                                                                    "value": "potm_mls",
                                                                                                },
                                                                                                {
                                                                                                    "label": "POTM Eredivisie",
                                                                                                    "value": "potm_eredivisie",
                                                                                                },
                                                                                                {
                                                                                                    "label": "POTM Serie A",
                                                                                                    "value": "potm_serie_a",
                                                                                                },
                                                                                                {
                                                                                                    "label": "POTM La Liga",
                                                                                                    "value": "potm_laliga",
                                                                                                },
                                                                                                {
                                                                                                    "label": "TOTGS - UCL",
                                                                                                    "value": "totgs_ucl",
                                                                                                },
                                                                                                {
                                                                                                    "label": "TOTGS - UEL",
                                                                                                    "value": "totgs_uel",
                                                                                                },
                                                                                                {
                                                                                                    "label": "TOTGS - UECL",
                                                                                                    "value": "totgs_uecl",
                                                                                                },
                                                                                                {
                                                                                                    "label": "RTTK - UCL",
                                                                                                    "value": "rttk_ucl",
                                                                                                },
                                                                                                {
                                                                                                    "label": "RTTK - UEL",
                                                                                                    "value": "rttk_uel",
                                                                                                },
                                                                                                {
                                                                                                    "label": "RTTK - UECL",
                                                                                                    "value": "rttk_uecl",
                                                                                                },
                                                                                                {
                                                                                                    "label": "FUT Champs",
                                                                                                    "value": "fut_champs",
                                                                                                },
                                                                                            ],
                                                                                            value="normal",
                                                                                            className="fifa-dropdown",
                                                                                        ),
                                                                                        width=2,
                                                                                    ),
                                                                                    dbc.Col(
                                                                                        dbc.Input(
                                                                                            id="num-upgrades",
                                                                                            placeholder="Number of Upgrades",
                                                                                            className="custom-input",
                                                                                            type="number",
                                                                                            min=1,
                                                                                            max=5,
                                                                                            value=1,
                                                                                            step=1,
                                                                                        ),
                                                                                        width=1,
                                                                                    ),
                                                                                    dbc.Tooltip(
                                                                                        "Enter the number of upgrades (1-5)",
                                                                                        target="num-upgrades",
                                                                                        placement="top",
                                                                                    ),
                                                                                    dbc.Tooltip(
                                                                                        "Formation to render team in",
                                                                                        target="formation-dropdown",
                                                                                        placement="top",
                                                                                    ),
                                                                                    dbc.Tooltip(
                                                                                        "Type of card for players (and team)",
                                                                                        target="card-type-dropdown",
                                                                                        placement="top",
                                                                                    ),
                                                                                ],
                                                                            ),
                                                                            dbc.Row(
                                                                                [
                                                                                    dbc.Col(
                                                                                        dcc.Graph(
                                                                                            id="bar-graph-1",
                                                                                            className="custom-graph-container",
                                                                                        ),
                                                                                        width=4,
                                                                                    ),
                                                                                    dbc.Col(
                                                                                        dcc.Graph(
                                                                                            id="bar-graph-2",
                                                                                            className="custom-graph-container",
                                                                                        ),
                                                                                        width=4,
                                                                                    ),
                                                                                    dbc.Col(
                                                                                        dcc.Graph(
                                                                                            id="bar-graph-3",
                                                                                            className="custom-graph-container",
                                                                                        ),
                                                                                        width=4,
                                                                                    ),
                                                                                ],
                                                                            ),
                                                                        ],
                                                                        width=9,
                                                                    ),
                                                                    dbc.Col(
                                                                        children=[
                                                                            dcc.Loading(
                                                                                id="loading",
                                                                                type="graph",
                                                                                children=[
                                                                                    html.Img(
                                                                                        id="fifa-player-image",
                                                                                        style={
                                                                                            "maxWidth": "100%",
                                                                                            "maxHeight": "400px",
                                                                                            "display": "block",
                                                                                            "margin-left": "auto",
                                                                                            "margin-right": "auto",
                                                                                            "margin-top": "-15px",
                                                                                            "margin-bottom": "-15px",
                                                                                        },
                                                                                    ),
                                                                                ],
                                                                            ),
                                                                        ],
                                                                        width=3,
                                                                        className="custom-graph-container",
                                                                    ),
                                                                ]
                                                            ),
                                                            html.Div(
                                                                id="modal-body-content",
                                                                className="custom-table-container",
                                                            ),
                                                        ],
                                                        style={
                                                            "backgroundColor": "rgba(30, 41, 59, 0.8)"
                                                        },
                                                    ),
                                                    dbc.ModalFooter(
                                                        [
                                                            dbc.Row(
                                                                [
                                                                    dbc.Col(
                                                                        [
                                                                            dbc.Checklist(
                                                                                options=[
                                                                                    {
                                                                                        "label": "Render Player Face",
                                                                                        "value": 1,
                                                                                    }
                                                                                ],
                                                                                value=[
                                                                                    0
                                                                                ],  # default is Off
                                                                                id="render-images-checkbox",
                                                                                switch=True,
                                                                            ),
                                                                        ],
                                                                        width=12,
                                                                        className="mb-2",
                                                                        style={
                                                                            "margin-right": "-10px"
                                                                        },
                                                                    ),
                                                                    dbc.Col(
                                                                        [
                                                                            dbc.Checklist(
                                                                                options=[
                                                                                    {
                                                                                        "label": "Render Badges",
                                                                                        "value": 1,
                                                                                    }
                                                                                ],
                                                                                value=[
                                                                                    0
                                                                                ],
                                                                                id="render-club-images-checkbox",
                                                                                switch=True,
                                                                            ),
                                                                        ],
                                                                        width=12,
                                                                        className="mb-2",
                                                                        style={
                                                                            "margin-right": "-10px"
                                                                        },
                                                                    ),
                                                                    dbc.Col(
                                                                        [
                                                                            dbc.Checklist(
                                                                                options=[
                                                                                    {
                                                                                        "label": "Use Alternate Attribute Calculation",
                                                                                        "value": 1,
                                                                                    }
                                                                                ],
                                                                                value=[
                                                                                    0
                                                                                ],
                                                                                id="alt-mapping-checkbox",
                                                                                switch=True,
                                                                            ),
                                                                        ],
                                                                        width=12,
                                                                        className="mb-2",
                                                                        style={
                                                                            "margin-right": "-10px"
                                                                        },
                                                                    ),
                                                                    dbc.Col(
                                                                        [
                                                                            dbc.Checklist(
                                                                                options=[
                                                                                    {
                                                                                        "label": "Render Bench",
                                                                                        "value": 1,
                                                                                    }
                                                                                ],
                                                                                value=[
                                                                                    0
                                                                                ],
                                                                                id="render-subs-checkbox",
                                                                                switch=True,
                                                                            ),
                                                                        ],
                                                                        width=12,
                                                                        className="mb-2",
                                                                        style={
                                                                            "margin-right": "-10px"
                                                                        },
                                                                    ),
                                                                ]
                                                            ),
                                                            html.Div(
                                                                id="club-cards-error",
                                                                className="text-danger",
                                                            ),
                                                            html.Button(
                                                                "Generate Cards for Club",
                                                                id="club-cards-button",
                                                                className="btn btn-secondary",
                                                            ),
                                                            dbc.Button(
                                                                "Previous",
                                                                id="back-button",
                                                                className="btn btn-secondary",
                                                                n_clicks=0,
                                                            ),
                                                            dbc.Button(
                                                                "Next",
                                                                id="forward-button",
                                                                className="btn btn-secondary",
                                                                n_clicks=0,
                                                            ),
                                                            dcc.Loading(
                                                                id="loading-indicator",
                                                                children=html.Div(
                                                                    id="loading-output"
                                                                ),
                                                                type="graph",
                                                                style={
                                                                    "display": "none"
                                                                },
                                                            ),
                                                            dcc.Upload(
                                                                id="player-image-upload",
                                                                children=html.Div(
                                                                    [
                                                                        html.Button(
                                                                            "Upload Player Image",
                                                                            id="upload-player-image-button",
                                                                            className="btn btn-secondary",
                                                                        ),
                                                                    ]
                                                                ),
                                                                multiple=False,
                                                                accept="image/*",
                                                            ),
                                                            dbc.Button(
                                                                "Clear Upload",
                                                                id="clear-upload-button",
                                                                className="btn btn-secondary",
                                                            ),
                                                            html.Button(
                                                                "Close",
                                                                id="close-modal-button",
                                                                className="btn btn-secondary",
                                                                n_clicks=0,
                                                            ),
                                                            dcc.Store(
                                                                id="stored-filename"
                                                            ),
                                                            dcc.Store(
                                                                id="current-index-store",
                                                                storage_type="memory",
                                                            ),
                                                            dcc.Store(
                                                                id="selected-club-store",
                                                                storage_type="memory",
                                                            ),
                                                        ],
                                                        style={
                                                            "backgroundColor": "rgba(30, 41, 59, 0.8)"
                                                        },
                                                    ),
                                                ],
                                                id="detail-modal",
                                                is_open=False,
                                                className="details-modal",
                                            ),
                                            dash_table.DataTable(
                                                id="table-sorting-filtering", data=[{}]
                                            ),
                                            html.Div(
                                                [
                                                    html.Label(
                                                        "Number of Top Results:",
                                                        style={"margin-right": "10px"},
                                                    ),
                                                    dcc.Slider(
                                                        id="top-results-slider",
                                                        min=1,
                                                        max=100,
                                                        value=50,
                                                        marks={
                                                            i: str(i)
                                                            for i in range(1, 101, 10)
                                                        },
                                                        step=1,
                                                        className="custom-slider",
                                                    ),
                                                ],
                                                style={
                                                    "padding": "10px",
                                                    "flex": "1 0 45%",
                                                },
                                            ),
                                        ],
                                        style={
                                            "display": "flex",
                                            "flex-wrap": "nowrap",
                                        },
                                    ),
                                    html.Div(
                                        [
                                            html.Div(
                                                [
                                                    html.Label(
                                                        "Select X-axis:",
                                                        style={"margin-right": "10px"},
                                                    ),
                                                    dcc.Dropdown(
                                                        id="x-axis-dropdown",
                                                        options=[
                                                            {
                                                                "label": readable,
                                                                "value": axis,
                                                            }
                                                            for axis, readable in axes_map.items()
                                                        ],
                                                        value="overall_score",
                                                        className="custom-dropdown",
                                                    ),
                                                ],
                                                style={"width": "100%", "gap": "10px"},
                                            ),
                                            html.Div(
                                                [
                                                    html.Label(
                                                        "Select Y-axis:",
                                                        style={"margin-right": "10px"},
                                                    ),
                                                    dcc.Dropdown(
                                                        id="y-axis-dropdown",
                                                        options=[
                                                            {
                                                                "label": readable,
                                                                "value": axis,
                                                            }
                                                            for axis, readable in axes_map.items()
                                                        ],
                                                        value="upper_value_range",
                                                        className="custom-dropdown",
                                                    ),
                                                ],
                                                style={"width": "100%", "gap": "10px"},
                                            ),
                                            html.Div(
                                                [
                                                    html.Label(
                                                        "Select Graph Type:",
                                                        style={"margin-right": "10px"},
                                                    ),
                                                    dcc.Dropdown(
                                                        id="graph-type-dropdown",
                                                        options=[
                                                            {
                                                                "label": "Scatter Plot",
                                                                "value": "scatter",
                                                            },
                                                            {
                                                                "label": "Line Graph",
                                                                "value": "line",
                                                            },
                                                            {
                                                                "label": "Bar Chart",
                                                                "value": "bar",
                                                            },
                                                            {
                                                                "label": "Histogram",
                                                                "value": "histogram",
                                                            },
                                                            {
                                                                "label": "Box Plot",
                                                                "value": "box",
                                                            },
                                                            {
                                                                "label": "Violin Plot",
                                                                "value": "violin",
                                                            },
                                                            {
                                                                "label": "Area Chart",
                                                                "value": "area",
                                                            },
                                                            {
                                                                "label": "Heatmap",
                                                                "value": "heatmap",
                                                            },
                                                        ],
                                                        value="scatter",
                                                        className="custom-dropdown",
                                                    ),
                                                ],
                                                style={"width": "100%", "gap": "10px"},
                                            ),
                                        ],
                                        style={
                                            "display": "flex",
                                            "flex-wrap": "nowrap",
                                            "width": "100%",
                                            "gap": "10px",
                                        },
                                    ),
                                ],
                                style={
                                    "display": "none",
                                    "flex-wrap": "nowrap",
                                    "width": "100%",
                                    "gap": "10px",
                                },
                            ),
                        ]
                    ),
                    html.Div(id="player-scores-output", style={"display": "none"}),
                    dcc.Store(id="table-data-store"),
                    html.Div(id="results-container", style={"color": "#FFFFFF"}),
                    html.Div(id="value-scores-output"),
                    html.Div(id="similar-players-output"),
                    dbc.Tooltip(
                        "Number of results for each page.",
                        target="page-size-dropdown",  # This should match the id of the Input
                        placement="right",  # Can be 'top', 'bottom', 'left', 'right'
                    ),
                    html.Div(
                        id="page-size-container",
                        children=[
                            dcc.Dropdown(
                                id="page-size-dropdown",
                                options=[
                                    {"label": "10", "value": 10},
                                    {"label": "20", "value": 20},
                                    {"label": "50", "value": 50},
                                ],
                                value=10,
                                clearable=False,
                                style={
                                    "width": "100px",
                                    "minWidth": "100px",
                                },
                            ),
                        ],
                        style={"display": "none", "color": "#FFFFFF"},
                    ),
                    html.Div(id="main-content", children=[]),
                ],
                id="output-container",
                style={
                    "backgroundColor": "#1E293B",
                    "color": "#FFFFFF",
                    "minHeight": "85vh",
                    "flex": "1",
                    "padding": "10px",
                },
            ),
            footer,
        ],
        style={
            "display": "flex",
            "flex-direction": "column",
            "min-height": "100vh",
            "margin": "0",
        },
    )
