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
    """提取JS代码中的中文字符串，排除注释中的中文"""
    # 先移除单行注释
    content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
    # 移除多行注释
    content = re.sub(r'/\*[\s\S]*?\*/', '', content)
    
    # 查找字符串中的中文
    # 匹配双引号字符串
    double_quotes = re.finditer(r'"([^"\\]|\\[\s\S])*?"', content)
    # 匹配单引号字符串
    single_quotes = re.finditer(r"'([^'\\]|\\[\s\S])*?'", content)
    # 匹配反引号字符串(模板字符串)
    backticks = re.finditer(r"`([^`\\]|\\[\s\S])*?`", content)
    
    all_strings = []
    for match_iter in [double_quotes, single_quotes, backticks]:
        for match in match_iter:
            string_content = match.group(0)
            # 检查字符串是否包含中文
            if re.search(r'[\u4e00-\u9fff]', string_content):
                # 检查该字符串是否被t()函数包裹
                start_pos = match.start()
                # 向前查找t(，允许空格
                pre_content = content[:start_pos].strip()
                if not (pre_content.endswith('t(') or 
                        re.search(r't\s*\(\s*$', pre_content)):
                    all_strings.append((string_content, match.start()))
    
    # 检查JSX中的中文文本
    jsx_texts = re.finditer(r'>([^<>]*?[\u4e00-\u9fff][^<>]*?)<', content)
    for match in jsx_texts:
        text = match.group(1).strip()
        if text and re.search(r'[\u4e00-\u9fff]', text):
            # 检查是否被t()函数包裹
            if not ('{t(' in text or re.search(r'{.*?t\s*\([^)]*\).*?}', text)):
                all_strings.append((text, match.start()))
    
    return all_strings

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
        
        # 打印中文字符串示例 (最多显示3个)
        strings = results[path]
        for i, (string, _) in enumerate(strings[:3]):
            string_display = string.strip()
            if len(string_display) > 50:
                string_display = string_display[:47] + "..."
            prefix = "│   " * (len(dirs) + 1) + "├── "
            print(f"{prefix}发现: {string_display}")
        
        if len(strings) > 3:
            prefix = "│   " * (len(dirs) + 1) + "└── "
            print(f"{prefix}... 还有 {len(strings) - 3} 处未翻译的中文")

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
    excluded_dirs = [excluded_dir]
    
    print(f"开始扫描 {web_dir} 目录下的JS文件...")
    print(f"排除目录: {', '.join(excluded_dirs)}")
    
    results = scan_js_files(web_dir, excluded_dirs)
    print_tree(results, web_dir)
    
    print(f"\n总计: {len(results)} 个文件中发现未翻译的中文")

if __name__ == "__main__":
    main()
