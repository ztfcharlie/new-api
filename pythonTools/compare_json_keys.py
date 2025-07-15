import json
import os

def load_json_file(file_path):
    """加载JSON文件并返回数据"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            return json.load(f)
    except Exception as e:
        print(f"读取文件 {file_path} 时出错: {str(e)}")
        return None

def collect_all_keys(data, prefix="", result=None):
    """递归收集JSON中所有的键，包括嵌套结构"""
    if result is None:
        result = {}
    
    if isinstance(data, dict):
        for key, value in data.items():
            current_key = f"{prefix}.{key}" if prefix else key
            
            if isinstance(value, dict):
                # 递归处理嵌套字典
                collect_all_keys(value, current_key, result)
            else:
                # 添加叶节点键
                result[current_key] = value
    
    return result

def compare_json_files(file1_path, file2_path, file1_name, file2_name):
    """比较两个JSON文件的键值差异"""
    # 加载两个JSON文件
    data1 = load_json_file(file1_path)
    data2 = load_json_file(file2_path)
    
    if data1 is None or data2 is None:
        return
    
    # 收集所有键
    keys1 = collect_all_keys(data1)
    keys2 = collect_all_keys(data2)
    
    # 找出文件1比文件2多出的键
    keys_only_in_1 = {k: v for k, v in keys1.items() if k not in keys2}
    
    # 找出文件2比文件1多出的键
    keys_only_in_2 = {k: v for k, v in keys2.items() if k not in keys1}
    
    # 输出结果
    print(f"\n{file1_name} 比 {file2_name} 多出 {len(keys_only_in_1)} 个键:")
    if keys_only_in_1:
        for i, (key, value) in enumerate(keys_only_in_1.items(), 1):
            print(f"{i}. {key}: {repr(value)}")
    else:
        print("无")
    
    print(f"\n{file2_name} 比 {file1_name} 多出 {len(keys_only_in_2)} 个键:")
    if keys_only_in_2:
        for i, (key, value) in enumerate(keys_only_in_2.items(), 1):
            print(f"{i}. {key}: {repr(value)}")
    else:
        print("无")
    
    # 将结果保存到文件
    result = {
        f"{file1_name}_unique_keys": keys_only_in_1,
        f"{file2_name}_unique_keys": keys_only_in_2
    }
    
    output_file = "json_keys_comparison_result.json"
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(result, f, ensure_ascii=False, indent=2)
    
    print(f"\n比较结果已保存到: {output_file}")

if __name__ == "__main__":
    # 文件路径
    zh_file_path = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales\es.json"
    en_file_path = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales\en.json"
    
    # 确保文件存在
    if not os.path.exists(zh_file_path):
        print(f"错误：找不到中文翻译文件: {zh_file_path}")
    elif not os.path.exists(en_file_path):
        print(f"错误：找不到英文翻译文件: {en_file_path}")
    else:
        print("开始比较两个JSON文件的键值差异...")
        compare_json_files(zh_file_path, en_file_path, "中文文件", "英文文件")