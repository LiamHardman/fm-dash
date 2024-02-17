import dash_html_components as html


def create_header():
    return html.Header(
        className="home-header",
        children=[
            html.Div(
                className="header-links",
                children=[
                    html.A(
                        "Home",
                        href="/",
                        className="header-link",
                        style={"margin-right": "10px"},
                    ),
                    html.A(
                        "Quick Start Guide",
                        href="/quick-start",
                        className="header-link",
                        style={"margin-right": "10px"},
                    ),
                    html.A(
                        "FAQ",
                        href="/faq",
                        className="header-link",
                        style={"margin-right": "10px"},
                    ),
                    html.A(
                        "Docs",
                        href="/docs",
                        className="header-link",
                        style={"margin-right": "10px"},
                    ),
                    html.A(
                        "Weights Info",
                        href="/weights",
                        className="header-link",
                        style={"margin-right": "10px"},
                    ),
                ],
                style={"display": "inline-block"},
            ),
            html.Div(
                className="header-downloads",
                children=[
                    html.A(
                        "Download Scouting View",
                        href="/assets/fm-dash-search.fmf",
                        className="header-download-link",
                        style=button_style,
                        download="fm-dash-search.fmf",
                    ),
                    html.A(
                        "Download Squad View",
                        href="/assets/fm-dash-squad-view.fmf",
                        className="header-download-link",
                        style=button_style,
                        download="fm-dash-squad-view.fmf",
                    ),
                    html.A(
                        "Download Regen Faces",
                        href="https://api.fm-dash.com/regen_facepack/regen.zip",
                        className="header-download-link",
                        style=button_style,
                    ),
                ],
                style={"display": "inline-block", "margin-left": "auto"},
            ),
        ],
        style={
            "text-align": "center",
            "padding": "20px",
            "background-color": "#2B3A52",
            "color": "white",
            "display": "flex",
            "justify-content": "space-between",
        },
    )


def create_footer():
    return html.Footer(
        className="home-footer",
        children=[
            html.A(
                href="https://github.com/LiamHardman/fm-dash",
                children=[html.Img(src="/assets/github.svg", className="footer-icon")],
            ),
            html.A(
                href="https://discord.gg/9gbwDrNnzg",
                children=[html.Img(src="/assets/discord.png", className="footer-icon")],
            ),
        ],
        style={
            "text-align": "center",
            "padding": "20px",
            "background-color": "#2B3A52",
            "color": "white",
        },
    )


button_style = {
    "background-color": "#007BFF",
    "color": "white",
    "padding": "10px 15px",
    "text-align": "center",
    "text-decoration": "none",
    "display": "inline-block",
    "font-size": "16px",
    "margin": "4px 2px",
    "cursor": "pointer",
    "border": "none",
    "border-radius": "5px",
}
