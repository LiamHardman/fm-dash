# Import necessary modules and components
from dash import html
import home_screen
import faq_layout
import docs_layout
import weight_layout
from app_layout import get_layout
import app_config
import load_markdown

yaml_config = app_config.load_config()
attributes = yaml_config["attributes"]
attributes_physical = list(attributes["physical"].values())
attributes_mental = list(attributes["mental"].values())
attributes_technical = list(attributes["technical"].values())


def get_page_layout(pathname, yaml_config, weights_data, markdown_files):
    """
    Determines and returns the layout for a given pathname.

    Parameters:
    - pathname (str): The URL pathname indicating the current page being accessed.
    - yaml_config (dict): The application configuration loaded from a YAML file.
    - weights_data (dict): Data related to weights used in the application.
    - markdown_files (dict): Markdown content for documentation or info pages.

    Returns:
    - A Dash layout (html.Div or similar) corresponding to the given pathname.
    """
    if pathname == "/":
        return home_screen.home_page_layout()
    elif pathname == "/app":
        return get_layout(yaml_config["axes_map"])
    elif pathname == "/faq":
        return faq_layout.faq_page_layout()
    elif pathname == "/weights":
        return weight_layout.weights_page_layout(weights_data)
    elif pathname == "/docs":
        return docs_layout.docs_page_layout(markdown_files)
    elif pathname.startswith("/weights/"):
        simplified_role = pathname.split("/")[-1]
        role = next(
            (
                r
                for r in weights_data.keys()
                if weight_layout.simplify_role_name(r) == simplified_role
            ),
            None,
        )
        if role:
            role_data = weights_data.get(role, {})
            return weight_layout.generate_layout_for_role(
                attributes_physical, attributes_mental, attributes_technical, role_data
            )
        else:
            return "Role not found"
    elif pathname[1:] + ".md" in markdown_files:
        content = markdown_files[pathname[1:] + ".md"]
        return load_markdown.generate_markdown_layout(content)
    else:
        return "Page not found"
