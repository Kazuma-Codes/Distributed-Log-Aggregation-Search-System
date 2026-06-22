# =============================================================================
# Log Generator — Container Image
# Distributed Log Aggregation & Search System
# =============================================================================
FROM python:3.12-slim

LABEL maintainer="log-platform-team"
LABEL description="Realistic log generator for the Distributed Log Aggregation & Search System"

WORKDIR /app

# --- Dependencies (cached layer) -------------------------------------------
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# --- Application code & templates ------------------------------------------
COPY log-generator.py .
COPY templates/ ./templates/

# --- Non-root user ----------------------------------------------------------
RUN useradd --create-home appuser
USER appuser

ENTRYPOINT ["python", "log-generator.py"]
CMD ["--rate", "100", "--format", "microservice", "--output", "stdout"]
