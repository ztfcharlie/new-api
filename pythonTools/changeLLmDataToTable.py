import json

def format_list_to_lines(items):
    """将列表转换为HTML换行格式"""
    return "<br>".join(f"{item}" for item in items)

def replace_not_disclosed(text):
    """替换'Not officially disclosed'为'-'"""
    if isinstance(text, str):
        return text.replace("Not officially disclosed", "-")
    return text

def generate_html_table(data):
    # HTML头部
    html = """
<html>
<head>
    <style>
        table { 
            border-collapse: collapse; 
            width: 100%; 
            margin: 20px 0;
        }
        th, td { 
            border: 1px solid black; 
            padding: 8px; 
            text-align: left; 
            vertical-align: top;
        }
        th { 
            background-color: #f2f2f2; 
            font-weight: bold;
        }
        td ul {
            margin: 0;
            padding-left: 20px;
        }
        .feature-cell {
            max-width: 300px;
        }
    </style>
</head>
<body>
    <table>
        <thead>
            <tr>
                <th>Provider</th>
                <th>Model Name</th>
                <th>Release Date</th>
                <th>Parameter Scale</th>
                <th>Features</th>
                <th>Application Scenarios</th>
            </tr>
        </thead>
        <tbody>
"""
    
    # 处理数据
    for item in data:  # 移除了[0]的索引
        provider_name = item["provider"]
        models = item["models"]
        
        # 处理每个模型
        for i, model in enumerate(models):
            html += "<tr>"
            
            # 只在第一行添加provider名称
            if i == 0:
                html += f'<td rowspan="{len(models)}">{provider_name}</td>'
            
            # 添加模型信息，features和application_scenarios使用换行格式
            # 使用replace_not_disclosed处理parameter_scale
            html += f"""
                <td>{model["model_name"]}</td>
                <td>{model["release_date"]}</td>
                <td>{replace_not_disclosed(model["parameter_scale"])}</td>
                <td class="feature-cell">{format_list_to_lines(model["features"])}</td>
                <td class="feature-cell">{format_list_to_lines(model["application_scenarios"])}</td>
            """
            html += "</tr>\n"
    
    # HTML尾部
    html += """</tbody>
    </table>
</body>
</html>
"""
    
    return html

def save_to_file(html_content, filename="llm_models.html"):
    with open(filename, "w", encoding="utf-8") as f:
        f.write(html_content)

def main():
    # 指定JSON文件路径
    json_file = "llm_data.json"
    
    try:
        # 尝试不同的编码方式读取文件
        encodings = ['utf-8', 'gbk', 'gb2312', 'iso-8859-1']
        content = None
        
        for encoding in encodings:
            try:
                with open(json_file, "r", encoding=encoding) as f:
                    content = f.read()
                print(f"成功使用 {encoding} 编码读取文件")
                break
            except UnicodeDecodeError:
                continue
        
        if content is None:
            raise Exception("无法使用任何已知编码读取文件")
            
        # 解析JSON
        data = json.loads(content)
        
        # 生成HTML表格
        html_content = generate_html_table(data)
        
        # 保存到文件
        save_to_file(html_content)
        
        print(f"HTML表格已生成并保存到 llm_models.html")
        
    except FileNotFoundError:
        print(f"错误: 找不到文件 {json_file}")
    except json.JSONDecodeError as e:
        print(f"错误: JSON格式不正确 - {str(e)}")
    except Exception as e:
        print(f"发生错误: {str(e)}")

if __name__ == "__main__":
    main()