FROM python:3.11-slim

WORKDIR /app

# Install dependencies with extended timeout and retry tolerance
COPY requirements.txt .
RUN pip install --default-timeout=1000 --retries=10 --no-cache-dir -r requirements.txt

# Copy application files (dataset, model, code)
COPY . .

EXPOSE 8001

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8001"]
