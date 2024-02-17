import dash_html_components as html
import header_footer
import yaml
import os


def load_config():
    # Try to get the path from the environment variable
    config_path = os.getenv("FMD_CONF_LOCATION", "config.yml")

    # Read the configuration file from the path
    with open(config_path) as f:
        return yaml.safe_load(f)


yaml_config = load_config()

attributes = yaml_config["attributes"]
attributes = yaml_config["attributes"]
attributes_physical = list(attributes["physical"].values())
attributes_mental = list(attributes["mental"].values())
attributes_technical = list(attributes["technical"].values())


def get_long_name(short_name):
    for category_name in ["physical", "mental", "technical"]:
        category_dict = attributes[category_name]
        for long_name, short in category_dict.items():
            if short == short_name:
                return long_name
    return short_name


def simplify_role_name(role_name):
    return "".join(e for e in role_name if e.isalnum()).lower()


def weights_page_layout(weights_data):
    defense_roles = [
        role
        for role in weights_data
        if role.startswith("D") and not role.startswith("DM") or role.startswith("WB")
    ]
    midfield_roles = [
        role for role in weights_data if role.startswith("M") or role.startswith("DM")
    ]
    attack_roles = [role for role in weights_data if role.startswith("S")]

    def create_role_links(roles):
        return [
            html.A(
                role,
                href=f"/weights/{simplify_role_name(role)}",
                style={"display": "block", "fontSize": "150%"},
            )
            for role in roles
        ]

    defense_links = create_role_links(defense_roles)
    midfield_links = create_role_links(midfield_roles)
    attack_links = create_role_links(attack_roles)

    header_style = {
        "textAlign": "center",
        "color": "white",
        "fontSize": "150%",
        "padding": "10px",
        "border": "1px solid #ddd",
    }
    cell_style = {"textAlign": "left", "padding": "10px", "border": "1px solid #ddd"}

    table_layout = html.Table(
        [
            html.Thead(
                html.Tr(
                    [
                        html.Th("Defense", style=header_style),
                        html.Th("Midfield", style=header_style),
                        html.Th("Attack", style=header_style),
                    ]
                )
            ),
            html.Tbody(
                [
                    html.Tr(
                        [
                            html.Td(html.Div(defense_links), style=cell_style),
                            html.Td(html.Div(midfield_links), style=cell_style),
                            html.Td(html.Div(attack_links), style=cell_style),
                        ]
                    )
                ]
            ),
        ],
        style={"width": "100%", "margin": "auto", "borderCollapse": "collapse"},
    )

    layout = html.Div(
        [
            header_footer.create_header(),
            html.Div(
                [  # Content container
                    html.H1(
                        "Select a Role",
                        style={
                            "textAlign": "center",
                            "color": "white",
                            "fontSize": "200%",
                        },
                    ),
                    table_layout,
                ],
                style={"flex": "1"},
            ),
            header_footer.create_footer(),
        ],
        style={"display": "flex", "flexDirection": "column", "minHeight": "100vh"},
    )

    return layout


def generate_layout_for_role(
    attributes_physical, attributes_mental, attributes_technical, role_data
):
    if not isinstance(role_data, dict):
        return html.Div(
            "Error: Role data is not in the expected format.",
            style={"color": "red", "fontSize": "16px"},
        )

    def get_color(value):
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

    physical_attributes = {
        k: v / 5 for k, v in role_data.items() if k in attributes_physical
    }
    mental_attributes = {
        k: v / 5 for k, v in role_data.items() if k in attributes_mental
    }
    technical_attributes = {
        k: v / 5 for k, v in role_data.items() if k in attributes_technical
    }

    header_style = {"backgroundColor": "#2B3A52", "color": "white", "fontSize": "32px"}
    cell_style = {"fontSize": "24px"}

    rows = []
    max_length = max(
        len(physical_attributes), len(mental_attributes), len(technical_attributes)
    )

    for i in range(max_length):
        row = []
        for category, attributes in zip(
            [technical_attributes, mental_attributes, physical_attributes],
            [attributes_technical, attributes_mental, attributes_physical],
        ):
            attr_short_name = attributes[i] if i < len(attributes) else ""
            attr_long_name = get_long_name(attr_short_name)
            attr_value = category.get(attr_short_name, "")
            color = get_color(attr_value) if attr_short_name in category else ""
            row.append(html.Td(attr_long_name, style=cell_style))
            row.append(
                html.Td(str(int(attr_value)), style={**cell_style, "color": color})
                if attr_short_name
                else html.Td("")
            )
        rows.append(html.Tr(row))

    table = html.Table(
        [
            html.Tr(
                [
                    html.Th("Technical", style=header_style),
                    html.Th("", style=header_style),
                    html.Th("Mental", style=header_style),
                    html.Th("", style=header_style),
                    html.Th("Physical", style=header_style),
                    html.Th("", style=header_style),
                ]
            )
        ]
        + rows,
        style={"margin": "auto", "width": "60%", "textAlign": "left"},
    )

    return html.Div(
        [
            header_footer.create_header(),
            html.Div(table, style={"overflowX": "auto"}),
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
