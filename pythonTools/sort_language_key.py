import json
import os

def reorder_translations():
    # 文件路径
    en_file_path = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales\zh.json"
    zh_file_path = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales\en.json"
    
    # 检查文件是否存在
    if not os.path.exists(en_file_path):
        print(f"错误: 英文翻译文件不存在: {en_file_path}")
        return
    
    if not os.path.exists(zh_file_path):
        print(f"错误: 中文翻译文件不存在: {zh_file_path}")
        return
    
    try:
        # 读取英文翻译文件
        with open(en_file_path, 'r', encoding='utf-8') as f:
            en_data = json.load(f)
        
        # 读取中文翻译文件
        with open(zh_file_path, 'r', encoding='utf-8') as f:
            zh_data = json.load(f)
        
        print(f"英文翻译文件中的键数量: {len(en_data)}")
        print(f"中文翻译文件中的键数量: {len(zh_data)}")
        
        # 去除重复键（如果有的话）
        unique_zh_data = {}
        for key, value in zh_data.items():
            if key not in unique_zh_data:
                unique_zh_data[key] = value
        
        # 按照英文文件的键顺序重新排列中文翻译
        ordered_zh_data = {}
        missing_keys = []
        extra_keys = set(unique_zh_data.keys()) - set(en_data.keys())
        
        # 首先按照英文文件的顺序添加键值对
        for key in en_data.keys():
            if key in unique_zh_data:
                ordered_zh_data[key] = unique_zh_data[key]
            else:
                # 如果中文翻译中没有该键，记录下来并使用英文值作为默认值
                missing_keys.append(key)
                ordered_zh_data[key] = en_data[key]
        
        # 然后添加中文文件中存在但英文文件中不存在的键值对（如果有的话）
        for key in extra_keys:
            ordered_zh_data[key] = unique_zh_data[key]
        
        # 将重新排序后的中文翻译写回文件
        with open(zh_file_path, 'w', encoding='utf-8') as f:
            json.dump(ordered_zh_data, f, ensure_ascii=False, indent=2)
        
        print(f"中文翻译文件已按照英文文件的键顺序重新排列")
        
        if missing_keys:
            print(f"发现 {len(missing_keys)} 个在英文文件中存在但在中文文件中不存在的键:")
            for key in missing_keys:
                print(f"  - {key}")
            print("这些键已添加到中文文件中，并使用英文值作为默认值")
        
        if extra_keys:
            print(f"发现 {len(extra_keys)} 个在中文文件中存在但在英文文件中不存在的键:")
            for key in extra_keys:
                print(f"  - {key}")
            print("这些键已保留在中文文件中")
        
        if len(unique_zh_data) < len(zh_data):
            print(f"已从中文文件中移除 {len(zh_data) - len(unique_zh_data)} 个重复的键")
        
    except json.JSONDecodeError as e:
        print(f"JSON解析错误: {e}")
    except Exception as e:
        print(f"处理文件时出错: {e}")

if __name__ == "__main__":
    reorder_translations()