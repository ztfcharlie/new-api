#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import re
import sys
from collections import defaultdict

def is_excluded(path, excluded_dirs):
    """检查路径是否在排除目录中"""
    for excluded in excluded_dirs:
        if excluded in path:
            return True
    return False

def extract_chinese_strings(content):
    """提取JS代码中的中文字符串，排除注释中的中文和已被t()函数包裹的中文"""
    # 先移除单行注释
    content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
    # 移除多行注释
    content = re.sub(r'/\*[\s\S]*?\*/', '', content)
    
    all_strings = []
    
    # 1. 检测字符串中的中文
    def find_chinese_in_strings():
        results = []
        
        # 查找所有字符串（单引号、双引号和反引号）
        string_patterns = [
            r'"([^"\\]|\\[\s\S])*?"',  # 双引号字符串
            r"'([^'\\]|\\[\s\S])*?'",  # 单引号字符串
            r"`([^`\\]|\\[\s\S])*?`"   # 反引号字符串(模板字符串)
        ]
        
        for pattern in string_patterns:
            for match in re.finditer(pattern, content):
                string_content = match.group(0)
                
                # 检查字符串是否包含中文
                if not re.search(r'[\u4e00-\u9fff]', string_content):
                    continue
                
                # 获取字符串的位置
                start_pos = match.start()
                end_pos = match.end()
                
                # 检查该字符串是否被t()函数包裹
                # 向前查找最近的非空白字符序列
                pre_content = content[:start_pos].rstrip()
                
                # 检查是否在t()函数内
                in_t_function = False
                
                # 检查是否是t('中文')或t("中文")或t(`中文`)的形式
                if pre_content.endswith('t('):
                    in_t_function = True
                elif re.search(r't\s*\($', pre_content):
                    in_t_function = True
                
                # 如果不在t()函数内，添加到结果中
                if not in_t_function:
                    # 检查是否在其他t()函数调用中
                    # 例如：someVar ? t('中文1') : t('中文2')
                    # 或者：`${t('中文')}`
                    
                    # 提取实际的字符串内容（去掉引号）
                    inner_content = string_content[1:-1]
                    
                    # 检查是否在更大的上下文中被t()包裹
                    surrounding_code = content[max(0, start_pos-30):min(len(content), end_pos+30)]
                    if re.search(r't\s*\(\s*[\'"`]' + re.escape(inner_content) + r'[\'"`]\s*\)', surrounding_code):
                        continue
                    
                    results.append((string_content, start_pos))
        
        return results
    
    # 2. 检测JSX中的中文文本
    def find_chinese_in_jsx():
        results = []
        # 查找JSX标签之间的中文文本
        for match in re.finditer(r'>([^<>]*?[\u4e00-\u9fff][^<>]*?)<', content):
            text = match.group(1).strip()
            if text and re.search(r'[\u4e00-\u9fff]', text):
                # 检查是否被t()函数包裹
                if not ('{t(' in text or re.search(r'{.*?t\s*\([^)]*\).*?}', text)):
                    results.append((text, match.start()))
        return results
    
    # 3. 检测对象属性中的中文
    def find_chinese_in_object_properties():
        results = []
        # 查找对象属性中的中文，如 { label: '中文', value: 'value' }
        for match in re.finditer(r'[{,]\s*[a-zA-Z0-9_]+\s*:\s*[\'"`](.*?[\u4e00-\u9fff].*?)[\'"`]', content):
            prop_value = match.group(1)
            if re.search(r'[\u4e00-\u9fff]', prop_value):
                # 获取完整的属性值字符串
                start_idx = match.start(1) - 1  # 减1是为了包含引号
                end_idx = match.end(1) + 1      # 加1是为了包含引号
                full_string = content[start_idx:end_idx]
                
                # 检查是否在t()函数内
                surrounding_code = content[max(0, start_idx-30):min(len(content), end_idx+30)]
                if not re.search(r't\s*\(\s*' + re.escape(full_string) + r'\s*\)', surrounding_code):
                    results.append((full_string, start_idx))
        return results
    
    # 合并所有结果
    all_strings.extend(find_chinese_in_strings())
    all_strings.extend(find_chinese_in_jsx())
    all_strings.extend(find_chinese_in_object_properties())
    
    # 过滤掉可能的误报
    filtered_strings = []
    for string_content, pos in all_strings:
        # 过滤掉只包含代码块而不是实际中文字符串的情况
        if string_content.strip().startswith('{') and string_content.strip().endswith('}'):
            # 检查是否真的包含中文字符，而不仅仅是代码块
            inner_content = string_content.strip()[1:-1]
            if not re.search(r'[\u4e00-\u9fff]', inner_content):
                continue
            
            # 如果是代码块，检查是否有明显的中文字符串
            chinese_strings_in_block = re.findall(r'[\'"`](.*?[\u4e00-\u9fff].*?)[\'"`]', inner_content)
            if not chinese_strings_in_block:
                continue
            
            # 将代码块中的中文字符串作为结果
            for chinese_str in chinese_strings_in_block:
                if re.search(r'[\u4e00-\u9fff]', chinese_str):
                    filtered_strings.append((f'"{chinese_str}"', pos))
        else:
            # 确保字符串确实包含中文
            if re.search(r'[\u4e00-\u9fff]', string_content):
                filtered_strings.append((string_content, pos))
    
    return filtered_strings

def scan_js_files(root_dir, excluded_dirs):
    """扫描JS文件并检测未翻译的中文"""
    results = defaultdict(list)
    file_count = 0
    scanned_files = 0
    
    print(f"开始扫描目录: {root_dir}")
    
    if not os.path.exists(root_dir):
        print(f"错误: 目录 {root_dir} 不存在!")
        return results
    
    for root, dirs, files in os.walk(root_dir):
        # 跳过排除的目录
        if is_excluded(root, excluded_dirs):
            print(f"跳过排除目录: {root}")
            continue
            
        for file in files:
            if file.endswith(('.js', '.jsx', '.ts', '.tsx')):
                file_path = os.path.join(root, file)
                file_count += 1
                
                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        content = f.read()
                    
                    scanned_files += 1
                    if file_count % 100 == 0:
                        print(f"已扫描 {file_count} 个文件...")
                    
                    chinese_strings = extract_chinese_strings(content)
                    if chinese_strings:
                        # 存储相对路径
                        rel_path = os.path.relpath(file_path, root_dir)
                        results[rel_path] = chinese_strings
                        print(f"找到未翻译中文: {rel_path} ({len(chinese_strings)}处)")
                except Exception as e:
                    print(f"处理文件 {file_path} 时出错: {e}", file=sys.stderr)
    
    print(f"扫描完成: 共扫描 {scanned_files} 个JS文件，找到 {len(results)} 个包含未翻译中文的文件")
    return results

def print_tree(results, root_dir):
    """以目录树的形式打印结果"""
    if not results:
        print("没有找到未翻译的中文字符串")
        return
    
    print(f"找到 {len(results)} 个包含未翻译中文的文件:")
    print(f"{root_dir}")
    
    # 按路径排序
    sorted_paths = sorted(results.keys())
    
    # 构建目录树
    current_dirs = []
    for path in sorted_paths:
        parts = path.split(os.sep)
        file_name = parts[-1]
        dirs = parts[:-1]
        
        # 打印目录结构
        for i, dir_name in enumerate(dirs):
            if i >= len(current_dirs):
                prefix = "│   " * i + "├── "
                print(f"{prefix}{dir_name}/")
                current_dirs.append(dir_name)
            elif current_dirs[i] != dir_name:
                prefix = "│   " * i + "├── "
                print(f"{prefix}{dir_name}/")
                current_dirs[i] = dir_name
                current_dirs = current_dirs[:i+1]
        
        # 打印文件名
        prefix = "│   " * len(dirs) + "├── "
        print(f"{prefix}{file_name}")
        
        # 打印中文字符串示例 (最多显示5个)
        strings = results[path]
        for i, (string, _) in enumerate(strings[:5]):
            # 清理字符串显示
            string_display = string.strip()
            # 如果是多行字符串，只显示第一行
            if '\n' in string_display:
                string_display = string_display.split('\n')[0] + '...'
            # 截断过长的字符串
            if len(string_display) > 50:
                string_display = string_display[:47] + "..."
            prefix = "│   " * (len(dirs) + 1) + "├── "
            print(f"{prefix}发现: {string_display}")
        
        if len(strings) > 5:
            prefix = "│   " * (len(dirs) + 1) + "└── "
            print(f"{prefix}... 还有 {len(strings) - 5} 处未翻译的中文")

def main():
    # 使用相对路径，适应Windows环境
    # 获取当前工作目录
    current_dir = os.getcwd()
    print(f"当前工作目录: {current_dir}")
    
    # 确定web目录的位置
    web_dir = input("请输入web目录的路径 (默认为当前目录下的web子目录): ").strip() or os.path.join(current_dir, "web")
    
    if not os.path.exists(web_dir):
        print(f"错误: 目录 {web_dir} 不存在!")
        return
    
    # 确定排除目录
    excluded_dir = os.path.join(web_dir, "dist")
    excluded_dirs = [excluded_dir, "node_modules", ".git", "build"]
    
    print(f"开始扫描 {web_dir} 目录下的JS文件...")
    print(f"排除目录: {', '.join(excluded_dirs)}")
    
    results = scan_js_files(web_dir, excluded_dirs)
    print_tree(results, web_dir)
    
    print(f"\n总计: {len(results)} 个文件中发现未翻译的中文")

if __name__ == "__main__":
    main()
