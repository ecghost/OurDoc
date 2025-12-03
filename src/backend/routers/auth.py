from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, EmailStr
from email.mime.text import MIMEText
from email.utils import formataddr
from email.header import Header
from passlib.hash import bcrypt
from fastapi import Request
import smtplib
import random
from routers.dataset import register_dataset, reset_password_dataset

router = APIRouter()

SMTP_HOST = "smtp.qq.com"
SMTP_PORT = 465
SMTP_USER = "1933920868@qq.com"
SMTP_PASS = "eltsxxlfrvfobfei"

verify_codes = {}

class RegisterModel(BaseModel):
    email: str
    username: str
    password: str
    verifyCode: str

class ResetPasswordModel(BaseModel):
    email: EmailStr
    newPassword: str
    verifyCode: str

class SendCodeModel(BaseModel):
    email: EmailStr


@router.post("/send-code")
async def send_code(data: SendCodeModel):
    code = str(random.randint(100000, 999999))
    verify_codes[data.email] = code

    # 发送邮件（这里先打印代替）发送邮箱功能已经实现，用success=True代替
    # success = send_email_code(data.email, code)
    success = True
    if not success:
        raise HTTPException(status_code=500, detail="邮件发送失败，请稍后再试")
    else:
        print(f"[验证码] 发送给 {data.email} 的验证码是：{code}")

    return {"msg": "验证码已发送"}


def send_email_code(to_email: str, code: str):
    subject = "您的验证码"
    content = f"您的验证码是：{code}，有效期 5 分钟，请不要泄露给他人。"

    message = MIMEText(content, "plain", "utf-8")
    message["From"] = formataddr(("系统邮件", SMTP_USER))
    message["To"] = Header(to_email, "utf-8")
    message["Subject"] = Header(subject, "utf-8")

    try:
        server = smtplib.SMTP_SSL(SMTP_HOST, SMTP_PORT)
        server.login(SMTP_USER, SMTP_PASS)
        server.sendmail(SMTP_USER, [to_email], message.as_string())
        server.quit()
        return True
    except Exception as e:
        print("发送邮件失败：", e)
        return False


@router.post("/register")
async def register(data: RegisterModel, request: Request):
    try:
        body = await request.json()
        print("收到的 body =", body)
        if data.verifyCode != verify_codes.get(data.email):
            raise HTTPException(status_code=400, detail="验证码错误")

        password = data.password[:72]
        hashed = bcrypt.hash(password)

        """
            这里接数据库逻辑，此处为注册，可以提供：
            注册用户的邮箱: data.email
            注册用户的密码: password
            注册用户的密码(hash): hashed
        """   
        msg = register_dataset(data.username, data.email, hashed)
        if msg:
            print(f"[注册成功] {data.email} 密码哈希 {hashed}")
            return {"msg": "密码重置成功"}

        else:
            raise HTTPException(status_code=400, detail=f"注册失败: {msg}")
    except Exception as e:
        import traceback
        traceback.print_exc()
        return {"error": str(e)}



@router.post("/reset-password")
async def reset_password(data: ResetPasswordModel):
    if data.verifyCode != verify_codes.get(data.email):
        raise HTTPException(status_code=400, detail="验证码错误")

    password = data.newPassword[:72]
    hashed = bcrypt.hash(password)

    """
        这里接数据库逻辑，此处为注册，可以提供：
        修改密码用户的邮箱: data.email
        修改密码用户的新密码: password
        修改密码用户的新密码(hash): hashed
    """   
    msg = reset_password_dataset(data.email, hashed)
    if msg:
        print(f"[密码重置] {data.email} 新密码哈希 {hashed}")
        return {"msg": "密码重置成功"}
    
    else:
        raise HTTPException(status_code=400, detail=f"密码重置失败: {msg}")
