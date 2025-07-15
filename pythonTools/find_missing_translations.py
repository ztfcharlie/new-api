import json
import os

def find_missing_translations(en_file_path, zh_file_path, output_file_path):
    """
    查找存在于英文翻译文件但不存在于中文翻译文件的键值对
    
    Args:
        en_file_path: 英文翻译文件路径
        zh_file_path: 中文翻译文件路径
        output_file_path: 输出结果文件路径
    """
    # 读取英文翻译文件
    with open(en_file_path, 'r', encoding='utf-8') as f:
        en_data = json.load(f)
    
    # 读取中文翻译文件
    with open(zh_file_path, 'r', encoding='utf-8') as f:
        zh_data = json.load(f)
    
    # 递归查找嵌套字典中的差异
    def find_diff(en_dict, zh_dict, current_path="", result=None):
        if result is None:
            result = {}
            
        for key, value in en_dict.items():
            new_path = f"{current_path}.{key}" if current_path else key
            
            # 如果键不存在于中文翻译中
            if key not in zh_dict:
                # 将值设置为空字符串
                if isinstance(value, dict):
                    # 如果是嵌套字典，递归处理
                    nested_result = {}
                    collect_keys(value, "", nested_result)
                    for nested_key, _ in nested_result.items():
                        full_key = f"{new_path}.{nested_key}" if nested_key else new_path
                        result[full_key] = ""
                else:
                    result[new_path] = ""
            # 如果两者都是字典，递归检查
            elif isinstance(value, dict) and isinstance(zh_dict[key], dict):
                find_diff(value, zh_dict[key], new_path, result)
            # 如果键存在但值为空
            elif key in zh_dict and (zh_dict[key] == "" or zh_dict[key] is None):
                result[new_path] = ""
                
        return result
    
    # 收集嵌套字典中的所有键
    def collect_keys(d, current_path="", result=None):
        if result is None:
            result = {}
            
        for key, value in d.items():
            new_path = f"{current_path}.{key}" if current_path else key
            
            if isinstance(value, dict):
                collect_keys(value, new_path, result)
            else:
                result[new_path] = ""
                
        return result
    
    # 查找差异
    missing_translations = find_diff(en_data, zh_data)
    
    # 将结果写入输出文件
    with open(output_file_path, 'w', encoding='utf-8') as f:
        json.dump(missing_translations, f, ensure_ascii=False, indent=2)
    
    print(f"找到 {len(missing_translations)} 个缺失的翻译，已保存到 {output_file_path}")
    return missing_translations

if __name__ == "__main__":
    # 文件路径
    en_file_path = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales\en.json"
    zh_file_path = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales\de.json"
    output_file_path = r"missing_translations.json"
    
    # 确保文件存在
    if not os.path.exists(en_file_path):
        print(f"错误：找不到英文翻译文件: {en_file_path}")
    elif not os.path.exists(zh_file_path):
        print(f"错误：找不到中文翻译文件: {zh_file_path}")
    else:
        find_missing_translations(en_file_path, zh_file_path, output_file_path)