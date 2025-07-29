import logging
from datetime import datetime, timedelta

from sqlalchemy import (
    create_engine, Column, Integer, String,
    DateTime, and_
)
from sqlalchemy.orm import sessionmaker, declarative_base, Session
from passlib.context import CryptContext
from pydantic import BaseModel, EmailStr

from .secrets import get_secret
from .logging import get_logger

logger = get_logger(__name__)

# Load DB credentials
try:
    secrets = get_secret()
except Exception as e:
    logger.error("Failed to load DB secrets: %s", e)
    raise RuntimeError("DB credentials could not be loaded.") from e

# Setup DB
DB_USER = secrets["username"]
DB_PASS = secrets["password"]
DB_HOST = secrets["host"]
DB_PORT = secrets["port"]
DB_NAME = secrets["dbname"]
DATABASE_URL = f"postgresql://{DB_USER}:{DB_PASS}@{DB_HOST}:{DB_PORT}/{DB_NAME}"

engine = create_engine(DATABASE_URL)
SessionLocal = sessionmaker(bind=engine, autoflush=False, autocommit=False)
Base = declarative_base()
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


# Pydantic schema
class UserCreate(BaseModel):
    username: str
    first_name: str | None = None
    last_name: str | None = None
    email: EmailStr
    phone_number: str | None = None
    gender: str | None = None
    address: str | None = None
    password: str

    class Config:
        orm_mode = True


# SQLAlchemy model
class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True)
    username = Column(String, nullable=False, unique=True, index=True)
    first_name = Column(String)
    last_name = Column(String)
    email = Column(String, nullable=False, unique=True, index=True)
    phone_number = Column(String)
    gender = Column(String)
    address = Column(String)
    hashed_password = Column(String, nullable=False)

    verification_code = Column(String, nullable=True)
    verification_code_expiry = Column(DateTime, nullable=True)
    reset_code = Column(String, nullable=True)
    reset_code_expiry = Column(DateTime, nullable=True)
    is_verified = Column(Integer, default=0)  # 0 = False, 1 = True


# DB utilities
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


# Auth helpers
def verify_password(plain_password: str, hashed_password: str) -> bool:
    return pwd_context.verify(plain_password, hashed_password)

def hash_password(password: str) -> str:
    return pwd_context.hash(password)


# CRUD
def create_user(db: Session, user_data: UserCreate):
    try:
        hashed_pw = hash_password(user_data.password)
        user = User(
            username=user_data.username,
            first_name=user_data.first_name,
            last_name=user_data.last_name,
            email=user_data.email,
            phone_number=user_data.phone_number,
            gender=user_data.gender,
            address=user_data.address,
            hashed_password=hashed_pw,
        )
        db.add(user)
        db.commit()
        db.refresh(user)
        logger.info("Created user: %s", user.email)
        return user
    except Exception as e:
        db.rollback()
        logger.error("Failed to create user: %s", e)
        raise


def get_user_by_email(db: Session, email: str):
    return db.query(User).filter(User.email == email).first()

def get_user_by_username(db: Session, username: str):
    return db.query(User).filter(User.username == username).first()

def get_user_by_id(db: Session, user_id: int):
    return db.query(User).filter(User.id == user_id).first()


# Code management
def store_verification_code(db: Session, email: str, code: str, expiry_minutes=15):
    user = get_user_by_email(db, email)
    if not user:
        raise ValueError("User not found")
    user.verification_code = code
    user.verification_code_expiry = datetime.utcnow() + timedelta(minutes=expiry_minutes)
    db.commit()

def verify_verification_code(db: Session, email: str, code: str):
    user = get_user_by_email(db, email)
    if not user:
        return False
    if (
        user.verification_code == code and
        user.verification_code_expiry and
        datetime.utcnow() <= user.verification_code_expiry
    ):
        return True
    return False

def mark_user_as_verified(db: Session, email: str):
    user = get_user_by_email(db, email)
    if user:
        user.is_verified = 1
        db.commit()

def store_reset_code(db: Session, email: str, code: str, expiry_minutes=15):
    user = get_user_by_email(db, email)
    if not user:
        raise ValueError("User not found")
    user.reset_code = code
    user.reset_code_expiry = datetime.utcnow() + timedelta(minutes=expiry_minutes)
    db.commit()

def verify_reset_code(db: Session, email: str, code: str):
    user = get_user_by_email(db, email)
    if not user:
        return False
    if (
        user.reset_code == code and
        user.reset_code_expiry and
        datetime.utcnow() <= user.reset_code_expiry
    ):
        return True
    return False

def reset_user_password(db: Session, email: str, new_password: str):
    user = get_user_by_email(db, email)
    if not user:
        raise ValueError("User not found")
    user.hashed_password = hash_password(new_password)
    user.reset_code = None
    user.reset_code_expiry = None
    db.commit()
    logger.info("Password reset for user: %s", email)
