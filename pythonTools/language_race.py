import json
import os
import glob

def process_translation_files():
    # 文件路径
    base_dir = r"D:\www\burncloud-aiapi-dev-local\web\src\i18n\locales"
    en_file_path = os.path.join(base_dir, "en.json")
    
    # 检查英文文件是否存在
    if not os.path.exists(en_file_path):
        print(f"错误: 英文翻译文件不存在: {en_file_path}")
        return
    
    try:
        # 读取英文翻译文件作为模板
        with open(en_file_path, 'r', encoding='utf-8') as f:
            en_data = json.load(f)
        
        print(f"英文翻译文件中的键数量: {len(en_data)}")
        
        # 获取目录中所有的JSON文件
        json_files = glob.glob(os.path.join(base_dir, "*.json"))
        
        # 处理每个翻译文件
        for file_path in json_files:
            # 跳过英文文件
            if file_path == en_file_path:
                continue
            
            file_name = os.path.basename(file_path)
            print(f"\n处理文件: {file_name}")
            
            try:
                # 读取目标翻译文件
                with open(file_path, 'r', encoding='utf-8') as f:
                    target_data = json.load(f)
                
                original_count = len(target_data)
                print(f"原始键数量: {original_count}")
                
                # 1. 去除重复键
                unique_target_data = {}
                duplicate_keys = []
                for key, value in target_data.items():
                    if key not in unique_target_data:
                        unique_target_data[key] = value
                    else:
                        duplicate_keys.append(key)
                
                # 2. 创建新的有序字典，按照英文文件的键顺序排序
                ordered_target_data = {}
                missing_keys = []
                
                # 首先添加英文文件中的键
                for key in en_data.keys():
                    if key in unique_target_data:
                        ordered_target_data[key] = unique_target_data[key]
                    else:
                        # 如果目标文件中没有该键，添加空字符串作为值
                        ordered_target_data[key] = ""
                        missing_keys.append(key)
                
                # 检查目标文件中有但英文文件中没有的键
                extra_keys = []
                for key in unique_target_data:
                    if key not in en_data:
                        extra_keys.append(key)
                        # 可选：保留这些额外的键
                        ordered_target_data[key] = unique_target_data[key]
                
                # 将处理后的数据写回文件
                with open(file_path, 'w', encoding='utf-8') as f:
                    json.dump(ordered_target_data, f, ensure_ascii=False, indent=2)
                
                # 输出处理结果
                print(f"处理后键数量: {len(ordered_target_data)}")
                
                if duplicate_keys:
                    print(f"移除了 {len(duplicate_keys)} 个重复的键")
                
                if missing_keys:
                    print(f"添加了 {len(missing_keys)} 个缺失的键（值设为空字符串）")
                
                if extra_keys:
                    print(f"保留了 {len(extra_keys)} 个在英文文件中不存在的键")
                
            except json.JSONDecodeError as e:
                print(f"解析文件 {file_name} 时出错: {e}")
            except Exception as e:
                print(f"处理文件 {file_name} 时出错: {e}")
        
        print("\n所有翻译文件处理完成！")
        
    except json.JSONDecodeError as e:
        print(f"解析英文文件时出错: {e}")
    except Exception as e:
        print(f"处理过程中出错: {e}")

if __name__ == "__main__":
    process_translation_files()