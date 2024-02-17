import dash_html_components as html
import dash_core_components as dcc
import header_footer


def home_page_layout():
    return html.Div(
        [
            header_footer.create_header(),
            html.Div(
                className="home-container",
                children=[
                    html.Img(
                        src="/assets/logo.png",
                        style={
                            "display": "block",
                            "margin-left": "auto",
                            "margin-right": "auto",
                            "margin-top": "20px",
                            "margin-bottom": "20px",
                            "max-width": "256px",
                            "max-height": "256px",
                        },
                    ),
                    html.H1(
                        "FM Data Analysis at a Glance. Now Open Source (soon)",
                        className="home-title",
                    ),
                    html.P(
                        "Download the squad & scouting views, get a web export, and get started below. ",
                        className="home-text",
                    ),
                    html.Button(
                        "Get Started",
                        id="start-button",
                        n_clicks=0,
                        className="start-button",
                    ),
                ],
                style={"text-align": "center", "padding": "50px"},
            ),
            header_footer.create_footer(),
        ],
        style={
            "background-color": "#1E293B",
            "color": "#FFFFFF",
            "height": "100vh",
            "display": "flex",
            "flex-direction": "column",
            "justify-content": "space-between",
        },
    )
