import aiohttp
import asyncio
import aiofiles
import json
import time
import os
from tqdm import tqdm

# 配置参数
API_URL = "https://ai.burncloud.com/v1/images/generations"
API_KEY = "sk-9DufUZISqD4FrlVzvO8nkioqEwxqjPkJ2OwSAibs7d9OKJLe" #burncloud test token
MODEL = "gpt-image-1"
CONCURRENCY = 100  # 1秒并发数量，100张图
MAX_TIME = 10  # 最大运行时间（秒）
QUESTION_COUNT = CONCURRENCY * MAX_TIME  # 要发送的请求总数量 (100*10=1000)

# 确保输出目录存在
OUTPUT_DIR = "responses"
os.makedirs(OUTPUT_DIR, exist_ok=True)

async def make_request(session, request_id):
    """发送单个请求到 API"""
    headers = {
        "Content-Type": "application/json",
        "Authorization": API_KEY
    }
    
    # 图像生成的请求体
    payload = {
        "model": MODEL,
        "prompt": "一个性感的动作，职业装，办公室，职业女性，带眼镜，苗条，白人",
        "n": 1,
        "size": "1024x1024",
        "quality": "medium",
        "responseFormat": "b64_json",
        "user": f"stress-test-{request_id}"
    }
    
    start_time = time.time()
    
    try:
        async with session.post(API_URL, headers=headers, json=payload) as response:
            response_text = await response.text()
            elapsed_time = time.time() - start_time
            
            # 尝试解析JSON响应
            try:
                response_json = json.loads(response_text)
                
                # 如果成功，移除base64数据以节省空间
                if response.status == 200 and "data" in response_json:
                    for item in response_json["data"]:
                        if "b64_json" in item:
                            # 只保留base64字符串的长度信息而不是内容本身
                            b64_length = len(item["b64_json"])
                            item["b64_json"] = f"[BASE64_DATA_REMOVED - {b64_length} bytes]"
            except json.JSONDecodeError:
                response_json = {"error": "Failed to decode JSON", "raw_response": response_text}
            
            # 输出请求信息到终端
            print(f"\nRequest {request_id}")
            print(f"Status code: {response.status}")
            print(f"Response time: {elapsed_time:.2f}s")
            
            success = response.status == 200
            if success:
                print(f"Image generation successful")
            else:
                print(f"Image generation failed: {response_text[:200]}..." if len(response_text) > 200 else response_text)
            
            return {
                "request_id": request_id,
                "status_code": response.status,
                "elapsed_time": elapsed_time,
                "success": success,
                "response_json": response_json,
                # 只有失败时才保存完整的响应文本
                "raw_response": response_text if not success else "[RESPONSE_TEXT_REMOVED]"
            }
    except Exception as e:
        elapsed_time = time.time() - start_time
        error_details = str(e)
        print(f"\nError processing request {request_id}")
        print(f"Error: {error_details}")
        return {
            "request_id": request_id,
            "status_code": None,
            "elapsed_time": elapsed_time,
            "success": False,
            "error": error_details,
            "error_type": type(e).__name__
        }

async def process_requests():
    """处理所有请求，确保在时间限制内完成"""
    start_time = time.time()
    
    # 创建信号量来控制并发
    semaphore = asyncio.Semaphore(CONCURRENCY)
    
    async def bounded_request(request_id):
        async with semaphore:
            async with aiohttp.ClientSession() as session:
                return await make_request(session, request_id)
    
    # 创建所有任务
    tasks = [bounded_request(i+1) for i in range(QUESTION_COUNT)]
    
    # 使用tqdm显示进度
    results = []
    for future in tqdm(asyncio.as_completed(tasks), total=len(tasks), desc="Processing requests"):
        result = await future
        results.append(result)
        
        # 检查是否超时
        if time.time() - start_time > MAX_TIME:
            print(f"\nReached time limit of {MAX_TIME} seconds. Processed {len(results)} out of {len(tasks)} requests.")
            # 不立即退出，让已经启动的请求完成
            break
    
    return results

async def analyze_failures(failed_requests):
    """分析失败的请求并生成详细报告"""
    if not failed_requests:
        return "No failed requests to analyze."
    
    # 按错误类型分组
    error_types = {}
    status_codes = {}
    
    for req in failed_requests:
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
    for i, req in enumerate(failed_requests[:10]):  # 最多显示10个失败样本
        report.append(f"\nFailure Sample #{i+1} (Request ID: {req['request_id']})")
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
    print(f"Preparing to send {QUESTION_COUNT} image generation requests at {CONCURRENCY} requests per second for {MAX_TIME} seconds")
    
    # 处理所有请求
    start_time = time.time()
    all_results = await process_requests()
    total_time = time.time() - start_time
    
    # 计算统计信息
    successful = [r for r in all_results if r["success"]]
    failed = [r for r in all_results if not r["success"]]
    
    avg_time = sum(r["elapsed_time"] for r in all_results) / len(all_results) if all_results else 0
    
    # 打印结果摘要
    print(f"\n--- Results Summary ---")
    print(f"Total requests processed: {len(all_results)} of {QUESTION_COUNT}")
    print(f"Total execution time: {total_time:.2f} seconds")
    print(f"Successful: {len(successful)}")
    print(f"Failed: {len(failed)}")
    print(f"Average response time: {avg_time:.2f} seconds")
    
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
        "average_response_time": avg_time,
        # 只保存失败的请求的完整信息
        "failed_results": failed
    }
    
    async with aiofiles.open(f"{OUTPUT_DIR}/summary.json", 'w', encoding='utf-8') as f:
        await f.write(json.dumps(summary, ensure_ascii=False, indent=2))
    
    print(f"Summary results saved to {OUTPUT_DIR}/summary.json")

if __name__ == "__main__":
    asyncio.run(main())