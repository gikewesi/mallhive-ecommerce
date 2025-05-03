FROM python:3.9-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY ../../micro-services/payment-service/requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY ../../micro-services/payment-service/ .

RUN adduser --disabled-password --no-create-home paymentuser
USER paymentuser

EXPOSE 4100

CMD ["uvicorn", "payment:app", "--host", "0.0.0.0", "--port", "4100"]
