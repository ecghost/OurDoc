from fastapi import APIRouter, Query
from typing import List
from pydantic import BaseModel
from datetime import datetime
from routers.dataset import create_doc_dataset, get_content_dataset, update_dataset
import zlib

router = APIRouter()

class Room(BaseModel):
    room_name: str
    user_id: str

class SaveContentData(BaseModel):
    room_id: str
    content: str


def generate_room_id(name: str) -> str:
    hash_int = zlib.crc32(name.encode("utf-8"))
    room_id = str(hash_int % 1000000).zfill(6)
    return room_id


@router.post("/content/createdoc")
async def createdoc(data: Room):
    room_name = data.room_name
    room_id = generate_room_id(room_name)
    create_time = datetime.now().strftime("%Y-%m-%d")

    user_id = data.user_id
    content = ""
    overall_permission = 1
    permission = 1

    res = create_doc_dataset(room_id, room_name, create_time, user_id, content, overall_permission, permission)
    return res



@router.get("/content/getcontent")
async def get_content(room_id: str = Query(...)):
    contents = get_content_dataset(room_id)
    return {"room_id": room_id, "room_name": "", "content": contents} 



@router.post("/content/update")
async def update(data: SaveContentData):
    room_id = data.room_id
    content = data.content
    update_dataset(room_id, content)
    return {"msg": "保存成功", "room_id": room_id}