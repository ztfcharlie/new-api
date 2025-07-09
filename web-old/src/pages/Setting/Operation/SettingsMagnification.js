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

def extract_chinese_strings(content, file_path):
    """提取JS代码中的中文字符串，排除注释中的中文和已被t()函数包裹的中文"""
    # 先移除单行注释
    content_without_comments = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
    # 移除多行注释
    content_without_comments = re.sub(r'/\*[\s\S]*?\*/', '', content_without_comments)
    
    all_strings = []
    
    # 1. 检测字符串字面量中的中文（单引号、双引号和反引号字符串）
    def find_chinese_in_string_literals():
        results = []
        
        # 查找所有字符串字面量
        string_patterns = [
            (r'"([^"\\]|\\[\s\S])*?"', '"'),  # 双引号字符串
            (r"'([^'\\]|\\[\s\S])*?'", "'"),  # 单引号字符串
        ]
        
        for pattern, quote_type in string_patterns:
            for match in re.finditer(pattern, content_without_comments):
                string_with_quotes = match.group(0)
                
                # 检查字符串是否包含中文
                if not re.search(r'[\u4e00-\u9fff]', string_with_quotes):
                    continue
                
                # 提取字符串内容（去掉引号）
                string_content = string_with_quotes[1:-1]
                
                # 获取字符串的位置
                start_pos = match.start()
                
                # 检查该字符串是否被t()函数包裹
                # 1. 检查直接包裹: t('中文')
                pre_content = content_without_comments[:start_pos].rstrip()
                if pre_content.endswith('t(') or re.search(r't\s*\($', pre_content):
                    continue
                
                # 2. 检查更广泛的上下文: 如 t('中文') 或 {t('中文')} 或 ${t('中文')}
                surrounding_code = content_without_comments[max(0, start_pos-50):min(len(content_without_comments), start_pos+len(string_with_quotes)+50)]
                t_pattern = r't\s*\(\s*' + re.escape(string_with_quotes) + r'\s*\)'
                if re.search(t_pattern, surrounding_code):
                    continue
                
                # 3. 检查是否在JSX属性中被包裹，如 placeholder={t('中文')}
                jsx_attr_pattern = r'=\s*\{\s*t\s*\(\s*' + re.escape(string_with_quotes) + r'\s*\)\s*\}'
                if re.search(jsx_attr_pattern, surrounding_code):
                    continue
                
                # 如果通过了所有检查，添加到结果中
                results.append((string_with_quotes, start_pos))
        
        return results
    
    # 2. 检测模板字符串中的中文
    def find_chinese_in_template_strings():
        results = []
        
        # 查找所有模板字符串
        for match in re.finditer(r'`([^`\\]|\\[\s\S])*?`', content_without_comments):
            template_string = match.group(0)
            
            # 检查模板字符串是否包含中文
            if not re.search(r'[\u4e00-\u9fff]', template_string):
                continue
            
            # 获取模板字符串的位置
            start_pos = match.start()
            
            # 检查整个模板字符串是否被t()函数包裹
            pre_content = content_without_comments[:start_pos].rstrip()
            if pre_content.endswith('t(') or re.search(r't\s*\($', pre_content):
                continue
            
            # 检查模板字符串中的中文部分是否都在t()函数中
            # 分离模板字符串中的表达式和文本部分
            template_content = template_string[1:-1]  # 去掉反引号
            parts = re.split(r'(\${[^}]*?})', template_content)
            
            has_untranslated_chinese = False
            chinese_parts = []
            
            for part in parts:
                # 如果是表达式部分 ${...}
                if part.startswith('${') and part.endswith('}'):
                    # 检查表达式中是否有t()函数调用
                    expr = part[2:-1]  # 去掉 ${ 和 }
                    if 't(' in expr and re.search(r'[\u4e00-\u9fff]', expr):
                        # 表达式中有t()函数调用且包含中文，认为已翻译
                        continue
                # 如果是文本部分
                else:
                    # 检查文本部分是否包含中文
                    if re.search(r'[\u4e00-\u9fff]', part):
                        has_untranslated_chinese = True
                        chinese_parts.append(part)
            
            # 如果模板字符串中有未翻译的中文，添加到结果中
            if has_untranslated_chinese:
                # 只添加包含未翻译中文的部分，而不是整个模板字符串
                for chinese_part in chinese_parts:
                    if chinese_part.strip():  # 确保不是空白字符
                        results.append((f'"{chinese_part}"', start_pos))
            
        return results
    
    # 3. 检测JSX中的中文文本
    def find_chinese_in_jsx():
        results = []
        # 查找JSX标签之间的中文文本
        for match in re.finditer(r'>([^<>]*?[\u4e00-\u9fff][^<>]*?)<', content_without_comments):
            text = match.group(1).strip()
            if text and re.search(r'[\u4e00-\u9fff]', text):
                # 检查是否被t()函数包裹
                if not ('{t(' in text or re.search(r'{.*?t\s*\([^)]*\).*?}', text)):
                    results.append((f'"{text}"', match.start()))
        return results
    
    # 合并所有结果
    all_strings.extend(find_chinese_in_string_literals())
    all_strings.extend(find_chinese_in_template_strings())
    all_strings.extend(find_chinese_in_jsx())
    
    # 过滤和清理结果
    filtered_strings = []
    for string_content, pos in all_strings:
        # 确保字符串确实包含中文
        if re.search(r'[\u4e00-\u9fff]', string_content):
            # 清理字符串，去掉前后的引号和多余的空白
            cleaned_content = string_content.strip()
            if cleaned_content.startswith(('"', "'", '`')) and cleaned_content.endswith(('"', "'", '`')):
                cleaned_content = cleaned_content[1:-1]
            
            # 过滤掉一些明显不是需要翻译的内容
            if cleaned_content and not cleaned_content.isspace():
                # 过滤掉只包含代码的字符串
                if not (cleaned_content.startswith('{') and cleaned_content.endswith('}')):
                    filtered_strings.append((cleaned_content, pos))
    
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
                    
                    chinese_strings = extract_chinese_strings(content, file_path)
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
            if len(string) > 50:
                string_display = string[:47] + "..."
            else:
                string_display = string
            
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
    excluded_dirs = [excluded_dir, "node_modules", ".git", "build", "locales"]
    
    print(f"开始扫描 {web_dir} 目录下的JS文件...")
    print(f"排除目录: {', '.join(excluded_dirs)}")
    
    results = scan_js_files(web_dir, excluded_dirs)
    print_tree(results, web_dir)
    
    print(f"\n总计: {len(results)} 个文件中发现未翻译的中文")

if __name__ == "__main__":
    main()
