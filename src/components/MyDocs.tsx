// MyDocsPage.tsx
import React, {useEffect, useMemo, useState} from 'react'
import {
    Layout,
    Menu,
    Card,
    Radio,
    Button,
    Modal,
    Input,
    List,
    Avatar,
    Space,
    message,
    Popconfirm,
    Table,
    Tag,
    Typography,
} from 'antd'
import type {MenuProps} from 'antd'
import {
    PlusOutlined,
    DeleteOutlined,
    UserAddOutlined,
    ExclamationCircleOutlined,
    EditOutlined,
    LeftOutlined
} from '@ant-design/icons'
import {useNavigate} from "react-router-dom";

import axios from 'axios';
import { useLocation } from 'react-router-dom';

const {Sider, Content} = Layout
const {Search} = Input
const {Text} = Typography

// ----------------- types -----------------
type DocVisibility = 0 | 1 | 2 | 3

type DocumentItem = {
    room_id: string
    room_name: string
    create_time: string
    overall_permission: DocVisibility
    // permissions only used when visibility === 'partial'
    permissions: Record<
        string,
        { id: string; user_name: string; email: string; permission: 2 | 3 }
    >
}

type User = {
    id: string
    user_name: string
    email: string
    avatarColor?: string
}


const MyDocsPage: React.FC = () => {
    
    const location = useLocation();
    const userId = location.state?.userId || localStorage.getItem('UserId') || '';
    // navigate
    const navigate = useNavigate()
    // docs state (would normally come from API)
    const [docs, setDocs] = useState<DocumentItem[]>([])
    const [selectedDocId, setSelectedDocId] = useState<string | null>(null)

    // modal & user search state
    const [addUserModalVisible, setAddUserModalVisible] = useState(false)
    const [userSearchText, setUserSearchText] = useState('')
    const [selectedUsersInModal, setSelectedUsersInModal] = useState<
        Array<{ user: User; permission: 2 | 3 }>
    >([])

    // rename modal state
    const [renameModalVisible, setRenameModalVisible] = useState(false)
    const [renameValue, setRenameValue] = useState('')

    // simulate list of all users (would normally come from API)
    const [allUsers, setAllUsers] = useState<User[]>([])

    useEffect(() => {
        const fetchAll = async () => {
            try {
                const [docsRes, usersRes] = await Promise.all([
                    axios.get("http://localhost:8000/mydocs/getdocs", {
                        params: { user_id: userId }
                    }),
                    axios.get("http://localhost:8000/mydocs/getusers"),
                ]);

                setDocs(docsRes.data.docs);   // 后端返回 {docs: [...]}
                setAllUsers(usersRes.data.users); // {users: [...]}

                if (docsRes.data.docs.length > 0) {
                    setSelectedDocId(docsRes.data.docs[0].id);
                }
            } catch (err) {
                console.error("加载文档失败", err);
                message.error("加载文档失败");
            }
        };

        fetchAll();
    }, [userId]);

    const selectedDoc = useMemo(
        () => docs.find((d) => d.room_id === selectedDocId) ?? null,
        [docs, selectedDocId],
    )

    // ----- handlers for visibility -----
    const handleVisibilityChange = async  (val: DocVisibility) => {
        if (!selectedDoc) return
        const newDocs = docs.map((d) =>
            d.room_id === selectedDoc.room_id ? {...d, overall_permission: val} : d,
        )
        setDocs(newDocs)
        // api:向后端同步

        try {
            await axios.post("http://localhost:8000/mydocs/update_visibility", {
                room_id: selectedDoc.room_id,
                overall_permission: val,
            });
            // message.success("保存可见性设置成功");
        } catch (err) {
            console.error("同步可见性失败", err);
            message.error("同步可见性失败");
        }

        message.success('保存可见性设置')
    }

    // ----- permissions helper functions -----
    const listPermissions = selectedDoc
        //因为permissions是一个Record[]，所以values提取出Record的所有的键值对的值
        ? Object.values(selectedDoc.permissions).sort((a, b) =>
            a.user_name.localeCompare(b.user_name),
        )
        : []

    const handleRemovePermission = async (userId: string) => {
        if (!selectedDoc) return

        try {
            const res = await axios.post('http://localhost:8000/mydocs/remove_user', {
                room_id: selectedDoc.room_id,
                user_id: userId,
            })
            if (res.data.success) {
                const newPermissions = { ...selectedDoc.permissions }
                delete newPermissions[userId]
                const newDocs = docs.map((d) =>
                    d.room_id === selectedDoc.room_id ? { ...d, permissions: newPermissions } : d
                )
                setDocs(newDocs)
                message.success('已移除该用户权限')
            } else {
                message.error(res.data.message || '删除失败')
            }
        } catch (err) {
            console.error('删除权限失败', err)
            message.error('网络错误，删除失败')
        }
    }

    const handleChangePermission = async (userId: string, permission: 2|3) => {
        if (!selectedDoc) return
        try {
            const res = await axios.post('http://localhost:8000/mydocs/change_permission', {
                room_id: selectedDoc.room_id,
                user_id: userId,
                permission: permission,
            });

            if (res.data.success) {
                const newPermissions = {
                    ...selectedDoc.permissions,
                    [userId]: {
                        ...(selectedDoc.permissions[userId] || { id: userId, user_name: '', email: '' }),
                        permission,
                    },
                };

                const u = allUsers.find((x) => x.id === userId);
                if (u) {
                    newPermissions[userId].user_name = u.user_name;
                    newPermissions[userId].email = u.email;
                }

                const newDocs = docs.map((d) =>
                    d.room_id === selectedDoc.room_id ? { ...d, permissions: newPermissions } : d
                );
                setDocs(newDocs);
                message.success('权限修改成功');
            } else {
                message.error(res.data.message || '修改失败');
            }
        } catch (err) {
            console.error('修改权限失败', err);
            message.error('网络错误，修改失败');
        }
    }

    // ----- add users modal logic -----
    const openAddUsers = () => {
        setSelectedUsersInModal([])
        setUserSearchText('')
        setAddUserModalVisible(true)
    }

    const closeAddUsers = () => {
        setAddUserModalVisible(false)
    }

    const filteredUsersForModal = useMemo(() => {
        const q = userSearchText.trim().toLowerCase()
        if (!q) return allUsers
        return allUsers.filter(
            (u) => u.user_name.toLowerCase().includes(q) || u.email.toLowerCase().includes(q),
        )
    }, [allUsers, userSearchText])

    const toggleSelectUserInModal = (user: User) => {
        const exists = selectedUsersInModal.find((s) => s.user.id === user.id)
        if (exists) {
            setSelectedUsersInModal((prev) => prev.filter((s) => s.user.id !== user.id))
        } else {
            setSelectedUsersInModal((prev) => [...prev, {user, permission: 3}])
        }
    }

    const setPermissionForSelectedUser = (userId: string, permission: 2 | 3) => {
        setSelectedUsersInModal((prev) =>
            prev.map((p) => (p.user.id === userId ? {...p, permission} : p)),
        )
    }

    const handleAddUsersSubmit = async () => {
        if (!selectedDoc) {
            message.error('请选择要添加权限的文档')
            return
        }
        if (selectedUsersInModal.length === 0) {
            message.warning('请先选择用户')
            return
        }
        const payload = {
            room_id: selectedDoc.room_id,
            users: selectedUsersInModal.map((s) => ({
                user_id: s.user.id,
                permission: s.permission
            }))
        }
        try {
        // 调用后端接口
            const res = await axios.post('http://localhost:8000/mydocs/add_users', payload)

            if (res.data.success) {
                // 成功后更新前端状态
                const newPermissions = {...(selectedDoc.permissions || {})}
                selectedUsersInModal.forEach((s) => {
                    newPermissions[s.user.id] = {
                        id: s.user.id,
                        user_name: s.user.user_name,
                        email: s.user.email,
                        permission: s.permission,
                    }
                })
                const newDocs = docs.map((d) =>
                    d.room_id === selectedDoc.room_id ? {...d, permissions: newPermissions} : d,
                )

                setDocs(newDocs)
                setSelectedDocId(selectedDoc.room_id)
                setAddUserModalVisible(false)
                message.success('添加用户并设置权限成功')
            } else {
                message.error(res.data.message || '添加失败')
            }
        } catch (error) {
            console.error(error)
            message.error('网络错误，添加失败')
        }
    }

    // ----- rename logic -----
    const openRenameModal = () => {
        if (!selectedDoc) {
            message.warning('请选择要重命名的文档')
            return
        }
        setRenameValue(selectedDoc.room_name)
        setRenameModalVisible(true)
    }

    const closeRenameModal = () => {
        setRenameModalVisible(false)
        setRenameValue('')
    }

    const handleRenameSubmit = async () => {
        if (!selectedDoc) return
        const newName = renameValue.trim()
        if (!newName) {
            message.warning('文档名称不能为空')
            return
        }
        // optional: check duplicate name in docs
        if (docs.some((d) => d.room_id !== selectedDoc.room_id && d.room_name === newName)) {
            message.error('已有同名文档，请换个名字')
            return
        }
        try {
            const res = await axios.post('http://localhost:8000/mydocs/rename_room', {
                room_id: selectedDoc.room_id,
                room_name: newName,
            })

            if (res.data.success) {
                const newDocs = docs.map(d =>
                    d.room_id === selectedDoc.room_id
                        ? { ...d, room_name: newName }
                        : d
                )
                setDocs(newDocs)
                setRenameModalVisible(false)
                message.success('重命名成功')
            } else {
                message.error(res.data.message || '重命名失败')
            }
        } catch (err) {
            console.error('重命名失败', err)
            message.error('网络错误，重命名失败')
        }
    }

    // ----- delete document -----
    const handleDeleteDocument = async (docId: string) => {
        try {
            const res = await axios.post('http://localhost:8000/mydocs/delete_room', {
                room_id: docId,
            })

            if (res.data.success) {
                // 删除前端状态
                setDocs(prev => prev.filter(d => d.room_id !== docId))
                message.success('删除文档成功')

                // 如果删除的是当前选中 doc，切换选中
                setSelectedDocId(prev => {
                    if (prev !== docId) return prev
                    const remaining = docs.filter(d => d.room_id !== docId)
                    return remaining[0]?.room_id ?? null
                })
            } else {
                message.error(res.data.message || '删除文档失败')
            }
        } catch (err) {
            console.error('删除文档失败', err)
            message.error('网络错误，删除文档失败')
        }
        // setDocs((prev) => prev.filter((d) => d.room_id !== docId))
        // message.success('删除文档成功')
        // // pick another doc if any
        // setSelectedDocId((prev) => {
        //     if (prev !== docId) return prev
        //     const remaining = docs.filter((d) => d.room_id !== docId)
        //     return remaining[0]?.room_id ?? null
        // })
    }

    // ----- UI render helpers -----
    const menuItems: MenuProps['items'] = docs.map((d) => ({
        key: d.room_id,
        label: d.room_name,
    }))

    const permissionColumns = [
        {
            title: '用户',
            dataIndex: 'username',
            key: 'username',
            render: (_: any, record: any) => (
                <Space>
                    <Avatar style={{backgroundColor: '#87d068'}}>
                        {String(record.username || record.email || '?').charAt(0).toUpperCase()}
                    </Avatar>
                    <div>
                        <div>{record.username}</div>
                        <div style={{color: '#999', fontSize: 12}}>{record.email}</div>
                    </div>
                </Space>
            ),
        },
        {
            title: '权限',
            dataIndex: 'permission',
            key: 'permission',
            width: 240,
            render: (_: any, record: any) => (
                <Space>
                    <Button
                        size="small"
                        type={record.permission === 2 ? 'primary' : 'default'}
                        onClick={() => handleChangePermission(record.id, 2)}
                    >
                        可编辑
                    </Button>
                    <Button
                        size="small"
                        type={record.permission === 3 ? 'primary' : 'default'}
                        onClick={() => handleChangePermission(record.id, 3)}
                    >
                        只读
                    </Button>
                </Space>
            ),
        },
        {
            title: '操作',
            key: 'op',
            width: 120,
            render: (_: any, record: any) => (
                <Popconfirm
                    title={`确认移除 ${record.username} 的权限吗？`}
                    onConfirm={() => handleRemovePermission(record.id)}
                >
                    <Button danger size="small" icon={<DeleteOutlined/>}>
                        移除
                    </Button>
                </Popconfirm>
            ),
        },
    ]

    return (
        <Layout style={{height: '100%'}}>
            <Sider width={320} style={{background: '#fff'}}>
                <div style={{padding: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
                    <div style={{fontWeight: 600}}>我的文档</div>
                    <Button
                        icon={<PlusOutlined/>}
                        type="primary"
                        onClick={ async () => {
                            try {
                // 1. 调用后端创建文档接口
                                const res = await axios.post('http://localhost:8000/content/createdoc', {
                                    room_name: `新文档 ${docs.length + 1}`,
                                    user_id: userId
                                })

                                if (res.data.success) {
                                    // 2. 用后端返回的数据更新前端状态
                                    const newDoc: DocumentItem = {
                                        room_id: res.data.room_id,           // 后端生成的唯一 ID
                                        room_name: res.data.room_name,
                                        create_time: res.data.create_time,
                                        overall_permission: res.data.overall_permission,
                                        permissions: res.data.permissions || {},
                                    }

                                    setDocs(prev => [newDoc, ...prev])
                                    setSelectedDocId(newDoc.room_id)
                                    message.success('新文档创建成功')
                                } else {
                                    message.error(res.data.message || '创建失败')
                                }
                            } catch (err) {
                                console.error('创建文档失败', err)
                                message.error('网络错误，创建文档失败')
                            }
                        }}
                    >
                        新建
                    </Button>
                </div>

                <Menu
                    mode="inline"
                    selectedKeys={selectedDocId ? [selectedDocId] : []}
                    items={menuItems}
                    onClick={(info) => setSelectedDocId(String(info.key))}
                    style={{overflow: 'auto', borderRight: 0}}
                />
            </Sider>

            <Layout>
                <Content style={{padding: 24, minHeight: 280}}>
                    {!selectedDoc ? (
                        <Card>
                            <div>请选择左侧文档或创建新文档。</div>
                        </Card>
                    ) : (
                        <div style={{display: 'flex', gap: 16}}>
                            <div style={{flex: 1}}>
                                <Card
                                    title={selectedDoc.room_name}
                                    extra={
                                        <Space>
                                            <Button icon={<EditOutlined/>} onClick={openRenameModal}>
                                                重命名
                                            </Button>
                                        </Space>
                                    }
                                    style={{marginBottom: 16}}
                                >
                                    <div style={{marginBottom: 12}}>
                                        <Text strong>公开性</Text>
                                    </div>

                                    <Radio.Group
                                        onChange={(e) => handleVisibilityChange(e.target.value)}
                                        value={selectedDoc.overall_permission}
                                    >
                                        <Space direction="vertical">
                                            <Radio value={0}>私密（仅我可见）</Radio>
                                            <Radio value={1}>所有人可编辑</Radio>
                                            <Radio value={2}>所有人可阅读</Radio>
                                            <Radio value={3}>部分人可见 / 可编辑</Radio>
                                        </Space>
                                    </Radio.Group>

                                    {selectedDoc.overall_permission === 3 && (
                                        <div style={{marginTop: 20}}>
                                            <div style={{
                                                display: 'flex',
                                                justifyContent: 'space-between',
                                                alignItems: 'center',
                                                marginBottom: 12
                                            }}>
                                                <div>
                                                    <Text strong>已授权用户</Text>
                                                    <div
                                                        style={{color: '#888', fontSize: 12}}>以下用户有该文档的访问权限
                                                    </div>
                                                </div>

                                                <Space>
                                                    <Button icon={<UserAddOutlined/>} onClick={openAddUsers}>
                                                        添加用户
                                                    </Button>
                                                </Space>
                                            </div>

                                            <Table
                                                rowKey="userId"
                                                dataSource={listPermissions}
                                                columns={permissionColumns}
                                                pagination={false}
                                                locale={{emptyText: '当前没有已授权用户'}}
                                            />
                                        </div>
                                    )}
                                </Card>

                                <Card title="危险操作">
                                    <Space direction="vertical">
                                        <Popconfirm
                                            title="确认删除该文档？此操作无法撤销。"
                                            icon={<ExclamationCircleOutlined/>}
                                            onConfirm={() => handleDeleteDocument(selectedDoc.room_id)}
                                        >
                                            <Button danger icon={<DeleteOutlined/>}>
                                                删除文档
                                            </Button>
                                        </Popconfirm>
                                    </Space>
                                </Card>
                            </div>

                            {/* 右侧简要信息列 */}
                            <div style={{width: 320}}>
                                <Card title="文档信息">
                                    <div>
                                        <div><Text type="secondary">ID</Text></div>
                                        <div style={{marginBottom: 8}}>{selectedDoc.room_id}</div>

                                        <div><Text type="secondary">创建时间</Text></div>
                                        <div style={{marginBottom: 8}}>{selectedDoc.create_time}</div>

                                        <div><Text type="secondary">当前公开性</Text></div>
                                        <div style={{marginBottom: 8}}>
                                            <Tag>
                                                {selectedDoc.overall_permission === 0 && '私密'}
                                                {selectedDoc.overall_permission === 1 && '所有人可编辑'}
                                                {selectedDoc.overall_permission === 2 && '所有人可阅读'}
                                                {selectedDoc.overall_permission === 3 && '部分人可见'}
                                            </Tag>
                                        </div>
                                    </div>
                                </Card>

                                <Button type={'primary'} size="middle" icon={<LeftOutlined/>}
                                        onClick={() => navigate('/app')}
                                        style={{
                                            width: '100%',        // 横向占满
                                            marginTop: 16,        // 与 Card 的间距
                                        }}
                                >
                                    返回
                                </Button>
                            </div>
                        </div>
                    )}
                </Content>
            </Layout>

            {/* Add Users Modal */}
            <Modal
                title="添加用户并设置权限"
                open={addUserModalVisible}
                onCancel={closeAddUsers}
                onOk={handleAddUsersSubmit}
                cancelText="取消"
                okText={`添加（${selectedUsersInModal.length}）`}
                width={800}
            >
                <div style={{display: 'flex', gap: 16}}>
                    {/* left: search + user list */}
                    <div style={{width: 320}}>
                        <Search
                            placeholder="搜索用户名或邮箱"
                            value={userSearchText}
                            onChange={(e) => setUserSearchText(e.target.value)}
                            allowClear
                            style={{marginBottom: 12}}
                        />
                        <List
                            size="small"
                            bordered
                            style={{maxHeight: 400, overflow: 'auto'}}
                            dataSource={filteredUsersForModal}
                            renderItem={(user) => {
                                const selected = selectedUsersInModal.some((s) => s.user.id === user.id)
                                // disable if user already in doc permissions
                                const alreadyHas = selectedDoc ? !!selectedDoc.permissions[user.id] : false
                                return (
                                    <List.Item
                                        style={{
                                            cursor: alreadyHas ? 'not-allowed' : 'pointer',
                                            opacity: alreadyHas ? 0.5 : 1
                                        }}
                                        onClick={() => {
                                            if (alreadyHas) {
                                                message.info('该用户已包含在已授权列表中')
                                                return
                                            }
                                            toggleSelectUserInModal(user)
                                        }}
                                    >
                                        <List.Item.Meta
                                            avatar={<Avatar
                                                style={{backgroundColor: '#7265e6'}}>{user.user_name.charAt(0).toUpperCase()}</Avatar>}
                                            title={<div>{user.user_name}</div>}
                                            description={<div style={{fontSize: 12}}>{user.email}</div>}
                                        />
                                        <div>
                                            {selected ? <Tag>已选择</Tag> : null}
                                        </div>
                                    </List.Item>
                                )
                            }}
                        />
                    </div>

                    {/* right: selected users with permission toggles */}
                    <div style={{flex: 1}}>
                        <div style={{marginBottom: 12}}>
                            <Text strong>将要添加的用户</Text>
                        </div>

                        <List
                            bordered
                            dataSource={selectedUsersInModal}
                            locale={{emptyText: '未选择用户'}}
                            renderItem={(item) => (
                                <List.Item
                                    style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
                                    <List.Item.Meta
                                        avatar={<Avatar>{item.user.user_name.charAt(0).toUpperCase()}</Avatar>}
                                        title={item.user.user_name}
                                        description={<div style={{fontSize: 12}}>{item.user.email}</div>}
                                    />
                                    <div>
                                        <Space>
                                            <Button
                                                size="small"
                                                type={item.permission === 2 ? 'primary' : 'default'}
                                                onClick={() => setPermissionForSelectedUser(item.user.id, 2)}
                                            >
                                                可编辑
                                            </Button>
                                            <Button
                                                size="small"
                                                type={item.permission === 3 ? 'primary' : 'default'}
                                                onClick={() => setPermissionForSelectedUser(item.user.id, 3)}
                                            >
                                                只读
                                            </Button>
                                        </Space>
                                    </div>
                                </List.Item>
                            )}
                        />

                        {/*<div style={{ marginTop: 12, textAlign: 'right' }}>*/}
                        {/*  <Button onClick={() => { setSelectedUsersInModal([]) }} style={{ marginRight: 8 }}>*/}
                        {/*    清空*/}
                        {/*  </Button>*/}
                        {/*  <Button type="primary" onClick={handleAddUsersSubmit}>*/}
                        {/*    添加（{selectedUsersInModal.length}）*/}
                        {/*  </Button>*/}
                        {/*</div>*/}
                    </div>
                </div>
            </Modal>

            {/* Rename Modal */}
            <Modal
                title="重命名文档"
                open={renameModalVisible}
                onCancel={closeRenameModal}
                onOk={handleRenameSubmit}
                okText="保存"
                cancelText="取消"
            >
                <Input
                    value={renameValue}
                    onChange={(e) => setRenameValue(e.target.value)}
                    placeholder="输入新的文档名称"
                    maxLength={100}
                />
            </Modal>
        </Layout>
    )
}

export default MyDocsPage