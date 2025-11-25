from fastapi import FastAPI
from routers import auth, login, mainpage, content, mydocs
from fastapi.middleware.cors import CORSMiddleware
from routers import dataset

app = FastAPI()
origins = [
    "http://localhost:5173",
    "http://127.0.0.1:5173",
]

app.add_middleware(
    CORSMiddleware,
    # allow_origins=["*"],   # 或 ["http://localhost:3000"]
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],   # 必须保证 OPTIONS 在里面
    allow_headers=["*"],
)

app.include_router(auth.router, prefix="/auth", tags=["Auth"])
app.include_router(login.router)
app.include_router(mainpage.router)
app.include_router(content.router)
app.include_router(mydocs.router)

# app.include_router(dataset.router, prefix="/api/dataset", tags=["Dataset"])

for route in app.routes:
    print(route.path, route.name)