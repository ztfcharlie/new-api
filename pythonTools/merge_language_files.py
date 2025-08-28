import json
import os
import copy

def load_json_file(file_path):
    """加载JSON文件并返回数据"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            return json.load(f)
    except Exception as e:
        print(f"读取文件 {file_path} 时出错: {str(e)}")
        return None

def find_new_keys(original_data, extended_data, prefix="", result=None):
    """递归查找扩展文件中相对于原始文件新增的键值对"""
    if result is None:
        result = {}
    
    # 遍历扩展文件中的所有键值对
    for key, value in extended_data.items():
        # 构建当前键的完整路径
        current_path = f"{prefix}.{key}" if prefix else key
        
        # 如果键在原始数据中不存在
        if key not in original_data:
            # 如果值是字典，则递归添加整个子树
            if isinstance(value, dict):
                nested_dict = {}
                collect_nested_dict(value, nested_dict)
                result[key] = nested_dict
            else:
                # 直接添加键值对
                result[key] = value
        # 如果键在原始数据中存在，且两边都是字典，则递归比较
        elif isinstance(value, dict) and isinstance(original_data[key], dict):
            nested_result = {}
            find_new_keys(original_data[key], value, current_path, nested_result)
            if nested_result:
                result[key] = nested_result
    
    return result

def collect_nested_dict(source_dict, target_dict):
    """复制整个嵌套字典结构"""
    for key, value in source_dict.items():
        if isinstance(value, dict):
            target_dict[key] = {}
            collect_nested_dict(value, target_dict[key])
        else:
            target_dict[key] = value

def merge_json_files(original_file_path, extended_file_path, output_file_path):
    """合并原始文件和扩展文件，生成新文件"""
    # 加载两个JSON文件
    original_data = load_json_file(original_file_path)
    extended_data = load_json_file(extended_file_path)
    
    if original_data is None or extended_data is None:
        return False
    
    # 创建原始数据的深拷贝作为新文件的基础
    new_data = copy.deepcopy(original_data)
    
    # 查找扩展文件中新增的键值对
    new_keys = find_new_keys(original_data, extended_data)
    
    # 将新增的键值对添加到新数据中
    for key, value in new_keys.items():
        new_data[key] = value
    
    # 将结果写入输出文件
    try:
        with open(output_file_path, 'w', encoding='utf-8') as f:
            json.dump(new_data, f, ensure_ascii=False, indent=2)
        print(f"合并完成，新文件已保存到: {output_file_path}")
        
        # 输出统计信息
        print(f"原始文件键数量: {count_keys(original_data)}")
        print(f"扩展文件键数量: {count_keys(extended_data)}")
        print(f"新增键数量: {count_keys(new_keys)}")
        print(f"新文件键数量: {count_keys(new_data)}")
        return True
    except Exception as e:
        print(f"保存文件时出错: {str(e)}")
        return False

def count_keys(data, prefix=""):
    """计算JSON中所有键的数量，包括嵌套结构"""
    count = 0
    if isinstance(data, dict):
        for key, value in data.items():
            current_key = f"{prefix}.{key}" if prefix else key
            if isinstance(value, dict):
                count += count_keys(value, current_key)
            else:
                count += 1
    return count

if __name__ == "__main__":
    # 文件路径
    original_file_path = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales\en-origin.json"
    extended_file_path = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales\en.json"
    output_file_path = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales\en-new.json"
    
    # 确保文件存在
    if not os.path.exists(original_file_path):
        print(f"错误：找不到原始语言文件: {original_file_path}")
    elif not os.path.exists(extended_file_path):
        print(f"错误：找不到扩展语言文件: {extended_file_path}")
    else:
        print("开始合并语言文件...")
        merge_json_files(original_file_path, extended_file_path, output_file_path)