"""
Single function module to return the docs_layout page.
"""

import dash_html_components as html

import load_markdown
import header_footer


def docs_page_layout(markdown_files):
    """
    Creates the layout for the documentation page.

    Parameters
    ----------
    markdown_files : list
        A list of markdown files to display.

    Returns
    -------
    dash_html_components.Div
        The layout for the documentation page.
    """

    docs_links = load_markdown.generate_docs_links(markdown_files)

    return html.Div(
        [
            header_footer.create_header(),
            html.Div(
                className="docs-container",
                children=[
                    html.H1("FM-Dash Docs", className="docs-title"),
                    html.Div(
                        className="table-of-contents",
                        children=[
                            html.H2("Table of Contents", className="toc-title"),
                            *docs_links,
                        ],
                        style={"padding": "10px"},
                    ),
                ],
                style={
                    "padding": "20px",
                    "background-color": "#1E293B",
                    "color": "white",
                },
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
