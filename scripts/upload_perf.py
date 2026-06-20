#!/usr/bin/env python3
import sys
import os
from minio import Minio

def upload_file(bucket, object_name, file_path):
    client = Minio(
        "localhost:9000",
        access_key="minioadmin",
        secret_key="minioadmin123",
        secure=False
    )
    
    # 如果 bucket 不存在则创建
    if not client.bucket_exists(bucket):
        client.make_bucket(bucket)
        print(f"创建 bucket: {bucket}")
    
    client.fput_object(bucket, object_name, file_path)
    print(f"上传成功: {bucket}/{object_name}")

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("用法: python3 upload_perf.py <bucket> <object_name> <file_path>")
        sys.exit(1)
    
    upload_file(sys.argv[1], sys.argv[2], sys.argv[3])
