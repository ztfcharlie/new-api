import os
from bs4 import BeautifulSoup

def extract_fifth_column_from_html_table(input_file, output_file):
    """
    从HTML表格中提取第五列的文本内容，并去除HTML标签
    
    参数:
    input_file -- 输入HTML文件路径
    output_file -- 输出文本文件路径
    """
    try:
        # 读取输入HTML文件
        with open(input_file, 'r', encoding='utf-8') as f:
            html_content = f.read()
        
        # 使用BeautifulSoup解析HTML
        soup = BeautifulSoup(html_content, 'html.parser')
        
        # 查找所有表格行
        rows = soup.find_all('tr')
        
        # 提取第五列的文本
        extracted_texts = []
        for row in rows:
            # 获取行中的所有单元格
            cells = row.find_all(['td', 'th'])
            
            # 检查是否有足够的单元格
            if len(cells) >= 6:  # 需要至少5列
                fifth_cell = cells[5]  # 第5列（索引为4）
                
                # 提取单元格文本并去除HTML标签
                cell_text = fifth_cell.get_text(separator=' ', strip=True)
                
                if cell_text:  # 只添加非空内容
                    extracted_texts.append(cell_text)
        
        # 写入输出文件
        with open(output_file, 'w', encoding='utf-8') as f:
            for text in extracted_texts:
                f.write(f"{text}\n")
        
        print(f"已将提取的文本保存到 {output_file}")
        print(f"共提取了 {len(extracted_texts)} 行数据")
        return True
    
    except Exception as e:
        print(f"处理HTML表格时出错: {str(e)}")
        return False

if __name__ == "__main__":
    # 固定的输入文件路径
    input_file = r"D:\www\burncloud-aiapi-dev-local\pythonTools\responses\table.txt"
    
    # 输出文件路径（在输入文件同目录下）
    output_dir = os.path.dirname(input_file)
    output_file = os.path.join(output_dir, "extracted_column.txt")
    
    print(f"从HTML表格文件 {input_file} 中提取第5列...")
    
    success = extract_fifth_column_from_html_table(input_file, output_file)
    
    if success:
        print("处理完成")
    else:
        print("处理失败")