def remove_duplicates(input_file, output_file):
    """
    从文件中删除重复行并将唯一行写入新文件
    
    参数:
    input_file (str): 输入文件路径
    output_file (str): 输出文件路径
    
    返回:
    tuple: (唯一行数量, 重复行数量)
    """
    # 读取所有行
    with open(input_file, 'r', encoding='utf-8') as f:
        lines = f.readlines()
    
    # 统计原始行数
    total_lines = len(lines)
    
    # 使用集合去重，同时保持原始顺序
    unique_lines = []
    seen = set()
    
    for line in lines:
        line = line.strip()
        if line not in seen:
            seen.add(line)
            unique_lines.append(line)
    
    # 计算重复行数量
    duplicates_count = total_lines - len(unique_lines)
    
    # 将唯一行写入新文件
    with open(output_file, 'w', encoding='utf-8') as f:
        for line in unique_lines:
            f.write(line + '\n')
    
    return len(unique_lines), duplicates_count

if __name__ == "__main__":
    input_file = "./questions.txt"
    output_file = "./questionnew.txt"
    
    unique_count, duplicate_count = remove_duplicates(input_file, output_file)
    
    print(f"原始文件行数: {unique_count + duplicate_count}")
    print(f"唯一行数: {unique_count}")
    print(f"重复行数: {duplicate_count}")
    
    if duplicate_count > 0:
        print(f"已删除 {duplicate_count} 行重复内容并保存到 {output_file}")
    else:
        print("文件中没有重复行")