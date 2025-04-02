import os
import re
from collections import defaultdict

def contains_chinese(content):
    """检查内容是否包含中文字符"""
    chinese_pattern = re.compile(r'[\u4e00-\u9fff]')
    # 排除注释中的中文
    # 移除单行注释
    content = re.sub(r'//.*$', '', content, flags=re.MULTILINE)
    # 移除多行注释
    content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)
    return bool(chinese_pattern.search(content))

def scan_go_files(root_dir):
    """扫描Go文件并返回包含中文的文件路径"""
    chinese_files = defaultdict(list)
    
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
                        if contains_chinese(content):
                            # 使用相对路径
                            rel_path = os.path.relpath(file_path, root_dir)
                            dir_name = os.path.dirname(rel_path)
                            if dir_name == '':
                                dir_name = '.'
                            chinese_files[dir_name].append(filename)
                except UnicodeDecodeError:
                    print(f"Warning: Unable to read {file_path}")
                except Exception as e:
                    print(f"Error processing {file_path}: {str(e)}")
    
    return chinese_files

def print_tree(files_dict):
    """以树状结构打印文件"""
    def sort_key(path):
        """排序键，将点号路径排在前面"""
        return (0 if path == '.' else 1, path)
    
    print("\n发现包含中文的文件树结构：")
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
    print("正在查找包含中文的Go文件...")
    
    # 扫描文件
    chinese_files = scan_go_files(root_dir)
    
    if chinese_files:
        print_tree(chinese_files)
        
        # 打印统计信息
        total_files = sum(len(files) for files in chinese_files.values())
        total_dirs = len(chinese_files)
        print(f"\n统计信息:")
        print(f"发现 {total_files} 个包含中文的文件")
        print(f"分布在 {total_dirs} 个目录中")
    else:
        print("\n未发现包含中文的Go文件")

if __name__ == "__main__":
    main()
