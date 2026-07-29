#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
FastToken 技能安装器 (install-fasttoken-skills.py)

从 https://openfasttoken.example 下载 FastToken 技能压缩包并解压。

仅依赖 Python 标准库，跨平台（Windows / macOS / Linux）。

用法:
    python3 install-fasttoken-skills.py
    python3 install-fasttoken-skills.py --url https://openfasttoken.example/skills/fasttoken-skills.zip
    python3 install-fasttoken-skills.py --insecure      # 跳过 TLS 证书校验

环境变量:
    FASTTOKEN_SKILLS_DIR   指定技能目录（绝对路径）；设置后自动装入，否则解压到当前目录
    FASTTOKEN_SKILLS_URL   覆盖下载地址
"""
import os
import sys
import ssl
import zipfile
import tempfile
import shutil
import urllib.request

DEFAULT_URL = "https://openfasttoken.example/skills/fasttoken-skills.zip"


def parse_args(argv):
    opts = {"url": None, "insecure": False}
    i = 1
    while i < len(argv):
        a = argv[i]
        if a in ("-h", "--help"):
            print(__doc__)
            sys.exit(0)
        elif a == "--url":
            i += 1
            opts["url"] = argv[i]
        elif a == "--insecure":
            opts["insecure"] = True
        elif a.startswith("--url="):
            opts["url"] = a.split("=", 1)[1]
        else:
            sys.stderr.write("未知参数: %s\n" % a)
            sys.exit(2)
        i += 1
    return opts


def download(url, dest, insecure):
    ctx = ssl.create_default_context()
    if insecure:
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE
    req = urllib.request.Request(
        url, headers={"User-Agent": "fasttoken-skills-installer/1.0"}
    )
    with urllib.request.urlopen(req, context=ctx, timeout=120) as resp:
        total = int(resp.headers.get("Content-Length", "0") or "0")
        downloaded = 0
        with open(dest, "wb") as f:
            while True:
                buf = resp.read(8192)
                if not buf:
                    break
                f.write(buf)
                downloaded += len(buf)
                if total:
                    sys.stdout.write("\r  下载进度: %3d%%" % (downloaded * 100 // total))
                    sys.stdout.flush()
        if total:
            sys.stdout.write("\n")
    return downloaded


def main():
    opts = parse_args(sys.argv)
    url = opts["url"] or os.environ.get("FASTTOKEN_SKILLS_URL", DEFAULT_URL)

    print("FastToken 技能安装器")
    print("  下载地址 : %s" % url)

    auto_dir = os.environ.get("FASTTOKEN_SKILLS_DIR")
    tmp = tempfile.mkdtemp(prefix="ft-skills-")
    try:
        zip_path = os.path.join(tmp, "fasttoken-skills.zip")
        print("  正在下载压缩包 ...")
        size = download(url, zip_path, opts["insecure"])
        if size == 0:
            sys.stderr.write("  下载失败：未获取到文件内容。\n")
            return 1
        print("  正在解压 (%d 字节) ..." % size)
        with zipfile.ZipFile(zip_path) as z:
            if auto_dir:
                os.makedirs(auto_dir, exist_ok=True)
                z.extractall(auto_dir)
                installed = sorted(
                    {n.split("/", 1)[0] for n in z.namelist() if "/" in n and not n.startswith("/")}
                )
                print("  完成。已安装到技能目录:")
                for n in installed:
                    print("    - %s/" % n)
            else:
                staging = os.path.join(os.getcwd(), "fasttoken-skills")
                if os.path.exists(staging):
                    shutil.rmtree(staging, ignore_errors=True)
                z.extractall(staging)
                print("  完成。已解压到: %s" % staging)
                print("  请将其中的 fasttoken/ 文件夹移动到你的技能目录，然后重启助手。")
    except Exception as e:  # noqa: BLE001
        sys.stderr.write("  安装失败: %s\n" % e)
        return 1
    finally:
        shutil.rmtree(tmp, ignore_errors=True)

    if auto_dir:
        print("请重启助手以加载新技能。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
