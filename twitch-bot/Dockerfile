FROM python:3.11-slim

WORKDIR /workspace

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/ ./app

ENV PYTHONPATH=/workspace
CMD ["python", "-m", "app.main"]
