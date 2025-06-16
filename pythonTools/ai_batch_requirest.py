import aiohttp
import asyncio
import aiofiles
import json
import time
import os
from tqdm import tqdm
import random
import uuid

# 配置参数
API_URL = "https://hk.burncloud.com/v1/chat/completions"
API_KEY = "sk-UAj9ccP3URcRPm1L83YNMhnDTCf36Rn0OIlL7nrAiiGIB17D"  # 请替换为您的 API 密钥
MODEL = "deepseek-v3-0324"
REQUESTS_PER_SECOND = 50  # 每秒发送的请求数
MAX_CONCURRENT_REQUESTS = 500  # 最大并发请求数量
QUESTIONS_FILE = "questions.txt"  # 包含问题的文本文件
MAX_TIME = 60  # 最大运行时间（秒）
TOTAL_REQUESTS = 2000  # 总共要发送的请求数量

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

async def make_request(session, request_id, question_id, question, cycle_num=0):
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
            print(f"\nRequest {request_id}: Question {question_id} (Cycle {cycle_num}): {question}")
            print(f"Answer: {answer[:100]}..." if len(answer) > 100 else f"Answer: {answer}")
            print(f"Response time: {elapsed_time:.2f}s")
            
            # 将响应保存到文件 - 使用唯一的请求ID来避免覆盖
            filename = f"{OUTPUT_DIR}/request_{request_id}_q{question_id}_cycle{cycle_num}.json"
            async with aiofiles.open(filename, 'w', encoding='utf-8') as f:
                result = {
                    "request_id": request_id,
                    "question": question,
                    "question_id": question_id,
                    "cycle": cycle_num,
                    "elapsed_time": elapsed_time,
                    "response": response_json,
                    "answer": answer,
                    "timestamp": time.time()
                }
                await f.write(json.dumps(result, ensure_ascii=False, indent=2))
            
            return {
                "request_id": request_id,
                "question_id": question_id,
                "cycle": cycle_num,
                "question": question,
                "status_code": response.status,
                "elapsed_time": elapsed_time,
                "success": response.status == 200,
                "filename": filename,
                "answer": answer,
                "timestamp": time.time()
            }
    except Exception as e:
        elapsed_time = time.time() - start_time
        print(f"\nError processing request {request_id}: Question {question_id} (Cycle {cycle_num}): {question}")
        print(f"Error: {str(e)}")
        return {
            "request_id": request_id,
            "question_id": question_id,
            "cycle": cycle_num,
            "question": question,
            "status_code": None,
            "elapsed_time": elapsed_time,
            "success": False,
            "error": str(e),
            "timestamp": time.time()
        }

async def process_questions(questions):
    """处理所有问题，每秒发送固定数量的请求，并控制最大并发数"""
    start_time = time.time()
    time_limit_reached = False
    
    # 创建信号量来控制最大并发数
    max_concurrency_semaphore = asyncio.Semaphore(MAX_CONCURRENT_REQUESTS)
    
    # 创建一个共享的HTTP会话
    async with aiohttp.ClientSession() as session:
        # 计算需要循环的次数
        total_cycles = (TOTAL_REQUESTS + len(questions) - 1) // len(questions)
        print(f"Will cycle through questions approximately {total_cycles} times to reach {TOTAL_REQUESTS} requests")
        
        # 生成所有需要处理的问题（可能会重复）
        all_questions = []
        for cycle in range(total_cycles):
            for i, question in enumerate(questions):
                if len(all_questions) < TOTAL_REQUESTS:
                    # 使用全局唯一ID作为请求ID
                    request_id = len(all_questions) + 1
                    all_questions.append((request_id, i+1, question, cycle+1))
        
        # 将问题分成每秒处理的批次
        batches = [all_questions[i:i+REQUESTS_PER_SECOND] for i in range(0, len(all_questions), REQUESTS_PER_SECOND)]
        
        all_tasks = []
        results = []
        active_tasks = set()
        completed_count = 0
        
        # 创建进度条
        progress_bar = tqdm(total=len(all_questions), desc="Processing requests")
        
        # 每秒发送一批请求
        for batch_idx, batch in enumerate(batches):
            if time.time() - start_time > MAX_TIME:
                time_limit_reached = True
                print(f"\nReached time limit of {MAX_TIME} seconds.")
                break
                
            batch_tasks = []
            for request_id, question_id, question, cycle_num in batch:
                # 使用信号量控制最大并发数
                async def bounded_request():
                    async with max_concurrency_semaphore:
                        return await make_request(session, request_id, question_id, question, cycle_num)
                
                task = asyncio.create_task(bounded_request())
                batch_tasks.append(task)
                all_tasks.append(task)
                active_tasks.add(task)
                
                # 设置回调来处理完成的任务
                def task_done_callback(completed_task):
                    nonlocal completed_count
                    active_tasks.discard(completed_task)
                    completed_count += 1
                    progress_bar.update(1)
                
                task.add_done_callback(task_done_callback)
            
            # 不等待批次完成，直接等待1秒后发送下一批
            await asyncio.sleep(1)
            
            # 打印当前状态
            print(f"\nSent batch {batch_idx+1}/{len(batches)} ({len(batch)} requests). "
                  f"Active requests: {len(active_tasks)}, Completed: {completed_count}")
        
        # 等待所有已启动的请求完成
        if all_tasks:
            print("\nWaiting for all in-progress requests to complete...")
            for task in tqdm(asyncio.as_completed(all_tasks), total=len(all_tasks), desc="Completing requests"):
                result = await task
                results.append(result)
        
        progress_bar.close()
    
    return results, time_limit_reached

async def main():
    """主函数：读取问题、发送请求并收集结果"""
    # 读取问题
    questions = await read_questions(QUESTIONS_FILE)
    
    if not questions:
        print("No questions found or error reading questions file.")
        return
    
    print(f"Loaded {len(questions)} unique questions.")
    print(f"Configuration: {REQUESTS_PER_SECOND} requests/second, max {MAX_CONCURRENT_REQUESTS} concurrent requests")
    print(f"Target: {TOTAL_REQUESTS} total requests")
    
    # 处理所有问题
    start_time = time.time()
    all_results, time_limit_reached = await process_questions(questions)
    total_time = time.time() - start_time
    
    # 计算统计信息
    successful = [r for r in all_results if r["success"]]
    failed = [r for r in all_results if not r["success"]]
    
    avg_time = sum(r["elapsed_time"] for r in all_results) / len(all_results) if all_results else 0
    
    # 统计每个循环的请求数
    cycles = {}
    for r in all_results:
        cycle = r.get("cycle", 0)
        if cycle in cycles:
            cycles[cycle] += 1
        else:
            cycles[cycle] = 1
    
    # 打印结果摘要
    print(f"\n--- Results Summary ---")
    print(f"Total requests processed: {len(all_results)} of {TOTAL_REQUESTS}")
    print(f"Total execution time: {total_time:.2f} seconds")
    print(f"Successful: {len(successful)}")
    print(f"Failed: {len(failed)}")
    print(f"Average response time: {avg_time:.2f} seconds")
    print(f"Requests per second: {len(all_results)/total_time:.2f}")
    
    print("\nRequests by cycle:")
    for cycle, count in sorted(cycles.items()):
        print(f"  Cycle {cycle}: {count} requests")
    
    # 分析停止原因
    if time_limit_reached:
        print(f"Script stopped due to time limit ({MAX_TIME} seconds)")
    else:
        print("Script completed processing all requests")
    
    # 保存汇总结果
    summary = {
        "total_questions_in_file": len(questions),
        "total_requests_target": TOTAL_REQUESTS,
        "processed_requests": len(all_results),
        "total_execution_time": total_time,
        "successful_requests": len(successful),
        "failed_requests": len(failed),
        "average_response_time": avg_time,
        "requests_per_second": len(all_results)/total_time if total_time > 0 else 0,
        "time_limit_reached": time_limit_reached,
        "cycles": cycles,
        "results": all_results
    }
    
    async with aiofiles.open(f"{OUTPUT_DIR}/summary.json", 'w', encoding='utf-8') as f:
        await f.write(json.dumps(summary, ensure_ascii=False, indent=2))
    
    print(f"Detailed results saved to {OUTPUT_DIR}/summary.json")

if __name__ == "__main__":
    asyncio.run(main())