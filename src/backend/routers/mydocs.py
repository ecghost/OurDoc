from fastapi import APIRouter, Query
from typing import List
from pydantic import BaseModel
from routers.dataset_trick import get_user_list, get_doc_list, update_visibility_dataset, add_user_permission_dataset, remove_user_dataset, change_user_permission_dataset, change_room_name_dataset, delete_room_dataset

def id_to_color(id: str) -> str:
    h = abs(hash(id))  
    r = (h & 0xFF0000) >> 16
    g = (h & 0x00FF00) >> 8
    b = h & 0x0000FF
    return f'#{r:02x}{g:02x}{b:02x}'

router = APIRouter()

class User(BaseModel):
    id: str
    user_name: str
    email: str
    avatarColor: str

class UsersResponse(BaseModel):
    users: List[User]

class PermissionItem(BaseModel):
    id: str
    user_name: str
    email: str
    permission: int

class DocumentItem(BaseModel):
    room_id: str
    room_name: str
    create_time: str
    overall_permission: int
    permissions: dict[str, PermissionItem] = {}

class DocsResponse(BaseModel):
    docs: List[DocumentItem]

class UpdateVisibilityRequest(BaseModel):
    room_id: str
    overall_permission: int

class AddUserPermission(BaseModel):
    user_id: str
    permission: int

class AddUsersSubmitRequest(BaseModel):
    room_id: str
    users: List[AddUserPermission]

class RemoveUserPermission(BaseModel):
    room_id: str
    user_id: str

class ChangeUserPermission(BaseModel):
    room_id: str
    user_id: str
    permission: int

class RenameDoc(BaseModel):
    room_id: str
    room_name: str

class RemoveDoc(BaseModel):
    room_id: str


@router.get("/mydocs/getusers", response_model=UsersResponse)
async def get_users():
    datas = get_user_list()  # 返回列表[{"id":.., "user_name":.., "email":..}, ...]
    for data in datas:
        data['avatarColor'] = id_to_color(data['id'])
    print(datas)
    return {"users": datas}


@router.get("/mydocs/getdocs", response_model=DocsResponse)
async def get_docs(user_id: str = Query(..., description="当前登录用户的ID")):
    datas = get_doc_list(user_id)
    print(datas)
    return {"docs": datas}


@router.post("/mydocs/update_visibility")
async def update_visibility(data: UpdateVisibilityRequest):
    modify = update_visibility_dataset(data.room_id, data.overall_permission)
    if modify:
        return {'msg': '修改成功'}
    else:
        return False
    

@router.post("/mydocs/add_users")
async def add_users(data: AddUsersSubmitRequest):
    for index in range(len(data.users)):
        modify = add_user_permission_dataset(data.room_id, data.users[index].user_id, data.users[index].permission)
    # print(modify)
    if modify:
        return { "success": True }
    else:
        return False
    

@router.post("/mydocs/remove_user")
async def remove_user(data: RemoveUserPermission):
    modify = remove_user_dataset(data.room_id, data.user_id)
    if modify:
        return { "success": True }
    else:
        return False
    

@router.post("/mydocs/change_permission")
async def change_permission(data: ChangeUserPermission):
    modify = change_user_permission_dataset(data.room_id, data.user_id, data.permission)
    if modify:
        return { "success": True }
    else:
        return False
    

@router.post("/mydocs/rename_room")
async def rename_room(data: RenameDoc):
    modify = change_room_name_dataset(data.room_id, data.room_name)
    if modify:
        return { "success": True }
    else:
        return False
    

@router.post("/mydocs/delete_room")
async def delete_room(data: RemoveDoc):
    modify = delete_room_dataset(data.room_id)
    if modify:
        return { "success": True }
    else:
        return False
    