import json
import os

def transform_json(input_path, output_filename):
    # 读取原始 JSON 文件
    with open(input_path, 'r', encoding='utf-8') as infile:
        data = json.load(infile)

    # 创建新的字典，键为原始键，值为键
    transformed_data = {key: key for key in data.keys()}

    # 获取原始文件的目录
    output_path = os.path.join(os.path.dirname(input_path), output_filename)

    # 将转换后的数据写入新的 JSON 文件
    with open(output_path, 'w', encoding='utf-8') as outfile:
        json.dump(transformed_data, outfile, ensure_ascii=False, indent=4)

    print(f"转换完成，文件保存为: {output_path}")

if __name__ == "__main__":
    # 输入原始语言文件的路径和保存的文件名
    input_file_path = input("请输入原语言文件的路径：")
    output_file_name = input("请输入要保存的文件名：")

    transform_json(input_file_path, output_file_name)
