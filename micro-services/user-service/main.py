from fastapi import FastAPI, Depends, BackgroundTasks, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import OAuth2PasswordRequestForm
from sqlalchemy.orm import Session
from starlette.responses import JSONResponse

from auth import (
    register_user,
    login_user,
    verify_email,
    resend_verification,
    forgot_password,
    reset_password,
)
from database import get_db, UserCreate
from logging import get_logger
from metrics import record_metric  # Optional: if you use custom metrics
from slowapi import Limiter
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded

import uuid
from pydantic import ValidationError

logger = get_logger(__name__)

app = FastAPI(title="User Service")

# Set up rate limiting
limiter = Limiter(key_func=get_remote_address)
app.state.limiter = limiter

# CORS configuration
origins = ["https://homepage.mallhive.com"]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Middleware to attach correlation ID to each request
@app.middleware("http")
async def add_correlation_id(request: Request, call_next):
    correlation_id = request.headers.get("X-Correlation-ID") or str(uuid.uuid4())
    request.state.correlation_id = correlation_id
    response = await call_next(request)
    response.headers["X-Correlation-ID"] = correlation_id
    return response

@app.exception_handler(RateLimitExceeded)
async def rate_limit_handler(request: Request, exc: RateLimitExceeded):
    return JSONResponse(status_code=429, content={"detail": "Rate limit exceeded. Try again later."})


# --- ROUTES ---

@app.post("/register")
@limiter.limit("5/minute")
async def api_register_user(user_data: dict, background_tasks: BackgroundTasks, db: Session = Depends(get_db)):
    """
    Register a new user.
    """
    try:
        user_obj = UserCreate(**user_data)
    except ValidationError as e:
        raise HTTPException(status_code=422, detail=e.errors())

    return await register_user(user_obj, background_tasks, db)


@app.post("/login")
@limiter.limit("10/minute")
async def api_login_user(form_data: OAuth2PasswordRequestForm = Depends(), db: Session = Depends(get_db)):
    """
    Authenticate user and return access token.
    """
    token = await login_user(form_data.username, form_data.password, db)
    return {"access_token": token, "token_type": "bearer"}


@app.post("/verify-email")
@limiter.limit("5/minute")
async def api_verify_email(data: dict, db: Session = Depends(get_db)):
    """
    Verify user email with code.
    """
    email = data.get("email")
    code = data.get("code")
    if not email or not code:
        raise HTTPException(status_code=400, detail="Email and code required")

    return await verify_email(email, code, db)


@app.post("/resend-verification")
@limiter.limit("5/minute")
async def api_resend_verification(data: dict, background_tasks: BackgroundTasks, db: Session = Depends(get_db)):
    """
    Resend email verification code.
    """
    email = data.get("email")
    if not email:
        raise HTTPException(status_code=400, detail="Email required")

    return await resend_verification(email, background_tasks, db)


@app.post("/forgot-password")
@limiter.limit("5/minute")
async def api_forgot_password(data: dict, background_tasks: BackgroundTasks, db: Session = Depends(get_db)):
    """
    Send reset code to email.
    """
    email = data.get("email")
    if not email:
        raise HTTPException(status_code=400, detail="Email required")

    return await forgot_password(email, background_tasks, db)


@app.post("/reset-password")
@limiter.limit("5/minute")
async def api_reset_password(data: dict, db: Session = Depends(get_db)):
    """
    Reset password using code sent to email.
    """
    email = data.get("email")
    code = data.get("code")
    new_password = data.get("new_password")

    if not all([email, code, new_password]):
        raise HTTPException(status_code=400, detail="Email, code, and new_password required")

    return await reset_password(email, code, new_password, db)
