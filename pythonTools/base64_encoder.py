import base64
import argparse

def file_to_base64(file_path, output_path=None):
    with open(file_path, 'rb') as file:
        encoded_content = base64.b64encode(file.read()).decode('utf-8')
        
    if output_path:
        with open(output_path, 'w') as output_file:
            output_file.write(encoded_content)
        print(f"Base64 编码已保存到 {output_path}")
    else:
        print(encoded_content)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="将文件转换为 Base64 编码")
    parser.add_argument("file", help="要编码的文件路径")
    parser.add_argument("-o", "--output", help="输出文件路径 (可选)")
    
    args = parser.parse_args()
    file_to_base64(args.file, args.output)
