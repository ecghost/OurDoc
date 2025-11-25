from fastapi import APIRouter, Query
from typing import List
from pydantic import BaseModel
from routers.dataset import main_page_dataset, get_edit_permission_dataset, get_read_permission_dataset

router = APIRouter()

class Room(BaseModel):
    room_id: str
    room_name: str
    owner_user_name: str
    permission: int


@router.get("/rooms", response_model=List[Room])
async def get_rooms(userid: str = Query(..., description="用户ID")):
    room_data = main_page_dataset(userid)
    return room_data


@router.get("/main/edit_permission", response_model=bool)
async def edit_permission(room_id: str = Query(...), user_id: str = Query(...)):
    # 获取权限值，假设 get_edit_permission_dataset 返回 0/1
    value = get_edit_permission_dataset(room_id, user_id)
    return bool(value)  # 强制转换成布尔值

@router.get("/main/read_permission", response_model=bool)
async def read_permission(room_id: str = Query(...), user_id: str = Query(...)):
    value = get_read_permission_dataset(room_id, user_id)
    print(value)
    return bool(value)