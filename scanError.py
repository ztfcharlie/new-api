import os
import re
from collections import defaultdict

def check_lang_t_calls(content):
    """检查文件中不合规的 lang.T 函数调用"""
    # 匹配 lang.T 函数调用的正则表达式
    # 匹配形如 lang.T("xxx") 的调用，但不匹配 lang.T(nil, "xxx") 或 lang.T(c, "xxx")
    pattern = r'lang\.T\s*\(\s*(?!nil|c\s*,)([^,\)]+)\s*\)'
    
    # 移除注释以避免误报
    # 移除单行注释
    content = re.sub(r'//.*$', '', content, flags=re.MULTILINE)
    # 移除多行注释
    content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)
    
    # 查找所有匹配
    matches = re.finditer(pattern, content)
    return any(matches)

def scan_go_files(root_dir):
    """扫描Go文件并返回包含不合规 lang.T 调用的文件路径"""
    invalid_files = defaultdict(list)
    
    for dirpath, dirnames, filenames in os.walk(root_dir):
        # 排除 web 目录
        if '/web' in dirpath or '\\web' in dirpath:
            continue
            
        # 只处理 .go 文件
        for filename in filenames:
            if filename.endswith('.go'):
                file_path = os.path.join(dirpath, filename)
                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        content = f.read()
                        if check_lang_t_calls(content):
                            # 使用相对路径
                            rel_path = os.path.relpath(file_path, root_dir)
                            dir_name = os.path.dirname(rel_path)
                            if dir_name == '':
                                dir_name = '.'
                            invalid_files[dir_name].append(filename)
                except UnicodeDecodeError:
                    print(f"Warning: Unable to read {file_path}")
                except Exception as e:
                    print(f"Error processing {file_path}: {str(e)}")
    
    return invalid_files

def print_tree(files_dict):
    """以树状结构打印文件"""
    def sort_key(path):
        """排序键，将点号路径排在前面"""
        return (0 if path == '.' else 1, path)
    
    print("\n发现包含不合规 lang.T 调用的文件树结构：")
    print("root")
    
    # 按目录排序
    for directory in sorted(files_dict.keys(), key=sort_key):
        # 处理目录显示
        if directory == '.':
            indent = "├── "
            file_indent = "│   ├── "
        else:
            parts = directory.split(os.sep)
            indent = "├── " + "│   " * (len(parts) - 1) + "├── "
            file_indent = "│   " * (len(parts)) + "├── "
            
            # 打印目录路径
            for i, part in enumerate(parts):
                print("│   " * i + "├── " + part)
        
        # 打印文件
        files = sorted(files_dict[directory])
        for i, file in enumerate(files):
            if i == len(files) - 1 and directory == list(files_dict.keys())[-1]:
                print(file_indent.replace("├", "└").replace("│", " ") + file)
            else:
                print(file_indent + file)

def main():
    """主函数"""
    # 获取当前目录作为根目录
    root_dir = os.getcwd()
    
    print(f"开始扫描目录: {root_dir}")
    print("正在查找包含不合规 lang.T 调用的Go文件...")
    
    # 扫描文件
    invalid_files = scan_go_files(root_dir)
    
    if invalid_files:
        print_tree(invalid_files)
        
        # 打印统计信息
        total_files = sum(len(files) for files in invalid_files.values())
        total_dirs = len(invalid_files)
        print(f"\n统计信息:")
        print(f"发现 {total_files} 个包含不合规 lang.T 调用的文件")
        print(f"分布在 {total_dirs} 个目录中")
    else:
        print("\n未发现包含不合规 lang.T 调用的Go文件")

if __name__ == "__main__":
    main()
