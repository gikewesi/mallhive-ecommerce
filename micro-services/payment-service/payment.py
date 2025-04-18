from fastapi import FastAPI, APIRouter, HTTPException
from pydantic import BaseModel
import os
import stripe
import boto3
import base64
import httpx

# ========== Configuration ==========
STRIPE_SECRET_KEY = os.getenv("STRIPE_SECRET_KEY", "your_stripe_secret_key_here")
ORDER_SERVICE_URL = os.getenv("ORDER_SERVICE_URL", "http://order-service/api/v1")
NOTIFICATION_SERVICE_URL = os.getenv("NOTIFICATION_SERVICE_URL", "http://notification-service/api/v1")

stripe.api_key = STRIPE_SECRET_KEY
kms_client = boto3.client("kms")

# ========== Models ==========
class PaymentRequest(BaseModel):
    order_id: str
    amount: float
    currency: str
    stripe_token: str
    user_email: str

# ========== Helper Functions ==========
def decrypt_stripe_token(encrypted_token: str) -> str:
    decrypted = kms_client.decrypt(
        CiphertextBlob=base64.b64decode(encrypted_token)
    )
    return decrypted["Plaintext"].decode("utf-8")

def validate_order(order_id: str, amount: float):
    response = httpx.get(f"{ORDER_SERVICE_URL}/orders/{order_id}")
    if response.status_code != 200:
        raise Exception("Order not found")
    if response.json()["amount"] != amount:
        raise Exception("Amount mismatch with order")

async def send_notification(email: str, payment_result):
    async with httpx.AsyncClient() as client:
        await client.post(f"{NOTIFICATION_SERVICE_URL}/notify", json={
            "email": email,
            "message": f"Your payment of ${payment_result.amount / 100:.2f} was successful."
        })

async def process_payment(req: PaymentRequest):
    validate_order(req.order_id, req.amount)
    decrypted_token = decrypt_stripe_token(req.stripe_token)

    charge = stripe.Charge.create(
        amount=int(req.amount * 100),
        currency=req.currency,
        source=decrypted_token,
        description=f"Payment for order {req.order_id}",
    )
    return charge

# ========== FastAPI Setup ==========
app = FastAPI(title="Payment Service", version="1.0")
router = APIRouter()

@router.post("/")
async def handle_payment(req: PaymentRequest):
    try:
        payment_result = await process_payment(req)
        await send_notification(req.user_email, payment_result)
        return {"status": "success", "payment_id": payment_result.id}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

app.include_router(router, prefix="/api/v1/payments")
