import aiohttp
import asyncio
import aiofiles
import json
import time
import os
from tqdm import tqdm

# 配置参数
API_URL = "https://ai.burncloud.com/v1/chat/completions"
API_KEY = "sk-7wiO34EMbl9hvC8n9ev2Y0CPwO7e5H2tKMlzl8U253B9a342"  # 请替换为您的 API 密钥
MODEL = "gpt-4o-mini"
CONCURRENCY = 100  # 并发数量，调整以确保1分钟内完成1000个请求
MAX_TIME = 60  # 最大运行时间（秒）
QUESTION_COUNT = 1  # 要发送的请求数量

# 确保输出目录存在
OUTPUT_DIR = "responses"
os.makedirs(OUTPUT_DIR, exist_ok=True)

async def make_request(session, question_id):
    """发送单个请求到 API"""
    headers = {
        "Content-Type": "application/json",
        "Authorization": API_KEY
    }
    
    # 固定问题为 "answer yes."
    question = "answer yes."
    
    payload = {
        "model": MODEL,
        "messages": [
            {
                "role": "user",
                "content": question
            }
        ],
        "stream_options": {"include_usage": True},
        "stream": False
    }
    
    start_time = time.time()
    
    try:
        async with session.post(API_URL, headers=headers, json=payload) as response:
            response_json = await response.json()
            elapsed_time = time.time() - start_time
            
            # 提取回答文本
            answer = ""
            if response.status == 200 and "choices" in response_json and len(response_json["choices"]) > 0:
                answer = response_json["choices"][0].get("message", {}).get("content", "No answer")
            
            # 输出问题和回答到终端
            print(f"\nQuestion {question_id}: {question}")
            print(f"Answer: {answer[:100]}..." if len(answer) > 100 else f"Answer: {answer}")
            print(f"Response time: {elapsed_time:.2f}s")
            
            # 将响应保存到文件
            filename = f"{OUTPUT_DIR}/response_{question_id}.json"
            async with aiofiles.open(filename, 'w', encoding='utf-8') as f:
                result = {
                    "question": question,
                    "question_id": question_id,
                    "elapsed_time": elapsed_time,
                    "response": response_json,
                    "answer": answer
                }
                await f.write(json.dumps(result, ensure_ascii=False, indent=2))
            
            return {
                "question_id": question_id,
                "question": question,
                "status_code": response.status,
                "elapsed_time": elapsed_time,
                "success": response.status == 200,
                "filename": filename,
                "answer": answer
            }
    except Exception as e:
        elapsed_time = time.time() - start_time
        print(f"\nError processing question {question_id}")
        print(f"Error: {str(e)}")
        return {
            "question_id": question_id,
            "question": question,
            "status_code": None,
            "elapsed_time": elapsed_time,
            "success": False,
            "error": str(e)
        }

async def process_questions():
    """处理所有问题，确保在时间限制内完成"""
    start_time = time.time()
    time_limit_reached = False
    
    # 创建信号量来控制并发
    semaphore = asyncio.Semaphore(CONCURRENCY)
    
    async def bounded_request(question_id):
        async with semaphore:
            async with aiohttp.ClientSession() as session:
                return await make_request(session, question_id)
    
    # 创建所有任务
    tasks = [bounded_request(i+1) for i in range(QUESTION_COUNT)]
    
    # 使用tqdm显示进度
    results = []
    for future in tqdm(asyncio.as_completed(tasks), total=len(tasks), desc="Processing requests"):
        result = await future
        results.append(result)
        
        # 检查是否超时
        if time.time() - start_time > MAX_TIME:
            time_limit_reached = True
            print(f"\nReached time limit of {MAX_TIME} seconds. Processed {len(results)} out of {len(tasks)} questions.")
            # 不立即退出，让已经启动的请求完成
            break
    
    # 等待所有已经启动的请求完成
    if time_limit_reached:
        print("Waiting for in-progress requests to complete...")
    
    return results, time_limit_reached

async def main():
    """主函数：发送请求并收集结果"""
    print(f"Preparing to send {QUESTION_COUNT} requests with the fixed question: 'answer yes.'")
    
    # 处理所有问题
    start_time = time.time()
    all_results, time_limit_reached = await process_questions()
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
    
    # 分析停止原因
    if time_limit_reached:
        print(f"Script stopped due to time limit ({MAX_TIME} seconds)")
    else:
        print("Script completed processing all requests")
    
    # 保存汇总结果
    summary = {
        "total_requests": QUESTION_COUNT,
        "processed_requests": len(all_results),
        "total_execution_time": total_time,
        "successful_requests": len(successful),
        "failed_requests": len(failed),
        "average_response_time": avg_time,
        "time_limit_reached": time_limit_reached,
        "results": all_results
    }
    
    async with aiofiles.open(f"{OUTPUT_DIR}/summary.json", 'w', encoding='utf-8') as f:
        await f.write(json.dumps(summary, ensure_ascii=False, indent=2))
    
    print(f"Detailed results saved to {OUTPUT_DIR}/summary.json")

if __name__ == "__main__":
    asyncio.run(main())