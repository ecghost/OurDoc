# 程序编译部署说明
本次程序需要部署五个主要环境：
React、Go、FastAPI、Opengauss Docker、gsql

下面将分别展示环境部署细节。

## 运行环境
Windows下Wsl环境，配置Ubuntu20.04。

## React环境
进入代码根目录，执行:
```
npm install
```
配置React相关环境。

同时可以通过如下指令：
```
bash bash_frontend.sh
```
运行前端服务，默认端口为5173。

## FastAPI环境

本次课程报告通过Anaconda配置。

通过如下命令在根目录创建虚拟环境并配置FastAPI相关环境：
```
conda create -n ourdoc python==3.10.19 --y
conda activate ourdoc
pip install -r requirement.txt
```

通过如下指令运行后端服务：
```
cd src/backend
bash bash_backend.sh
```
默认端口为9000。

## Go相关环境
配置正常Go的运行环境即可

## Opengauss Docker
下载openGauss-Docker-7.0.0-RC1-x86_64.tar

并且确定new/my-gauss-demo/docker-compose.yml中的内容是否指向下载的tar文件

运行docker-compose up --build

即可配置Opengauss整体运行环境，
通过如下指令运行数据库的服务端口：
```
cd src/backend/my-gauss-demo/app
bash bash_database.sh
```
默认端口为8080。

## gsql环境配置
配置下载目录，配置opengauss：
```
wget https://obs.cn-north-1.myhuaweicloud.com/dws/download/dws_client_9.1.x_redhat_x64.zip
unzip dws_client_9.1.x_redhat_x64.zip
source gsql_env.sh
```
即可完成配置。

我们的两个数据库分别保存在5433和5432端口，分别运行
```
gsql -h localhost -p (端口号) -U gaussdb -d postgres
```
并输入密码：Sakura030523!

即可连接数据库。
