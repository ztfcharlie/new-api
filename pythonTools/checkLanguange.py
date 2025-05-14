import json
import os

def process_translation_files():
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
        
        # 去除英文翻译中的重复键
        unique_en_data = {}
        duplicate_en_keys = []
        for key, value in en_data.items():
            if key not in unique_en_data:
                unique_en_data[key] = value
            else:
                duplicate_en_keys.append(key)
        
        # 去除中文翻译中的重复键
        unique_zh_data = {}
        duplicate_zh_keys = []
        for key, value in zh_data.items():
            if key not in unique_zh_data:
                unique_zh_data[key] = value
            else:
                duplicate_zh_keys.append(key)
        
        # 找出在英文文件中存在但在中文文件中不存在的键值对
        missing_translations = {}
        for key, value in unique_en_data.items():
            if key not in unique_zh_data:
                missing_translations[key] = value
        
        print(f"去重后英文翻译文件中的键数量: {len(unique_en_data)}")
        print(f"去重后中文翻译文件中的键数量: {len(unique_zh_data)}")
        print(f"英文文件中的重复键数量: {len(duplicate_en_keys)}")
        print(f"中文文件中的重复键数量: {len(duplicate_zh_keys)}")
        
        # 如果英文文件中有重复键，将去重后的英文翻译写回文件
        if len(duplicate_en_keys) > 0:
            with open(en_file_path, 'w', encoding='utf-8') as f:
                json.dump(unique_en_data, f, ensure_ascii=False, indent=2)
            print(f"已从英文文件中移除 {len(duplicate_en_keys)} 个重复的键: {duplicate_en_keys}")
        else:
            print("英文文件中没有重复的键")
        
        # 如果中文文件中有重复键，将去重后的中文翻译写回文件
        if len(duplicate_zh_keys) > 0:
            # 先保存去重后的中文数据
            with open(zh_file_path, 'w', encoding='utf-8') as f:
                json.dump(unique_zh_data, f, ensure_ascii=False, indent=2)
            print(f"已从中文文件中移除 {len(duplicate_zh_keys)} 个重复的键: {duplicate_zh_keys}")
        else:
            print("中文文件中没有重复的键")
        
        # 如果有缺失的翻译，添加到中文文件
        if missing_translations:
            # 将缺失的翻译添加到中文数据
            for key, value in missing_translations.items():
                unique_zh_data[key] = value
            
            # 将更新后的中文数据写回文件
            with open(zh_file_path, 'w', encoding='utf-8') as f:
                json.dump(unique_zh_data, f, ensure_ascii=False, indent=2)
            
            print(f"已添加 {len(missing_translations)} 个缺失的翻译到中文文件")
            print("缺失的键:")
            for key in missing_translations:
                print(f"  - {key}")
        else:
            print("没有发现缺失的翻译")
            
    except json.JSONDecodeError as e:
        print(f"JSON解析错误: {e}")
    except Exception as e:
        print(f"处理文件时出错: {e}")

if __name__ == "__main__":
    process_translation_files()