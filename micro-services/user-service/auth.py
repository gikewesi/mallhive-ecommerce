import logging
from fastapi import HTTPException, status, BackgroundTasks, Depends
from fastapi.security import OAuth2PasswordRequestForm
from sqlalchemy.orm import Session
from passlib.context import CryptContext
from jose import jwt, JWTError
from datetime import datetime, timedelta
from . import database, secrets
from logging import log_event

log_event("auth_success", "User logged in", {"user_id": "abc123"})

from .metrics import record_metric
import requests

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")
SECRET_KEY = secrets.JWT_SECRET_KEY
ALGORITHM = secrets.JWT_ALGORITHM
ACCESS_TOKEN_EXPIRE_MINUTES = 60

NOTIFICATION_URL = "http://notification.internal.mallhive.com/send"

# Utility: create JWT
def create_access_token(data: dict, expires_delta: timedelta = None):
    to_encode = data.copy()
    expire = datetime.utcnow() + (expires_delta or timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES))
    to_encode.update({"exp": expire})
    return jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)

# Utility: verify JWT and return user
def get_current_user(token: str = Depends(secrets.oauth2_scheme), db: Session = Depends(database.get_db)):
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        email = payload.get("sub")
        if not email:
            raise HTTPException(status_code=401, detail="Invalid token payload")
        user = database.get_user_by_email(db, email)
        if not user:
            raise HTTPException(status_code=404, detail="User not found")
        return user
    except JWTError:
        raise HTTPException(status_code=401, detail="Invalid token")

# Utility: send notification
def send_notification(to_email: str, subject: str, message: str):
    try:
        response = requests.post(NOTIFICATION_URL, json={
            "to": to_email,
            "subject": subject,
            "message": message
        }, timeout=5)
        response.raise_for_status()
    except Exception as e:
        logger.error(f"Failed to send notification to {to_email}: {e}")
        raise HTTPException(status_code=502, detail="Notification service failed")

# Register
def register_user(user_data: dict, db: Session, background_tasks: BackgroundTasks):
    if database.get_user_by_email(db, user_data["email"]):
        raise HTTPException(status_code=400, detail="Email already registered")

    user_data["password"] = pwd_context.hash(user_data["password"])
    user = database.create_user(db, user_data)

    code = database.store_verification_code(db, user.email)
    background_tasks.add_task(send_notification, user.email, "Verify Your Email", f"Your code is: {code}")
    logger.info(f"Verification code sent to {user.email}")
    record_metric("user.registered")
    return {"message": "User registered. Check your email for verification code."}

# Login
def login_user(form_data: OAuth2PasswordRequestForm, db: Session):
    user = database.get_user_by_email(db, form_data.username)
    if not user or not pwd_context.verify(form_data.password, user.password):
        raise HTTPException(status_code=401, detail="Invalid credentials")
    if not user.is_verified:
        raise HTTPException(status_code=403, detail="Email not verified")

    token = create_access_token(data={"sub": user.email})
    logger.info(f"{user.email} logged in")
    record_metric("user.logged_in")
    return {"access_token": token, "token_type": "bearer"}

# Verify email
def verify_email(email: str, code: str, db: Session):
    if not database.verify_verification_code(db, email, code):
        raise HTTPException(status_code=400, detail="Invalid verification code")

    database.mark_user_as_verified(db, email)
    logger.info(f"{email} verified their email")
    record_metric("user.verified")
    return {"message": "Email verified successfully"}

# Resend verification
def resend_verification(email: str, db: Session, background_tasks: BackgroundTasks):
    user = database.get_user_by_email(db, email)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    if user.is_verified:
        raise HTTPException(status_code=400, detail="User already verified")

    code = database.store_verification_code(db, email)
    background_tasks.add_task(send_notification, email, "Resend Verification", f"Your new code is: {code}")
    logger.info(f"Resent verification code to {email}")
    record_metric("verification.resent")
    return {"message": "Verification code resent"}

# Forgot password
def forgot_password(email: str, db: Session, background_tasks: BackgroundTasks):
    user = database.get_user_by_email(db, email)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    code = database.store_reset_code(db, email)
    background_tasks.add_task(send_notification, email, "Reset Password", f"Reset code: {code}")
    logger.info(f"Reset code sent to {email}")
    record_metric("password.reset.requested")
    return {"message": "Reset code sent"}

# Reset password
def reset_password(email: str, code: str, new_password: str, db: Session):
    if not database.verify_reset_code(db, email, code):
        raise HTTPException(status_code=400, detail="Invalid reset code")

    hashed = pwd_context.hash(new_password)
    database.reset_user_password(db, email, hashed)
    logger.info(f"{email} changed password")
    record_metric("password.reset.success")
    return {"message": "Password updated successfully"}
