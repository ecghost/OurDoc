import socket

host = "og2"       # 目标容器名
port = 5432        # openGauss 默认端口

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.settimeout(5)  # 超时 5 秒

try:
    sock.connect((host, port))
    print(f"SUCCESS: Can connect to {host}:{port}")
except Exception as e:
    print(f"FAILED: Cannot connect to {host}:{port}, reason: {e}")
finally:
    sock.close()
