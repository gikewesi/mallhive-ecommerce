import logging
import os
from datadog import initialize, api

DATADOG_ENABLED = bool(os.getenv("DATADOG_API_KEY") and os.getenv("DATADOG_APP_KEY"))

if DATADOG_ENABLED:
    initialize(
        api_key=os.getenv("DATADOG_API_KEY"),
        app_key=os.getenv("DATADOG_APP_KEY")
    )

def get_logger(name: str = "user_service") -> logging.Logger:
    """Returns a pre-configured logger with consistent formatting."""
    logger = logging.getLogger(name)
    if logger.hasHandlers():
        return logger

    logger.setLevel(logging.INFO)

    handler = logging.StreamHandler()
    formatter = logging.Formatter("[%(asctime)s] [%(levelname)s] [%(name)s] %(message)s")
    handler.setFormatter(formatter)

    logger.addHandler(handler)
    logger.propagate = False

    return logger

def log_event(event_type: str, message: str, metadata: dict = None):
    """Log a structured event locally and optionally forward to Datadog."""
    logger = get_logger("user_service")
    full_message = f"[{event_type}] {message}"
    if metadata:
        full_message += f" | Metadata: {metadata}"

    logger.info(full_message)

    if DATADOG_ENABLED:
        try:
            api.Event.create(
                title=f"user-service: {event_type}",
                text=full_message,
                tags=["service:user-service"]
            )
        except Exception as e:
            logger.warning(f"Failed to send event to Datadog: {e}")
