FROM python:3.9-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY ../../micro-services/user-service/requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY ../../micro-services/user-service/ .

RUN adduser --disabled-password --gecos '' appuser
USER appuser

EXPOSE 4600

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "4600"]
