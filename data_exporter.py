import dash
from dash.exceptions import PreventUpdate
import pandas as pd
from dash import dcc
from dash_core_components import send_data_frame
from opentelemetry import context
import io

from tracer_provider import tracer


def generate_download(n_clicks_csv, n_clicks_xlsx, n_clicks_md, n_clicks_html, data):
    """
    Generates a download file based on the button clicked and the provided data.

    Args:
        n_clicks_csv (int): Number of times the CSV export button was clicked.
        n_clicks_xlsx (int): Number of times the XLSX export button was clicked.
        n_clicks_md (int): Number of times the Markdown export button was clicked.
        n_clicks_html (int): Number of times the HTML export button was clicked.
        data (list): List of dictionaries representing the data to be exported.

    Returns:
        dash.core.Component: The appropriate `dcc.send_data_frame` or `dcc.send_bytes` component
        based on the button clicked.

    Raises:
        PreventUpdate: If the callback was not triggered or if the data is empty.
        Exception: If any other error occurs during the export process.
    """
    current_ctx = context.get_current()
    with tracer.start_as_current_span(
        "generate_download", context=current_ctx
    ) as current_span:
        try:
            columns_to_include = [
                "Name",
                "Position",
                "Age",
                "upper_value_range",
                "value_score",
                "overall_score",
                "physical_score",
                "mental_score",
                "technical_score",
            ]
            ctx = dash.callback_context
            if not ctx.triggered or not data:
                raise PreventUpdate

            button_id = ctx.triggered[0]["prop_id"].split(".")[0]
            df = pd.DataFrame.from_records(data)
            df = df[columns_to_include]

            if button_id == "export-csv-btn":
                buffer = io.StringIO()
                df.to_csv(buffer, index=False)
                file_size = buffer.tell()
                current_span.set_attribute("file_size_bytes", file_size)
                current_span.add_event("Exporting CSV")
                return dcc.send_data_frame(df.to_csv, filename="data.csv", index=False)

            elif button_id == "export-xlsx-btn":
                buffer = io.BytesIO()
                df.to_excel(buffer, index=False)
                file_size = buffer.tell()
                current_span.set_attribute("file_size_bytes", file_size)
                current_span.add_event("Exporting XLSX")
                return dcc.send_data_frame(
                    df.to_excel, filename="data.xlsx", index=False
                )

            elif button_id == "export-md-btn":
                md_string = df.to_markdown()
                file_size = len(md_string.encode())
                current_span.set_attribute("file_size_bytes", file_size)
                current_span.add_event("Exporting Markdown")
                return dcc.send_bytes(md_string.encode(), "data.md")

            elif button_id == "export-html-btn":
                current_span.add_event("Exporting HTML")
                html_string = df.to_html()

            styled_html = f"""
                <html>
                <head>
                <!-- styles and scripts -->
                </head>
                <body>
                    <table id="table-sorting-filtering">
                        {html_string}
                    </table>

                    <script>
                    // Initialize DataTables on the exported table
                    $(document).ready( function () {{
                        $('#table-sorting-filtering').DataTable();
                    }} );
                    </script>
                </body>
                </html>
            """
            file_size = len(styled_html.encode())
            current_span.set_attribute("file_size_bytes", file_size)
        except Exception as e:
            current_span.record_exception(e)
            raise
        return dcc.send_bytes(styled_html.encode(), "data.html")
