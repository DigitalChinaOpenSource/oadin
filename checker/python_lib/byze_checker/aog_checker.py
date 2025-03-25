import os
import time
import requests
import subprocess
import threading
import http.server
import socketserver
from urllib.parse import urlparse, parse_qs

WEB_SERVER_PORT = 5000
Byze_DOWNLOAD_URL = "http://120.232.136.73:31619/byzedev/byze.exe"
Byze_FOLDER = os.path.join(os.path.expanduser("~"), "Byze")
Byze_PATH = os.path.join(Byze_FOLDER, "byze.exe")

user_response = None


# 检查 Byze 服务器是否可用
def is_byze_available():
    try:
        response = requests.get("http://localhost:16688", timeout=3)
        return response.status_code == 200
    except requests.RequestException:
        return False


# 启动 Web 服务器，提供安装确认界面
class InstallPromptHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        global user_response
        parsed_path = urlparse(self.path)
        if parsed_path.path == "/install-prompt":
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.end_headers()
            self.wfile.write(INSTALL_PROMPT_HTML.encode("utf-8"))
        elif parsed_path.path == "/user-response":
            query = parse_qs(parsed_path.query)
            user_response = query.get("choice", ["false"])[0] == "true"
            self.send_response(200)
            self.end_headers()


def start_web_server():
    server = socketserver.TCPServer(("0.0.0.0", WEB_SERVER_PORT), InstallPromptHandler)
    threading.Thread(target=server.serve_forever, daemon=True).start()


# 在浏览器中打开 URL
def open_browser(url):
    if os.name == "nt":
        os.system(f'start {url}')
    elif os.uname().sysname == "Darwin":
        os.system(f'open {url}')
    else:
        os.system(f'xdg-open {url}')


# 下载 Byze.exe
def download_byze():
    os.makedirs(Byze_FOLDER, exist_ok=True)
    try:
        response = requests.get(Byze_DOWNLOAD_URL, stream=True)
        with open(Byze_PATH, "wb") as f:
            for chunk in response.iter_content(1024):
                f.write(chunk)
        return True
    except requests.RequestException:
        return False


# 启动 Byze 服务器
def install_byze():
    try:
        subprocess.Popen([Byze_PATH, "server", "start", "-d"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        time.sleep(3)  # 等待 Byze 启动
        return True
    except Exception:
        return False


# 导入 Byze 文件
def import_byze_file(byze_file_path):
    try:
        result = subprocess.run([Byze_PATH, "import", "--file", byze_file_path], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        return result.returncode == 0
    except Exception:
        return False


# HTML 页面
INSTALL_PROMPT_HTML = """
<html>
<body style="padding:20px;font-family:Arial">
    <h2>安装确认</h2>
    <p>需要安装 Byze 组件才能继续，是否允许？</p>
    <button onclick="respond(true)">同意安装</button>
    <button onclick="respond(false)">取消</button>
    <script>
        function respond(choice) {
            fetch('/user-response?choice=' + choice)
                .then(() => window.close());
        }
    </script>
</body>
</html>
"""


# 主入口
def ByzeInit(byze_file_path=None):
    global user_response
    if byze_file_path is None:
        byze_file_path = os.path.join(os.getcwd(), ".byze")

    if is_byze_available():
        print("Byze 已经运行，直接导入 .byze 文件...")
        import_byze_file(byze_file_path)
        return

    print("Byze 未运行，启动安装流程...")
    start_web_server()
    open_browser(f"http://localhost:{WEB_SERVER_PORT}/install-prompt")

    # 等待用户输入，最多 5 分钟
    start_time = time.time()
    while user_response is None and time.time() - start_time < 300:
        time.sleep(1)

    if not user_response:
        print("用户拒绝安装，退出。")
        return

    if not download_byze():
        print("Byze 下载失败，退出。")
        return

    if not install_byze():
        print("Byze 启动失败，退出。")
        return

    print("Byze 启动成功，导入 .byze 文件...")
    if import_byze_file(byze_file_path):
        print(f"成功导入 {byze_file_path}")
    else:
        print(f"导入失败: {byze_file_path}")


# 让 Python 直接 `import byzechecker` 即可调用
if __name__ == "__main__":
    ByzeInit()
