from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, EmailStr
from passlib.hash import bcrypt
from routers.dataset import login_dataset
# import httpx  # 如果要调用 Go API

router = APIRouter(prefix="/auth", tags=["Login"])

class LoginModel(BaseModel):
    email: EmailStr
    password: str

@router.post("/login")
async def login(data: LoginModel):
    hashed_password, id = login_dataset(data.email)
    if not hashed_password:
        raise HTTPException(status_code=400, detail="用户不存在")
    print(data.password)
    # 验证密码
    if not bcrypt.verify(data.password, hashed_password):
        raise HTTPException(status_code=400, detail="密码错误")

    return {
            "msg": "登录成功",
            "userId": id
            }
