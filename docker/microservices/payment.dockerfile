FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app
RUN apt-get update && apt-get install -y libbrlapi-dev
RUN apt-get update && apt-get install -y cloud-init


COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY . .

RUN adduser --disabled-password --no-create-home paymentuser
USER paymentuser

EXPOSE 4100

CMD ["uvicorn", "payment:app", "--host", "0.0.0.0", "--port", "4100"]
