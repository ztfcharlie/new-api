"""
这个一个给不同会员等级计算渠道销售价格的脚本（注意下面的定价原则）
对于同一商品有不同供应商和不同拿货价的情况下的定价原则
统一售价
	同一商品对用户展示统一售价
	避免因供应商不同造成用户混淆
	维持价格体系的简单透明
取高价基准
	以较高拿货价为基准定价
	确保在最不利情况下仍有合理利润
"""
def calculate_member_prices(cost_ratio):
    """
    计算会员价格
    :param cost_ratio: 拿货价相对于国际参考价的比例（例如0.6表示60%）
    """
    # 基础参数设置
    reference_price = 1
    cost_price = reference_price * cost_ratio
    tax_rate = 0.06
    
    # 计算高成本阈值
    HIGH_COST_THRESHOLD = 1 - tax_rate  # 0.94，即94%
    
    # 计算保本价格比例
    breakeven_ratio = cost_ratio / (1 - tax_rate)
    
    # 标准利润空间倍数（成本较低时使用）
    standard_markup_ratios = {
        "普通会员": 1.40,
        "银牌会员": 1.35,
        "金牌会员": 1.30,
        "铂金会员": 1.25,
        "钻石会员": 1.20,
        "保本组": 1.00  # 这个倍率会被实际的保本价覆盖
    }
    
    print(f"\n拿货价为国际参考价的 {cost_ratio:.1%}")
    print("-" * 50)
    
    # 判断是否为高成本情况
    is_high_cost = cost_ratio >= HIGH_COST_THRESHOLD
    
    if is_high_cost:
        print(f"【高成本保本策略】")
        print(f"由于成本较高（≥参考价的{HIGH_COST_THRESHOLD:.1%}），所有会员统一采用保本价格")
        print(f"保本价格比例：{breakeven_ratio:.3f}（覆盖成本和税费）")
    else:
        print("【标准利润策略】")
        print(f"保本价格比例：{breakeven_ratio:.3f}（覆盖成本和税费）")
    print("-" * 50)
    
    member_levels = ["保本组","普通会员", "银牌会员", "金牌会员", "铂金会员", "钻石会员"]
    
    for level in member_levels:
        if is_high_cost or level == "保本组":
            # 高成本情况或保本组：使用保本价
            final_ratio = breakeven_ratio
        else:
            # 正常情况：使用标准加价
            cost_based_ratio = (cost_price * standard_markup_ratios[level]) / reference_price
            final_ratio = min(1.0, cost_based_ratio)
        
        # 计算实际数据
        selling_price = final_ratio * reference_price
        tax = selling_price * tax_rate
        profit = selling_price - tax - cost_price
        profit_rate = profit / selling_price
        
        # 计算相对于成本价的加价比例
        markup_over_cost = (selling_price / cost_price - 1) * 100
        
        print(f"{level}:")
        print(f"  售价比例: {final_ratio:.3f} ({final_ratio:.1%})")
        print(f"  利润率: {profit_rate:.1%}")
        print(f"  较成本加价: {markup_over_cost:.1f}%")
        
        if level == "保本组":
            print("  📌 这是理论最低价格，仅覆盖成本和税费")
        elif profit_rate < 0:
            print("  ⚠️ 警告：当前价格会亏损！")
        elif profit_rate < 0.02 and not is_high_cost:
            print("  ⚠️ 警告：利润率较低！")

def main():
    print("会员价格计算器")
    print("=" * 50)
    print("说明：")
    print("1. 输入拿货价相对于国际参考价的比例")
    print("2. 当成本达到参考价的94%时，自动启用保本策略")
    print("3. 保本策略下所有会员统一价格，仅覆盖成本和税费")
    print("4. 税率设定为6%")
    print("=" * 50)
    print("示例：")
    print("- 0.6 表示拿货价为国际参考价的60%")
    print("- 0.95 表示拿货价为国际参考价的95%")
    print("- 输入q退出程序")
    print("=" * 50)
    
    while True:
        try:
            user_input = input("\n请输入比例 > ").strip().lower()
            
            if user_input == 'q':
                print("\n程序已退出")
                break
                
            ratio = float(user_input)
            if ratio <= 0:
                print("❌ 错误：比例必须大于0！")
                continue
            if ratio > 1:
                print("⚠️ 注意：拿货价高于参考价，将采用保本策略")
                
            calculate_member_prices(ratio)
            
        except ValueError:
            print("❌ 错误：输入无效！请输入有效的数字（例如：0.6）")
        except Exception as e:
            print(f"❌ 发生错误：{str(e)}")

if __name__ == "__main__":
    main()
