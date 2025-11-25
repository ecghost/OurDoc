from fastapi import FastAPI
from routers import auth, login, mainpage, content, mydocs
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],   # 或 ["http://localhost:3000"]
    allow_credentials=True,
    allow_methods=["*"],   # 必须保证 OPTIONS 在里面
    allow_headers=["*"],
)

app.include_router(auth.router, prefix="/auth", tags=["Auth"])
app.include_router(login.router)
app.include_router(mainpage.router)
app.include_router(content.router)
app.include_router(mydocs.router)

for route in app.routes:
    print(route.path, route.name)