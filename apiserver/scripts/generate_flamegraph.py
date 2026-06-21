#!/usr/bin/env python3
"""
火焰图生成脚本 (本地优先版)
优先使用本地文件，如果不存在则从 MinIO 下载
"""
import sys
import os
import subprocess
from minio import Minio

MINIO_ENDPOINT = "localhost:9000"
MINIO_ACCESS_KEY = "minioadmin"
MINIO_SECRET_KEY = "minioadmin123"
MINIO_BUCKET = "drop-data"
MINIO_SECURE = False

FLAMEGRAPH_DIR = os.path.expanduser("~/FlameGraph")
FLAMEGRAPH_PL = os.path.join(FLAMEGRAPH_DIR, "flamegraph.pl")
STACKCOLLAPSE_PL = os.path.join(FLAMEGRAPH_DIR, "stackcollapse-perf.pl")

def find_local_perf_data(task_id):
    local_paths = [
        f"/tmp/perf_{task_id}.data",
        f"/tmp/{task_id}.perf.data",
        f"/home/lwl/perf_{task_id}.data",
    ]
    for path in local_paths:
        if os.path.exists(path) and os.path.getsize(path) > 0:
            print(f"✅ 使用本地文件: {path}")
            return path
    return None

def download_perf_data(task_id):
    client = Minio(
        MINIO_ENDPOINT,
        access_key=MINIO_ACCESS_KEY,
        secret_key=MINIO_SECRET_KEY,
        secure=MINIO_SECURE
    )
    object_name = f"{task_id}/perf.data"
    local_file = f"/tmp/{task_id}.perf.data"

    try:
        client.fget_object(MINIO_BUCKET, object_name, local_file)
        print(f"✅ 从 MinIO 下载成功: {object_name}")
        return local_file
    except Exception as e:
        print(f"❌ 从 MinIO 下载失败: {e}")
        return None

def generate_flamegraph(perf_data_file, task_id):
    script_file = f"/tmp/{task_id}.script"
    cmd = f"sudo perf script -i {perf_data_file} -f > {script_file}"
    print(f"📊 执行: {cmd}")
    result = subprocess.run(cmd, shell=True, capture_output=True)
    if result.returncode != 0:
        print(f"❌ perf script 失败: {result.stderr.decode()}")
        return None

    if not os.path.exists(script_file) or os.path.getsize(script_file) < 100:
        print(f"❌ 脚本文件太小或不存在")
        return None

    with open(script_file, 'r') as f:
        lines = f.readlines()
        print(f"📊 脚本行数: {len(lines)}")

    collapsed_file = f"/tmp/{task_id}.collapsed"
    cmd = f"cat {script_file} | {STACKCOLLAPSE_PL} > {collapsed_file}"
    print(f"📊 执行: {cmd}")
    subprocess.run(cmd, shell=True, check=True)

    svg_file = f"/tmp/{task_id}.svg"
    cmd = f"cat {collapsed_file} | {FLAMEGRAPH_PL} > {svg_file}"
    print(f"📊 执行: {cmd}")
    subprocess.run(cmd, shell=True, check=True)

    file_size = os.path.getsize(svg_file)
    print(f"✅ 火焰图生成成功: {svg_file} ({file_size} bytes)")
    return svg_file

def upload_svg(task_id, svg_file):
    client = Minio(
        MINIO_ENDPOINT,
        access_key=MINIO_ACCESS_KEY,
        secret_key=MINIO_SECRET_KEY,
        secure=MINIO_SECURE
    )
    object_name = f"{task_id}/flamegraph.svg"
    try:
        client.fput_object(MINIO_BUCKET, object_name, svg_file)
        print(f"✅ 上传成功: {object_name}")
        return True
    except Exception as e:
        print(f"❌ 上传失败: {e}")
        return False

def main():
    if len(sys.argv) != 2:
        print("用法: python3 generate_flamegraph.py <task_id>")
        sys.exit(1)

    task_id = sys.argv[1]
    print(f"🔥 开始生成火焰图: {task_id}")

    perf_data = find_local_perf_data(task_id)
    if not perf_data:
        perf_data = download_perf_data(task_id)

    if not perf_data:
        sys.exit(1)

    svg_file = generate_flamegraph(perf_data, task_id)
    if not svg_file:
        sys.exit(1)

    upload_svg(task_id, svg_file)

    print(f"\n✅ 完成！")
    print(f"📁 本地文件: {svg_file}")
    print(f"🌐 浏览器访问: /tmp/{task_id}.svg")

if __name__ == "__main__":
    main()
