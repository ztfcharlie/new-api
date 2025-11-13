#!/usr/bin/env python3
"""
专门用于测试nginx IP限流配置的简化脚本
快速发送200个请求，检测IP限流是否生效
"""

import requests
import threading
import time
from concurrent.futures import ThreadPoolExecutor
import sys

def check_rate_limit(url, num_requests=200, concurrent=20):
    """检查nginx IP限流是否生效"""

    print(f"🚀 开始测试nginx IP限流配置")
    print(f"📍 目标URL: {url}")
    print(f"📊 发送请求数: {num_requests}")
    print(f"⚡ 并发数: {concurrent}")
    print("-" * 50)

    results = {
        'total': 0,
        'success': 0,
        'rate_limited': 0,
        'status_codes': {},
        'rate_limit_responses': []
    }

    lock = threading.Lock()

    def send_request(req_id):
        """发送单个请求"""
        try:
            headers = {
                'User-Agent': f'RateLimitTest/{req_id}',
                'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
            }

            response = requests.get(url, headers=headers, timeout=5)
            content = response.text[:300]  # 只取前300字符

            # 检测限流关键词
            is_rate_limited = False
            rate_limit_indicators = [
                "too many requests", "rate limit", "429",
                "请求过于频繁", "访问过于频繁", "连接数过多",
                "service unavailable", "503", "502"
            ]

            content_lower = content.lower()
            for indicator in rate_limit_indicators:
                if indicator in content_lower or str(response.status_code) in indicator:
                    is_rate_limited = True
                    break

            with lock:
                results['total'] += 1
                if response.status_code == 200 and not is_rate_limited:
                    results['success'] += 1
                elif is_rate_limited:
                    results['rate_limited'] += 1
                    results['rate_limit_responses'].append({
                        'req_id': req_id,
                        'status_code': response.status_code,
                        'content': content
                    })

                results['status_codes'][response.status_code] = results['status_codes'].get(response.status_code, 0) + 1

            return {
                'req_id': req_id,
                'status_code': response.status_code,
                'is_rate_limited': is_rate_limited,
                'content': content[:100]
            }

        except Exception as e:
            with lock:
                results['total'] += 1
                results['status_codes']['ERROR'] = results['status_codes'].get('ERROR', 0) + 1
            return None

    # 执行并发请求
    start_time = time.time()

    with ThreadPoolExecutor(max_workers=concurrent) as executor:
        futures = [executor.submit(send_request, i) for i in range(1, num_requests + 1)]

        # 等待所有请求完成
        for future in futures:
            try:
                future.result()
            except:
                pass

    total_time = time.time() - start_time

    # 显示结果
    print(f"\n📈 测试结果:")
    print(f"⏱️  总耗时: {total_time:.2f}秒")
    print(f"📊 总请求数: {results['total']}")
    print(f"✅ 成功请求: {results['success']}")
    print(f"🚫 被限流请求: {results['rate_limited']}")
    print(f"📈 成功率: {(results['success']/results['total'])*100:.1f}%")
    print(f"🚫 限流比例: {(results['rate_limited']/results['total'])*100:.1f}%")

    print(f"\n📊 HTTP状态码分布:")
    for code, count in sorted(results['status_codes'].items(), key=lambda x: str(x[0])):
        print(f"   {code}: {count}")

    # 显示限流响应示例
    if results['rate_limit_responses']:
        print(f"\n🚫 限流响应示例 (前3个):")
        for i, resp in enumerate(results['rate_limit_responses'][:3]):
            print(f"   {i+1}. 请求{resp['req_id']} (状态码{resp['status_code']}):")
            print(f"      {resp['content']}...")
    else:
        print(f"\n✅ 未检测到限流响应")

    # 分析结果
    print(f"\n🔍 IP限流配置分析:")
    rate_limit_percentage = (results['rate_limited'] / results['total']) * 100

    if rate_limit_percentage >= 10:
        print(f"   ✅ IP限流配置已生效 ({rate_limit_percentage:.1f}%请求被限流)")
    elif rate_limit_percentage >= 1:
        print(f"   ⚠️  IP限流部分生效 ({rate_limit_percentage:.1f}%请求被限流，可能需要调整阈值)")
    else:
        print(f"   ❌ IP限流未生效或配置过弱 ({rate_limit_percentage:.1f}%请求被限流)")

    # 建议配置
    print(f"\n💡 配置建议:")
    if rate_limit_percentage == 0:
        print(f"   建议检查nginx配置中的limit_req_zone和limit_req指令")
        print(f"   可能需要降低限流阈值以测试效果")
    elif rate_limit_percentage < 5:
        print(f"   当前限流较宽松，如需测试可以尝试:")
        print(f"   - 增加并发数到50-100")
        print(f"   - 增加请求数到500-1000")
        print(f"   - 降低nginx的限流阈值")
    else:
        print(f"   IP限流工作正常!")

    print("-" * 50)

    return rate_limit_percentage

def main():
    if len(sys.argv) < 2:
        print("用法: python test_nginx_rate_limit.py <URL> [请求数] [并发数]")
        print("示例: python test_nginx_rate_limit.py http://35.239.167.189:3000/console 200 20")
        sys.exit(1)

    url = sys.argv[1]
    num_requests = int(sys.argv[2]) if len(sys.argv) > 2 else 200
    concurrent = int(sys.argv[3]) if len(sys.argv) > 3 else 20

    # 快速测试
    rate_limit_percentage = check_rate_limit(url, num_requests, concurrent)

    # 如果首次测试限流比例低，建议进行压力测试
    if rate_limit_percentage < 5 and rate_limit_percentage > 0:
        print(f"\n🔄 首次测试限流比例较低，进行压力测试...")
        check_rate_limit(url, num_requests * 2, concurrent * 2)
    elif rate_limit_percentage == 0:
        print(f"\n🔄 未检测到限流，尝试更高压力测试...")
        check_rate_limit(url, 500, 50)

if __name__ == '__main__':
    main()