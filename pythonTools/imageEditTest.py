import requests
import json
import time

url = "https://devopenai.burncloud.com/v1/images/edits"

headers = {
    "Authorization": "Bearer sk-7wiO34EMbl9hvC8n9ev2Y0CPwO7e5H2tKMlzl8U253B9a342",
    "Cookie": "p_uv_id=8f067df6e1f4e00e11bee02949c1b070"
    # 注意：不要手动设置Content-Type，让requests库自动处理
}

# 表单数据
form_data = {
    "model": "gpt-image-1",
    "n": "1",
    "prompt": "请将图像转为星露谷风格，保留原图的细节 色彩控制：白平衡中性（5500K-6000k），去黄、提升亮度、增强色彩层次。",
    "quality": "medium",
    "size": "1536x1024"
}

# 图片文件路径
image_path = r"C:\Users\Administrator\Desktop\网站设计\GPU_hosting_data_center.png"

try:
    # 打开文件并创建文件对象
    with open(image_path, "rb") as img_file:
        files = {
            "image": ("GPU_hosting_data_center.png", img_file, "image/png")
        }
        
        # 发送请求，增加超时和重试机制
        try:
            response = requests.post(
                url, 
                headers=headers, 
                data=form_data, 
                files=files,
                timeout=60,  # 增加超时时间
                stream=False  # 禁用流式传输
            )
            
            # 输出状态码
            print(f"Status Code: {response.status_code}")
            
            # 尝试解析JSON响应
            try:
                result = response.json()
                print(json.dumps(result, indent=2, ensure_ascii=False))
            except json.JSONDecodeError:
                # 如果不是JSON格式，直接打印响应内容
                print("Response is not JSON. Raw content:")
                print(response.text[:1000])  # 只打印前1000个字符
                
        except requests.exceptions.RequestException as e:
            print(f"Request failed: {e}")
            
except FileNotFoundError:
    print(f"Error: File not found at path {image_path}")
except Exception as e:
    print(f"Unexpected error: {e}")
