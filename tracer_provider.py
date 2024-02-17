import os
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource

import app_config

yaml_config = app_config.load_config()
enable_tracing = yaml_config["enable_tracing"]
trace_endpoint = yaml_config["trace_endpoint"]
new_relic_api_key = os.getenv("NEW_RELIC_API_KEY")
if enable_tracing:
    resource = Resource(
        attributes={
            "service.name": "FM-Dash",
            "service.version": "0.3.0",
            "environment": "production",
        }
    )

    tracer_provider = TracerProvider(resource=resource)
    trace.set_tracer_provider(tracer_provider)

    # Change the endpoint to New Relic OTLP endpoint and add headers for authentication
    otlp_exporter = OTLPSpanExporter(
        endpoint="https://otlp.eu01.nr-data.net:4317",  # New Relic endpoint
        insecure=False,  # Set to False to use https
        headers=(
            ("api-key", new_relic_api_key),
        ),  # Include your New Relic API Key here
    )

    span_processor = BatchSpanProcessor(otlp_exporter)
    tracer_provider.add_span_processor(span_processor)

    # Get a tracer if tracing is enabled
    tracer = trace.get_tracer(__name__)

else:
    tracer_provider = TracerProvider()
    trace.set_tracer_provider(tracer_provider)
    tracer = trace.get_tracer(__name__)
