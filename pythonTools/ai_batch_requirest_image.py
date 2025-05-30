import aiohttp
import asyncio
import aiofiles
import json
import time
import os
import random
from tqdm import tqdm

# 配置参数
API_URL = "https://ai.burncloud.com/v1/images/generations"
API_KEY = "sk-9DufUZISqD4FrlVzvO8nkioqEwxqjPkJ2OwSAibs7d9OKJLe" #burncloud test token
MODEL = "gpt-image-1"
CONCURRENCY = 100  # 1秒并发数量，100张图
REQUEST_TIMEOUT = 180  # 单个请求的超时时间（秒）
QUESTION_COUNT = 6000  # 固定的请求总数量

# 确保输出目录存在
OUTPUT_DIR = "responses"
os.makedirs(OUTPUT_DIR, exist_ok=True)

# 100个不同的绘图提示
PROMPTS = [
    "一个性感的动作，职业装，办公室，职业女性，带眼镜，苗条，白人",
    "未来科技城市夜景，霓虹灯，高楼大厦，飞行汽车，全息投影广告",
    "宁静的日本传统庭院，樱花树，小溪，石灯笼，枯山水",
    "热带海滩日落，椰子树，金色沙滩，平静的海水，彩色天空",
    "古代中国皇宫内景，红色柱子，金色装饰，宫女，太监，皇帝",
    "神秘的森林场景，雾气，巨大的古树，蘑菇，小精灵，魔法光芒",
    "繁忙的纽约时代广场，人群，黄色出租车，广告牌，摩天大楼",
    "中世纪欧洲城堡，石墙，塔楼，护城河，旗帜，骑士",
    "宇宙太空站内部，宇航员，控制面板，地球视图，未来科技",
    "威尼斯水城风光，贡多拉船，运河，桥梁，意大利建筑",
    "非洲大草原日落，长颈鹿，大象，金合欢树，橙色天空",
    "北欧冬季小屋，雪景，极光，松树，温暖的窗户灯光",
    "水下珊瑚礁，热带鱼，海龟，阳光透过水面的光线",
    "蒸汽朋克风格的机械工厂，齿轮，管道，蒸汽，工人",
    "巴黎街头咖啡馆，户外座位，埃菲尔铁塔远景，鲜花",
    "中国水墨画风格的山水，雾气缭绕的山峰，小船，瀑布",
    "印度传统婚礼场景，新娘，新郎，彩色纱丽，花环，庆典",
    "荒凉的火星表面，宇航员，探测车，红色沙地，远处的基地",
    "维多利亚时代的英国客厅，古典家具，壁炉，茶具，贵族",
    "日本动漫风格的校园场景，樱花树，学生，制服，教学楼",
    "墨西哥亡灵节庆典，糖骷髅，万寿菊，彩色装饰，蜡烛",
    "现代极简主义公寓内景，大窗户，白色家具，绿植，城市景观",
    "古埃及金字塔内部，法老墓室，象形文字，黄金宝藏，火把",
    "旧西部荒野小镇，木制建筑，沙龙，牛仔，马匹，沙漠",
    "雨中的东京街道，霓虹灯反射，撑伞的人，出租车，水坑",
    "北极冰川风景，极地熊，浮冰，蓝色冰山，极光天空",
    "文艺复兴时期的意大利工作室，画家，画布，工具，自然光",
    "赛博朋克风格的夜间街道，全息广告，机械改造人，飞行器",
    "印度尼西亚巴厘岛梯田，绿色稻田，农民，热带树木，山景",
    "古希腊神庙，大理石柱，雕像，蓝天，地中海风光",
    "巴西狂欢节游行，舞者，彩色羽毛服装，鼓手，观众",
    "维京村庄，木制长屋，船只，战士，北欧风景",
    "澳大利亚内陆荒漠，红色沙土，尤加利树，袋鼠，日落",
    "伦敦雾天，大本钟，红色电话亭，黑色出租车，维多利亚建筑",
    "中国传统茶馆，木制家具，茶具，书法，竹子装饰",
    "摩洛哥马拉喀什市场，五彩缤纷的地毯，香料，灯笼，商贩",
    "俄罗斯冬宫内部，金色装饰，水晶吊灯，宫廷服饰，艺术品",
    "美国1950年代复古餐厅，霓虹灯，点唱机，红色卡座，服务员",
    "秘鲁马丘比丘遗址，石墙，梯田，安第斯山脉，雾气",
    "韩国传统韩屋村，瓦屋顶，木制门窗，庭院，传统服饰",
    "荷兰风车与郁金香田，彩色花田，蓝天，传统风车，运河",
    "泰国水上市场，小船，水果摊，传统服饰，热带环境",
    "爱尔兰绿色草原，石墙，羊群，古老城堡，多云天空",
    "新西兰山地风景，雪山，绿色草地，湖泊，指环王风格",
    "阿拉伯沙漠绿洲，棕榈树，帐篷，骆驼，沙丘，日落",
    "越南下龙湾，石灰岩岛屿，传统木船，绿色海水，多云天空",
    "芬兰桑拿小屋，木制内饰，热蒸汽，雪景窗外，松树",
    "西藏寺庙，喇嘛，经幡，雪山背景，金色屋顶",
    "加勒比海盗船，海盗，宝藏，大海，热带岛屿远景",
    "瑞士阿尔卑斯山小镇，木屋，雪山，绿色草地，牛群",
    "法国普罗旺斯薰衣草田，紫色花田，石头农舍，蓝天",
    "墨西哥玛雅金字塔，丛林环绕，石阶，古代雕刻，祭司",
    "智利复活节岛，摩艾石像，草地，海洋背景，多云天空",
    "苏格兰高地城堡，石墙，湖泊，雾气，绿色山丘",
    "南极科考站，企鹅，冰雪，科学家，极地设备",
    "新加坡未来城市景观，滨海湾花园，摩天大楼，灯光秀",
    "克罗地亚杜布罗夫尼克古城，红色屋顶，城墙，亚得里亚海",
    "葡萄牙里斯本彩色街道，电车，石板路，传统建筑，坡道",
    "挪威峡湾，陡峭山壁，蓝色水面，小村庄，游船",
    "希腊圣托里尼岛，白色建筑，蓝色圆顶，爱琴海景色，日落",
    "土耳其卡帕多西亚热气球，奇特岩石地貌，日出，彩色气球",
    "阿根廷探戈舞者，热情舞姿，红色裙子，黑色西装，老式舞厅",
    "肯尼亚马赛马拉大迁徙，角马群，狮子，大草原，金合欢树",
    "丹麦哥本哈根彩色港口，木船，五彩建筑，运河，咖啡馆",
    "以色列耶路撒冷老城，宗教建筑，石墙，市场，朝圣者",
    "哥伦比亚麦德林彩色社区，壁画，阶梯，山坡建筑，人们",
    "瑞典斯德哥尔摩老城，彩色建筑，鹅卵石街道，水道，船只",
    "奥地利维也纳音乐厅，古典装饰，管弦乐队，观众，指挥家",
    "柬埔寨吴哥窟，古庙，石雕，丛林环绕，日出光线",
    "古巴哈瓦那老城，复古汽车，彩色建筑，音乐家，海滨大道",
    "尼泊尔喜马拉雅山脉，雪山，祈祷旗，徒步者，蓝天",
    "匈牙利布达佩斯温泉浴场，拱门，彩色瓷砖，蒸汽，游泳者",
    "冰岛黑沙滩，玄武岩柱，海浪，极光，远处冰川",
    "波兰华沙老城，重建的彩色建筑，广场，历史气息，人群",
    "荷兰阿姆斯特丹运河，自行车，老房子，桥梁，船只",
    "马来西亚婆罗洲雨林，猩猩，热带植物，雾气，小溪",
    "捷克布拉格城堡，哥特式建筑，查理大桥，伏尔塔瓦河，夕阳",
    "埃塞俄比亚拉利贝拉岩石教堂，十字形建筑，朝圣者，古老仪式",
    "黎巴嫩贝鲁特老城，地中海风格建筑，市场，海岸线，日落",
    "格鲁吉亚高加索山村庄，石头房屋，葡萄园，雪山背景，传统服饰",
    "罗马尼亚特兰西瓦尼亚德古拉城堡，哥特式建筑，森林，雾气，满月",
    "保加利亚索菲亚亚历山大·涅夫斯基大教堂，金色圆顶，东正教风格",
    "斯里兰卡锡吉里耶狮子岩，古老阶梯，壁画，热带环境，云层",
    "乌兹别克斯坦撒马尔罕，蓝色圆顶，伊斯兰建筑，马赛克装饰，广场",
    "亚美尼亚修道院，石制建筑，十字架石碑，山地景观，古老教堂",
    "塞尔维亚贝尔格莱德老城，东西方融合建筑，多瑙河，咖啡馆，人群",
    "拉脱维亚里加老城，哥特式尖塔，彩色建筑，波罗的海风光，广场",
    "哈萨克斯坦草原，游牧民族，蒙古包，马匹，辽阔天空",
    "阿塞拜疆巴库老城，石墙，窄街，中东与欧洲混合风格，里海远景",
    "乌克兰基辅圣索菲亚大教堂，金色圆顶，东正教壁画，广场，游客",
    "缅甸蒲甘佛塔群，红砖塔，佛教僧侣，热气球，日出",
    "老挝琅勃拉邦清晨布施，僧侣队伍，传统建筑，晨雾，当地居民"
]

async def make_request(session, request_id):
    """发送单个请求到 API"""
    headers = {
        "Content-Type": "application/json",
        "Authorization": API_KEY
    }
    
    # 从提示列表中选择一个提示，如果请求ID超过了提示数量，则循环使用
    prompt_index = (request_id - 1) % len(PROMPTS)
    selected_prompt = PROMPTS[prompt_index]
    
    # 图像生成的请求体
    payload = {
        "model": MODEL,
        "prompt": selected_prompt,
        "n": 1,
        "size": "1024x1024",
        "quality": "medium",
        "responseFormat": "b64_json",
        "user": f"stress-test-{request_id}"
    }
    
    start_time = time.time()
    
    try:
        # 添加超时参数
        async with session.post(API_URL, headers=headers, json=payload, timeout=REQUEST_TIMEOUT) as response:
            response_text = await response.text()
            elapsed_time = time.time() - start_time
            
            # 尝试解析JSON响应
            try:
                response_json = json.loads(response_text)
                
                # 检查是否包含"Failed to request upstream address"错误
                upstream_error = False
                if "error" in response_json:
                    error_msg = str(response_json.get("error", ""))
                    if "Failed to request upstream address" in error_msg:
                        upstream_error = True
                
                # 如果成功且没有上游错误，移除base64数据以节省空间
                if response.status == 200 and "data" in response_json and not upstream_error:
                    for item in response_json["data"]:
                        if "b64_json" in item:
                            # 只保留base64字符串的长度信息而不是内容本身
                            b64_length = len(item["b64_json"])
                            item["b64_json"] = f"[BASE64_DATA_REMOVED - {b64_length} bytes]"
            except json.JSONDecodeError:
                response_json = {"error": "Failed to decode JSON", "raw_response": response_text}
                upstream_error = "Failed to request upstream address" in response_text
            
            # 输出请求信息到终端
            print(f"\nRequest {request_id}")
            print(f"Prompt: {selected_prompt[:50]}..." if len(selected_prompt) > 50 else f"Prompt: {selected_prompt}")
            print(f"Status code: {response.status}")
            print(f"Response time: {elapsed_time:.2f}s")
            
            # 判断请求是否成功 - 状态码为200且没有上游错误
            success = response.status == 200 and not upstream_error
            if success:
                print(f"Image generation successful")
            else:
                failure_reason = "upstream error" if upstream_error else f"status code {response.status}"
                print(f"Image generation failed ({failure_reason}): {response_text[:200]}..." if len(response_text) > 200 else response_text)
            
            return {
                "request_id": request_id,
                "prompt": selected_prompt,
                "status_code": response.status,
                "elapsed_time": elapsed_time,
                "success": success,
                "upstream_error": upstream_error,
                "response_json": response_json,
                # 只有失败时才保存完整的响应文本
                "raw_response": response_text if not success else "[RESPONSE_TEXT_REMOVED]"
            }
    except asyncio.TimeoutError:
        elapsed_time = time.time() - start_time
        error_message = f"Request timed out after {REQUEST_TIMEOUT} seconds"
        print(f"\nRequest {request_id}: {error_message}")
        return {
            "request_id": request_id,
            "prompt": selected_prompt,
            "status_code": None,
            "elapsed_time": elapsed_time,
            "success": False,
            "error": error_message,
            "error_type": "TimeoutError"
        }
    except Exception as e:
        elapsed_time = time.time() - start_time
        error_details = str(e)
        print(f"\nError processing request {request_id}")
        print(f"Error: {error_details}")
        return {
            "request_id": request_id,
            "prompt": selected_prompt,
            "status_code": None,
            "elapsed_time": elapsed_time,
            "success": False,
            "error": error_details,
            "error_type": type(e).__name__
        }

async def process_requests():
    """处理所有请求，确保每个请求都完成"""
    start_time = time.time()
    
    # 创建信号量来控制并发
    semaphore = asyncio.Semaphore(CONCURRENCY)
    
    async def bounded_request(request_id):
        async with semaphore:
            # 为每个请求创建一个新的会话
            async with aiohttp.ClientSession() as session:
                return await make_request(session, request_id)
    
    # 创建所有任务
    tasks = [bounded_request(i+1) for i in range(QUESTION_COUNT)]
    
    # 使用tqdm显示进度，等待所有任务完成
    results = []
    for future in tqdm(asyncio.as_completed(tasks), total=len(tasks), desc="Processing requests"):
        result = await future
        results.append(result)
    
    return results

async def analyze_failures(failed_requests):
    """分析失败的请求并生成详细报告"""
    if not failed_requests:
        return "No failed requests to analyze."
    
    # 按错误类型分组
    error_types = {}
    status_codes = {}
    upstream_errors = []
    
    for req in failed_requests:
        # 分析上游错误
        if req.get("upstream_error", False):
            upstream_errors.append(req)
        
        # 分析错误类型
        if "error_type" in req:
            error_type = req["error_type"]
            if error_type in error_types:
                error_types[error_type].append(req)
            else:
                error_types[error_type] = [req]
        
        # 分析状态码
        if "status_code" in req and req["status_code"] is not None:
            status_code = str(req["status_code"])
            if status_code in status_codes:
                status_codes[status_code].append(req)
            else:
                status_codes[status_code] = [req]
    
    # 生成报告
    report = []
    report.append(f"=== Failure Analysis Report ===")
    report.append(f"Total failed requests: {len(failed_requests)}")
    
    # 上游错误统计
    if upstream_errors:
        report.append(f"\n=== Upstream Errors ===")
        report.append(f"- Failed to request upstream address: {len(upstream_errors)} occurrences")
        report.append(f"  Note: These are requests with status code 200 but failed at the upstream service")
    
    # 错误类型统计
    if error_types:
        report.append("\n=== Error Types ===")
        for error_type, requests in error_types.items():
            report.append(f"- {error_type}: {len(requests)} occurrences")
            # 添加第一个错误的详细信息作为示例
            if requests:
                report.append(f"  Example error: {requests[0].get('error', 'No error message')}")
    
    # 状态码统计
    if status_codes:
        report.append("\n=== Status Codes ===")
        for status_code, requests in status_codes.items():
            report.append(f"- Status {status_code}: {len(requests)} occurrences")
    
    # 详细错误样本
    report.append("\n=== Detailed Failure Samples ===")
    
    # 首先展示上游错误的样本
    if upstream_errors:
        report.append("\n--- Upstream Error Samples ---")
        for i, req in enumerate(upstream_errors[:5]):  # 最多显示5个上游错误样本
            report.append(f"\nUpstream Error Sample #{i+1} (Request ID: {req['request_id']})")
            report.append(f"Prompt: {req['prompt']}")
            report.append(f"Status Code: {req.get('status_code', 'N/A')} (Note: Status code was 200 but request failed)")
            
            if "raw_response" in req:
                raw_resp = req['raw_response']
                # 如果响应太长，只显示前200个字符
                if len(raw_resp) > 200:
                    report.append(f"Raw Response (truncated): {raw_resp[:200]}...")
                else:
                    report.append(f"Raw Response: {raw_resp}")
    
    # 然后展示其他错误样本
    other_failures = [req for req in failed_requests if not req.get("upstream_error", False)]
    if other_failures:
        report.append("\n--- Other Error Samples ---")
        for i, req in enumerate(other_failures[:5]):  # 最多显示5个其他错误样本
            report.append(f"\nFailure Sample #{i+1} (Request ID: {req['request_id']})")
            report.append(f"Prompt: {req['prompt']}")
            report.append(f"Status Code: {req.get('status_code', 'N/A')}")
            
            if "error" in req:
                report.append(f"Error: {req['error']}")
            
            if "raw_response" in req:
                raw_resp = req['raw_response']
                # 如果响应太长，只显示前200个字符
                if len(raw_resp) > 200:
                    report.append(f"Raw Response (truncated): {raw_resp[:200]}...")
                else:
                    report.append(f"Raw Response: {raw_resp}")
            
            if "response_json" in req:
                try:
                    # 尝试获取错误消息
                    if "error" in req["response_json"]:
                        error_obj = req["response_json"]["error"]
                        if isinstance(error_obj, dict):
                            report.append(f"Error Type: {error_obj.get('type', 'Unknown')}")
                            report.append(f"Error Message: {error_obj.get('message', 'No message')}")
                        else:
                            report.append(f"Error Info: {error_obj}")
                except:
                    pass
    
    return "\n".join(report)

async def main():
    """主函数：发送请求并收集结果"""
    print(f"Preparing to send {QUESTION_COUNT} image generation requests with concurrency of {CONCURRENCY}")
    print(f"Using {len(PROMPTS)} different prompts")
    print(f"Single request timeout: {REQUEST_TIMEOUT} seconds")
    
    # 处理所有请求
    start_time = time.time()
    all_results = await process_requests()
    total_time = time.time() - start_time
    
    # 计算统计信息
    successful = [r for r in all_results if r["success"]]
    failed = [r for r in all_results if not r["success"]]
    upstream_errors = [r for r in all_results if r.get("upstream_error", False)]
    
    avg_time = sum(r["elapsed_time"] for r in all_results) / len(all_results) if all_results else 0
    requests_per_second = len(all_results) / total_time if total_time > 0 else 0
    
    # 打印结果摘要
    print(f"\n--- Results Summary ---")
    print(f"Total requests processed: {len(all_results)} of {QUESTION_COUNT}")
    print(f"Total execution time: {total_time:.2f} seconds")
    print(f"Successful: {len(successful)}")
    print(f"Failed: {len(failed)}")
    print(f"  - With 'Failed to request upstream address' error: {len(upstream_errors)}")
    print(f"  - With other errors: {len(failed) - len(upstream_errors)}")
    print(f"Average response time: {avg_time:.2f} seconds")
    print(f"Actual requests per second: {requests_per_second:.2f}")
    
    # 分析失败请求
    if failed:
        failure_report = await analyze_failures(failed)
        print("\n" + failure_report)
        
        # 保存失败分析报告
        async with aiofiles.open(f"{OUTPUT_DIR}/failure_analysis.txt", 'w', encoding='utf-8') as f:
            await f.write(failure_report)
        print(f"\nDetailed failure analysis saved to {OUTPUT_DIR}/failure_analysis.txt")
        
        # 保存所有失败请求的详细信息
        async with aiofiles.open(f"{OUTPUT_DIR}/failed_requests.json", 'w', encoding='utf-8') as f:
            await f.write(json.dumps(failed, ensure_ascii=False, indent=2))
        print(f"Failed request details saved to {OUTPUT_DIR}/failed_requests.json")
    
    # 保存汇总结果，但不包含成功请求的完整响应
    summary = {
        "total_requests": QUESTION_COUNT,
        "processed_requests": len(all_results),
        "total_execution_time": total_time,
        "successful_requests": len(successful),
        "failed_requests": len(failed),
        "upstream_errors": len(upstream_errors),
        "other_errors": len(failed) - len(upstream_errors),
        "average_response_time": avg_time,
        "actual_requests_per_second": requests_per_second,
        # 只保存失败的请求的完整信息
        "failed_results": failed
    }
    
    async with aiofiles.open(f"{OUTPUT_DIR}/summary.json", 'w', encoding='utf-8') as f:
        await f.write(json.dumps(summary, ensure_ascii=False, indent=2))
    
    print(f"Summary results saved to {OUTPUT_DIR}/summary.json")

if __name__ == "__main__":
    asyncio.run(main())