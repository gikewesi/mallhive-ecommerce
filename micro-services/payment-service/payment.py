from fastapi import FastAPI, APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional
import os
import stripe
import boto3
import base64
import httpx
from dotenv import load_dotenv

# ========== Load Environment Variables ==========
load_dotenv()

# ========== Configuration ==========
STRIPE_SECRET_KEY = os.getenv("STRIPE_SECRET_KEY", "your_stripe_secret_key_here")
PAYPAL_CLIENT_ID = os.getenv("PAYPAL_CLIENT_ID", "")
PAYPAL_CLIENT_SECRET = os.getenv("PAYPAL_CLIENT_SECRET", "")
PAYPAL_ENV = os.getenv("PAYPAL_ENV", "sandbox")

ORDER_SERVICE_URL = os.getenv("ORDER_SERVICE_URL", "http://order-service/api/v1")
NOTIFICATION_SERVICE_URL = os.getenv("NOTIFICATION_SERVICE_URL", "http://notification-service/api/v1")

stripe.api_key = STRIPE_SECRET_KEY
kms_client = boto3.client("kms")

# ========== Models ==========
class PaymentRequest(BaseModel):
    order_id: str
    amount: float
    currency: str
    encrypted_token: str
    user_email: str
    provider: Optional[str] = "stripe"  # Can be "stripe" or "paypal"

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
    if float(response.json().get("amount", 0)) != amount:
        raise Exception("Amount mismatch with order")

async def send_notification(email: str, amount: float):
    async with httpx.AsyncClient() as client:
        await client.post(f"{NOTIFICATION_SERVICE_URL}/notify", json={
            "email": email,
            "message": f"Your payment of ${amount:.2f} was successful."
        })

async def process_stripe_payment(req: PaymentRequest):
    decrypted_token = decrypt_stripe_token(req.encrypted_token)
    charge = stripe.Charge.create(
        amount=int(req.amount * 100),  # convert dollars to cents
        currency=req.currency,
        source=decrypted_token,
        description=f"Payment for order {req.order_id}",
    )
    return {"id": charge.id, "status": charge.status}

async def process_paypal_payment(req: PaymentRequest):
    # Placeholder for future PayPal integration
    raise NotImplementedError("PayPal support is not yet implemented")

async def process_payment(req: PaymentRequest):
    validate_order(req.order_id, req.amount)
    if req.provider == "stripe":
        return await process_stripe_payment(req)
    elif req.provider == "paypal":
        return await process_paypal_payment(req)
    else:
        raise Exception("Unsupported payment provider")

# ========== FastAPI Setup ==========
app = FastAPI(title="Payment Service", version="1.0")
router = APIRouter()

@router.post("/")
async def handle_payment(req: PaymentRequest):
    try:
        payment_result = await process_payment(req)
        await send_notification(req.user_email, req.amount)
        return {
            "status": "success",
            "provider": req.provider,
            "payment_id": payment_result["id"]
        }
    except NotImplementedError as nie:
        raise HTTPException(status_code=501, detail=str(nie))
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

app.include_router(router, prefix="/api/v1/payments")
