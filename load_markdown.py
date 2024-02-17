import os
import dash_html_components as html
import dash_core_components as dcc
import logging
import header_footer


def read_markdown_files(folder_path):
    markdown_files = {}
    try:
        for filename in os.listdir(folder_path):
            if filename.endswith(".md"):
                file_path = os.path.join(folder_path, filename)
                with open(file_path, "r") as file:
                    markdown_files[filename] = file.read()
    except FileNotFoundError:
        logging.error("Folder not found: %s", folder_path)
    except PermissionError:
        logging.error("Permission denied when accessing folder: %s", folder_path)
    except Exception as e:
        logging.error("Error reading markdown files: %s", str(e))
    return markdown_files


def generate_docs_links(markdown_files):
    links = []
    try:
        for filename in markdown_files.keys():
            title = filename.replace(".md", "").replace("-", " ").title()
            url_path = "/" + filename.replace(".md", "")
            links.append(dcc.Link(title, href=url_path))
            links.append(html.Br())
    except Exception as e:
        logging.error("Error generating documentation links: %s", str(e))
    return links


def generate_markdown_layout(content):
    return html.Div(
        [
            header_footer.create_header(),
            html.Div(
                dcc.Markdown(content),
                className="markdown-container",
                style={
                    "text-align": "left",
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
