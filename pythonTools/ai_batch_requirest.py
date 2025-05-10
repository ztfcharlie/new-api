import aiohttp
import asyncio
import aiofiles
import json
import time
import os
from tqdm import tqdm
import random

# 配置参数
API_URL = "https://gala.chataiapi.com/v1/chat/completions"
API_KEY = "sk-muyz3cHdpiKCLlH7cBsH0BSfMtQNveY9GshmdFqMnGU6LJ4n"  # 请替换为您的 API 密钥
MODEL = "gpt-4o-mini"
CONCURRENCY = 50  # 并发数量，调整以确保1分钟内完成1000个请求
QUESTIONS_FILE = "questions.txt"  # 包含问题的文本文件
MAX_TIME = 60  # 最大运行时间（秒）

# 确保输出目录存在
OUTPUT_DIR = "responses"
os.makedirs(OUTPUT_DIR, exist_ok=True)

async def read_questions(file_path):
    """从文件中读取问题列表，并统计重复问题"""
    try:
        async with aiofiles.open(file_path, 'r', encoding='utf-8') as f:
            questions = [line.strip() for line in await f.readlines()]
            # 过滤掉空行
            questions = [q for q in questions if q]
            
            # 统计重复问题
            question_count = {}
            for q in questions:
                if q in question_count:
                    question_count[q] += 1
                else:
                    question_count[q] = 1
                    
            duplicates = {q: count for q, count in question_count.items() if count > 1}
            total_duplicates = sum(count - 1 for count in question_count.values() if count > 1)
            
            print(f"Total questions in file: {len(questions)}")
            print(f"Unique questions: {len(question_count)}")
            print(f"Duplicate questions: {total_duplicates}")
            if duplicates:
                print("Examples of duplicated questions:")
                for i, (q, count) in enumerate(duplicates.items()):
                    if i < 5:  # 只显示前5个重复问题示例
                        print(f"  - '{q}' appears {count} times")
            
            # 确保问题唯一
            unique_questions = list(dict.fromkeys(questions))
            return unique_questions
    except Exception as e:
        print(f"Error reading questions file: {e}")
        return []

async def make_request(session, question_id, question):
    """发送单个请求到 API"""
    headers = {
        "Content-Type": "application/json",
        "Authorization": API_KEY
    }
    
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
        print(f"\nError processing question {question_id}: {question}")
        print(f"Error: {str(e)}")
        return {
            "question_id": question_id,
            "question": question,
            "status_code": None,
            "elapsed_time": elapsed_time,
            "success": False,
            "error": str(e)
        }

async def process_questions(questions):
    """处理所有问题，确保在时间限制内完成"""
    start_time = time.time()
    time_limit_reached = False
    
    # 创建信号量来控制并发
    semaphore = asyncio.Semaphore(CONCURRENCY)
    
    async def bounded_request(question_id, question):
        async with semaphore:
            async with aiohttp.ClientSession() as session:
                return await make_request(session, question_id, question)
    
    # 创建所有任务
    tasks = [bounded_request(i+1, question) for i, question in enumerate(questions)]
    
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
    """主函数：读取问题、发送请求并收集结果"""
    # 读取问题
    questions = await read_questions(QUESTIONS_FILE)
    
    if not questions:
        print("No questions found or error reading questions file.")
        return
    
    print(f"Loaded {len(questions)} unique questions.")
    
    # 处理所有问题
    start_time = time.time()
    all_results, time_limit_reached = await process_questions(questions)
    total_time = time.time() - start_time
    
    # 计算统计信息
    successful = [r for r in all_results if r["success"]]
    failed = [r for r in all_results if not r["success"]]
    
    avg_time = sum(r["elapsed_time"] for r in all_results) / len(all_results) if all_results else 0
    
    # 打印结果摘要
    print(f"\n--- Results Summary ---")
    print(f"Total questions processed: {len(all_results)} of {len(questions)}")
    print(f"Total execution time: {total_time:.2f} seconds")
    print(f"Successful: {len(successful)}")
    print(f"Failed: {len(failed)}")
    print(f"Average response time: {avg_time:.2f} seconds")
    
    # 分析停止原因
    if time_limit_reached:
        print(f"Script stopped due to time limit ({MAX_TIME} seconds)")
    else:
        print("Script completed processing all questions")
    
    # 保存汇总结果
    summary = {
        "total_questions": len(questions),
        "processed_questions": len(all_results),
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
