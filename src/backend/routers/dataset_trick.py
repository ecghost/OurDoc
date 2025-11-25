import os
import json
import zlib

trick_data_json = "json_trick/trick_data.json"
trick_data_home_json = "json_trick/home_data.json"
trick_data_home_permission_json = "json_trick/home_permision_data.json"
trick_data_home_content_json = "json_trick/home_content_data.json"



def generate_user_id(email: str) -> str:
    hash_int = zlib.crc32(email.encode("utf-8"))
    user_id = str(hash_int % 1000000).zfill(6)
    return user_id


def read_json(dataset_name):
    if dataset_name == 'user_table':
        with open(trick_data_json, "r") as f:
            data = json.load(f)
    elif dataset_name == 'user_room_table':
        with open(trick_data_home_json, "r") as f:
            data = json.load(f)
    elif dataset_name == 'room_permission_table':
        with open(trick_data_home_permission_json, "r") as f:
            data = json.load(f)
    elif dataset_name == 'room_content_table':
        with open(trick_data_home_content_json, "r") as f:
            data = json.load(f)
    return data


def write_json(dataset_name, data):
    if dataset_name == 'user_table':
        with open(trick_data_json, "w") as f:
            json.dump(data, f, indent=4)
    elif dataset_name == 'user_room_table':
        with open(trick_data_home_json, "w") as f:
            json.dump(data, f, indent=4)
    elif dataset_name == 'room_permission_table':
        with open(trick_data_home_permission_json, "w") as f:
            json.dump(data, f, indent=4)
    elif dataset_name == 'room_content_table':
        with open(trick_data_home_content_json, "w") as f:
            json.dump(data, f, indent=4)
    return data


# 主键查询，目前这个功能主要是作为查询整行数据
def read_dataset(dataset_name, main_key, goal_key):
    datas = read_json(dataset_name)
    if isinstance(main_key, tuple):
        results = [
            row for row in datas 
            if all(row[k] == v for k, v in zip(("room_id", "user_id"), main_key))
        ]
    else:
        results = [row for row in datas if row.get("id") == main_key or row.get("user_id") == main_key or row.get("room_id") == main_key]
    
    if not results:
        return None
    
    if goal_key != "*":
        return results[0].get(goal_key)
    return results


# 条件查询，这里主要是查询条件为...时的某个数据
def read_dataset_condition(dataset_name, key_name, key_value, goal_key):
    datas = read_json(dataset_name)
    results = [row for row in datas if row.get(key_name) == key_value]
    
    if not results:
        return None
    
    if goal_key != "*":
        return results[0].get(goal_key)
    return results


# 插入整行数据
def insert_data_into_dataset(dataset_name, data):
    datas = read_json(dataset_name)
    datas.append(data)
    write_json(dataset_name, datas)


# 查询某个键的所有值
def read_columns_values(dataset_name, keys):
    datas = read_json(dataset_name)
    if isinstance(keys, str):
        return [data[keys] for data in datas]
    elif isinstance(keys, tuple):
        return [{k: data[k] for k in keys} for data in datas]
    else:
        raise ValueError("keys 必须是 str 或 tuple")


# 条件查询多个值
def read_multidataset_condition(dataset_name, keys, condition_key, condition_value):
    datas = read_json(dataset_name)
    if isinstance(keys, str):
        return [data[keys] for data in datas if data.get(condition_key) == condition_value]
    elif isinstance(keys, tuple):
        return [
            {k: data[k] for k in keys}
            for data in datas
            if data.get(condition_key) == condition_value
        ]
    else:
        raise ValueError("keys 必须是 str 或 tuple")


# 修改某一个键的值，根据条件[某一个值对应]修改某一个键值的数据
def modify_dataset_condition(dataset_name, key_name, key_value, goal_key, goal_value):
    datas = read_json(dataset_name)
    modified = False

    for row in datas:
        if isinstance(key_name, tuple):
            # 多字段匹配
            if all(row.get(k) == v for k, v in zip(key_name, key_value)):
                row[goal_key] = goal_value
                modified = True
        else:
            # 单字段匹配
            if row.get(key_name) == key_value:
                row[goal_key] = goal_value
                modified = True
    
    if modified:
        write_json(dataset_name, datas)

    return modified


# 删除某一个主键对应的行
def remove_dataset_mainkey(dataset_name, main_key, main_value):
    datas = read_json(dataset_name)
    new_datas = []
    for data in datas:
        if isinstance(main_key, tuple):
            if data['room_id'] == main_value[0] and data['user_id'] == main_value[1]:
                continue
            else:
                new_datas.append(data)         
        else:
            if data[main_key] == main_value:
                continue
            else:
                new_datas.append(data)
    write_json(dataset_name, new_datas)
    return True


# 注册
def register_dataset(user_name, email, password):
    if not os.path.exists(trick_data_json):
        write_json("user_table", [])
    datas = read_dataset_condition("user_table", "email", email, password) 
    if datas != None:
        return False
    user_id = generate_user_id(email)
    if user_name == "":
        user_name = user_id
    datas = {"id": user_id, "user_name": user_id, "email": email, "password": password}
    insert_data_into_dataset("user_table", datas)
    return True


# 重置密码
def reset_password_dataset(email, password):
    if not os.path.exists(trick_data_json):
        write_json("user_table", [])
        return False
    modify = modify_dataset_condition("user_table", "email", email, "password", password)
    if modify:
        return True
    else:
        return False


# 登录
def login_dataset(email):
    if not os.path.exists(trick_data_json):
        write_json("user_table", [])
        return False
    password = read_dataset_condition("user_table", "email", email, "password")
    id = read_dataset_condition("user_table", "email", email, "id")
    return password, id


# 创建文档
def create_doc_dataset(room_id, room_name, create_time, user_id, content, overall_permission, permission):
    if not os.path.exists(trick_data_home_json):
        write_json("user_room_table", [])
    if not os.path.exists(trick_data_home_permission_json):
        write_json("room_permission_table", [])
    if not os.path.exists(trick_data_home_content_json):
        write_json("room_content_table", [])
    insert_data_into_dataset("user_room_table", {"room_id": room_id, "room_name": room_name, "create_time": create_time, "overall_permission": overall_permission, "owner_user_id": user_id})
    insert_data_into_dataset("room_permission_table", {"room_id": room_id, "user_id": user_id, "permission": permission})
    insert_data_into_dataset("room_content_table", {"room_id": room_id, "content": content})
    return {"room_id": room_id, "room_name": room_name, "create_time": create_time, "overall_permission":overall_permission, "msg": "创建成功", "success": True}



def main_page_dataset(userid):
    rooms = read_columns_values("user_room_table", ("room_id", "room_name", "owner_user_id", "overall_permission"))

    result = []

    for room in rooms:
        rid = room["room_id"]
        rname = room["room_name"]
        owner_id = room["owner_user_id"]
        overall_perm = room["overall_permission"]

        owner_name = read_dataset("user_table", owner_id, "user_name") or "未知"
        result.append({
            "room_id": rid,
            "room_name": rname,
            "owner_user_name": owner_name,
            "permission": overall_perm
        })
    return result


def get_content_dataset(room_id):
    data = read_dataset("room_content_table", room_id, "content")
    return data


def update_dataset(room_id, content):
    modify_dataset_condition("room_content_table", "room_id", room_id, "content", content)
    return True


def get_user_list():
    datas = read_columns_values("user_table", ('id', 'user_name', 'email'))
    return datas


def get_doc_list(user_id):
    datas = read_multidataset_condition("user_room_table", ("room_id", "room_name", "create_time", "overall_permission"), "owner_user_id", user_id)
    for data in datas:
        room_id = data['room_id']
        data_room_user_permission = read_multidataset_condition("room_permission_table", ("user_id", "permission"), "room_id", room_id)
        data_save = {}
        for user in data_room_user_permission:
            if user['permission'] not in [2, 3]:
                continue
            user_id_each = user['user_id']
            user_data = read_multidataset_condition("user_table", ("id", "user_name", "email"), "id", user_id_each)

            user_data[0]['permission'] = user['permission']
            data_save[user_id_each] = user_data[0]
        data['permissions'] = data_save
    # print(datas)
    return datas


def update_visibility_dataset(room_id, overall_permission):
    modify = modify_dataset_condition("user_room_table", "room_id", room_id, "overall_permission", overall_permission)
    return modify


def add_user_permission_dataset(room_id, user_id, permission):
    insert_data_into_dataset("room_permission_table", {"room_id": room_id, "user_id": user_id, "permission": permission})
    return True


def remove_user_dataset(room_id, user_id):
    remove = remove_dataset_mainkey("room_permission_table", ("room_id", "user_id"), (room_id, user_id))
    return remove


def change_user_permission_dataset(room_id, user_id, permission):
    modify = modify_dataset_condition("room_permission_table", ("room_id", "user_id") , (room_id, user_id), "permission", permission)
    return modify


def change_room_name_dataset(room_id, room_name):
    modify = modify_dataset_condition("user_room_table", "room_id", room_id, "room_name", room_name)
    return modify


def delete_room_dataset(room_id):
    remove_1 = remove_dataset_mainkey("user_room_table", "room_id", room_id)
    remove_2 = remove_dataset_mainkey("room_permission_table", "room_id", room_id)
    remove_3 = remove_dataset_mainkey("room_content_table", "room_id" , room_id)
    if remove_1 and remove_2 and remove_3:
        return True
    else:
        return False
    

def get_edit_permission_dataset(room_id, user_id):
    permission = read_dataset_condition("user_room_table", "room_id", room_id, "overall_permission")
    owner_user_id = read_dataset_condition("user_room_table", "room_id", room_id, "owner_user_id")
    if permission == 1 or user_id == owner_user_id:
        return True
    elif permission == 3:
        permission = read_dataset("room_permission_table", (room_id, user_id), "permission")
        if permission == None or permission == 2:
            return False
        else:
            return True
    else:
        return False
    

def get_read_permission_dataset(room_id, user_id):
    permission = read_dataset_condition("user_room_table", "room_id", room_id, "overall_permission")
    owner_user_id = read_dataset_condition("user_room_table", "room_id", room_id, "owner_user_id")
    if permission == 1 or permission == 2 or user_id == owner_user_id:
        return True
    elif permission == 3:
        permission = read_dataset("room_permission_table", (room_id, user_id), "permission")
        if permission == None:
            return False
        else:
            return True
    else:
        return False  