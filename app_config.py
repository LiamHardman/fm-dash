"""
Used to load initial configuration for the app, along with logging.
"""

import logging
import os
import yaml


class OtelLogFormatter(logging.Formatter):
    """
    Custom formatter for logging that adds OpenTelemetry trace and span IDs to the log output.
    """

    def format(self, record):
        record.otelTraceID = getattr(record, "otelTraceID", "N/A")
        record.otelSpanID = getattr(record, "otelSpanID", "N/A")
        record.otelTraceFlags = getattr(record, "otelTraceFlags", "N/A")
        return super().format(record)


def load_config():
    config_path = os.getenv("FMD_CONF_LOCATION", "config.yml")
    with open(config_path) as f:
        return yaml.safe_load(f)


def get_log_level(level_str):
    """
    Convert a string representation of a log level to its corresponding logging level object.
    """
    level_dict = {
        "CRITICAL": logging.CRITICAL,
        "ERROR": logging.ERROR,
        "WARNING": logging.WARNING,
        "INFO": logging.INFO,
        "DEBUG": logging.DEBUG,
        "NOTSET": logging.NOTSET,
    }
    return level_dict.get(level_str, logging.NOTSET)


def setup_logging(log_level_str):
    """
    Set up logging for the application.
    Relied upon for OTel implementation.
    """
    log_level = get_log_level(log_level_str)
    log_format = "%(asctime)s - %(levelname)s - %(message)s - [Trace: %(otelTraceID)s, Span: %(otelSpanID)s, Flags: %(otelTraceFlags)s]"
    formatter = OtelLogFormatter(log_format)

    logger = logging.getLogger()
    file_handler = None

    # Check if the file handler already exists to prevent adding it multiple times
    file_handler_exists = any(
        isinstance(handler, logging.FileHandler)
        and handler.baseFilename.endswith("fm-dash.log")
        for handler in logger.handlers
    )

    if not file_handler_exists:
        if not os.path.exists("logs/fm-dash"):
            os.makedirs("logs/fm-dash")

        file_handler = logging.FileHandler("logs/fm-dash/fm-dash.log")
        file_handler.setFormatter(formatter)
        logger.addHandler(file_handler)

    logger.setLevel(log_level)

    # Configure specific loggers explicitly
    for log_name in ["flask.app", "werkzeug", "dash"]:
        specific_logger = logging.getLogger(log_name)
        if file_handler:  # Check if file_handler is not None
            specific_logger.handlers = [file_handler]  # Explicitly set the handler
        specific_logger.setLevel(log_level)
        specific_logger.propagate = False
