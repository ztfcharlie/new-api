import json
import os
import collections
import re

def parse_json_preserve_duplicates(content):
    """
    解析JSON内容，同时保留重复键的信息
    
    Args:
        content: JSON文件内容
    
    Returns:
        data: 解析后的JSON数据
        duplicates: 包含重复键的信息
    """
    # 使用正则表达式查找所有键值对
    pattern = r'"([^"]+)"\s*:\s*("(?:\\.|[^"\\])*"|[^,}\]]*)'
    matches = re.findall(pattern, content)
    
    # 构建键值映射，同时记录重复键
    key_values = {}
    duplicates = collections.defaultdict(list)
    
    for key, value in matches:
        if key in key_values:
            duplicates[key].append(value)
        key_values[key] = value
    
    # 过滤出实际重复的键
    duplicates = {k: v for k, v in duplicates.items() if v}
    
    # 加载JSON数据
    data = json.loads(content)
    
    return data, duplicates

def find_and_remove_duplicates(file_path):
    """
    查找并移除JSON文件中的重复键
    
    Args:
        file_path: JSON文件路径
    
    Returns:
        cleaned_data: 移除重复键后的JSON数据
        duplicates_found: 找到的重复键信息
    """
    # 读取文件内容
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 查找文件中的重复键
    duplicates_found = {}
    
    # 方法1: 使用正则表达式检查明显的重复键
    pattern = r'"([^"]+)"\s*:'
    keys = re.findall(pattern, content)
    
    # 检查重复键
    key_counts = collections.Counter(keys)
    simple_duplicates = {k: v for k, v in key_counts.items() if v > 1}
    
    if simple_duplicates:
        print("发现重复键:")
        for key, count in simple_duplicates.items():
            print(f"键 '{key}' 出现了 {count} 次")
        duplicates_found.update(simple_duplicates)
    
    # 方法2: 递归检查并清理嵌套结构中的重复键
    try:
        # 加载JSON数据，同时保留重复键信息
        data, more_duplicates = parse_json_preserve_duplicates(content)
        
        if more_duplicates:
            for key, values in more_duplicates.items():
                if key not in duplicates_found:
                    duplicates_found[key] = len(values) + 1
        
        # 递归清理嵌套结构中的重复键
        cleaned_data = clean_nested_duplicates(data)
        
        return cleaned_data, duplicates_found
    
    except json.JSONDecodeError as e:
        print(f"JSON解析错误: {str(e)}")
        # 尝试定位错误位置附近的内容
        error_pos = e.pos
        context_start = max(0, error_pos - 50)
        context_end = min(len(content), error_pos + 50)
        context = content[context_start:context_end]
        print(f"错误位置附近的内容: \n{context}")
        return None, duplicates_found

def clean_nested_duplicates(data):
    """
    递归清理嵌套结构中的重复键
    
    Args:
        data: JSON数据
    
    Returns:
        cleaned_data: 清理后的JSON数据
    """
    if isinstance(data, dict):
        # 创建一个新的字典，保留每个键的最后一个值
        cleaned_dict = {}
        for key, value in data.items():
            # 递归处理嵌套值
            if isinstance(value, (dict, list)):
                cleaned_dict[key] = clean_nested_duplicates(value)
            else:
                cleaned_dict[key] = value
        return cleaned_dict
    
    elif isinstance(data, list):
        # 递归处理列表中的每个元素
        return [clean_nested_duplicates(item) for item in data]
    
    else:
        # 基本类型直接返回
        return data

def process_json_file(file_path):
    """
    处理JSON文件，移除重复键并保存回原文件
    
    Args:
        file_path: JSON文件路径
    """
    print(f"正在处理文件: {file_path}")
    
    if not os.path.exists(file_path):
        print(f"错误: 文件不存在")
        return
    
    try:
        # 创建备份文件
        backup_file = f"{file_path}.bak"
        with open(file_path, 'r', encoding='utf-8') as f:
            original_content = f.read()
        
        with open(backup_file, 'w', encoding='utf-8') as f:
            f.write(original_content)
        print(f"已创建备份文件: {backup_file}")
        
        # 查找并移除重复键
        cleaned_data, duplicates = find_and_remove_duplicates(file_path)
        
        if cleaned_data is None:
            print("处理失败，未能正确解析JSON文件")
            return
        
        if not duplicates:
            print("未发现重复的键，文件保持不变")
            return
        
        # 将清理后的数据保存回原文件
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(cleaned_data, f, ensure_ascii=False, indent=2)
        
        print(f"\n处理完成: 已移除 {len(duplicates)} 个重复的键")
        print(f"清理后的文件已保存回: {file_path}")
        print(f"原始文件已备份为: {backup_file}")
        
        # 将重复键信息保存到报告文件
        report_file = f"{os.path.splitext(file_path)[0]}_duplicates_report.json"
        with open(report_file, 'w', encoding='utf-8') as f:
            json.dump(duplicates, f, ensure_ascii=False, indent=2)
        print(f"重复键详细报告已保存到: {report_file}")
        
    except Exception as e:
        print(f"处理过程中出错: {str(e)}")

if __name__ == "__main__":
    # 文件路径
    json_file_path = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales\zh.json"
    process_json_file(json_file_path)