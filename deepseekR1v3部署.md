# 沐曦C500集群DeepSeek R1/V3模型部署技术文档(修订版)

## 1. 硬件配置清单

### 1.1 服务器规格要求 C500

| 项目 | 规格 | 数量 | 备注 |
|------|------|------|------|
| **服务器型号** | 沐曦C500 GPU服务器 | **8台** | 自带沐曦GPU，4台部署R1，4台部署V3 |
| **CPU** | Intel Xeon或AMD EPYC 64核+ | 8个 | 支持PCIe 5.0 |
| **内存** | DDR5 ECC 512GB+ | 8套 | 建议768GB |
| **系统盘** | NVMe SSD 500GB | 8块 | 高速系统盘 |
| **数据盘** | NVMe SSD 16TB+ | 8块 | 存储671B模型权重(约2.6TB每个模型) |
| **网卡** | 25Gbps以太网卡 | 8块 | 实际可达带宽 |
| **电源** | 冗余电源 3000W+ | 16个 | 双电源备份 |

### 1.2 集群分组规划

| 集群组 | 服务器编号 | 部署模型 | GPU配置 | 存储需求 |
|--------|-----------|----------|---------|----------|
| **R1集群** | C500-01~04 | **DeepSeek-R1-671B满血版** | 8卡×4台=32卡 | 2.6TB×4=10.4TB |
| **V3集群** | C500-05~08 | **DeepSeek-V3-671B** | 8卡×4台=32卡 | 2.6TB×4=10.4TB |
| **管理节点** | C500-01 | 兼任NFS+负载均衡 | - | 额外8TB共享存储 |

### 1.3 存储规划

| 存储类型 | 容量 | 用途 | 部署方式 |
|---------|------|------|---------|
| **系统存储** | 500GB×8 | 操作系统、软件 | 本地NVMe |
| **模型存储** | 16TB×8 | 模型权重、缓存 | 本地NVMe |
| **共享存储** | C500-01作为NFS | 模型备份、日志、配置 | 软件NFS |
| **临时存储** | 4TB×8 | 推理缓存 | 本地高速存储 |

---

## 2. 网络架构要求

### 2.1 网络拓扑（8台设备）

```
                  [外网接入]
                       |
                [核心交换机]
                       |
  ┌─────────────────────┼─────────────────────┐
  |                     |                     |
[管理网络]          [业务网络]           [存储网络]
1Gbps              25Gbps               10Gbps
  |                     |                     |
[IPMI/BMC]    [API网关+智能路由]      [内部NFS]
```

### 2.2 IP分配计划（8台设备）

| 服务器 | 集群组 | 管理IP | 业务IP | 集群IP | 存储IP |
|--------|--------|--------|--------|--------|--------|
| **C500-01** | R1集群 | 192.168.100.11 | 10.0.200.11 | 10.0.201.11 | 10.0.202.11 |
| **C500-02** | R1集群 | 192.168.100.12 | 10.0.200.12 | 10.0.201.12 | 10.0.202.12 |
| **C500-03** | R1集群 | 192.168.100.13 | 10.0.200.13 | 10.0.201.13 | 10.0.202.13 |
| **C500-04** | R1集群 | 192.168.100.14 | 10.0.200.14 | 10.0.201.14 | 10.0.202.14 |
| **C500-05** | V3集群 | 192.168.100.15 | 10.0.200.15 | 10.0.201.15 | 10.0.202.15 |
| **C500-06** | V3集群 | 192.168.100.16 | 10.0.200.16 | 10.0.201.16 | 10.0.202.16 |
| **C500-07** | V3集群 | 192.168.100.17 | 10.0.200.17 | 10.0.201.17 | 10.0.202.17 |
| **C500-08** | V3集群 | 192.168.100.18 | 10.0.200.18 | 10.0.201.18 | 10.0.202.18 |
| **API网关VIP** | 路由层 | - | 10.0.200.10 | - | - |

---

## 3. 模型自动路由方案

### 3.1 推荐方案：智能API网关路由

**优势**：
- ✅ 统一接口访问
- ✅ 自动模型识别和路由
- ✅ 负载均衡和故障转移
- ✅ 便于监控和管理

#### 3.1.1 路由架构

```
[用户请求] → [API网关] → [模型识别] → [路由决策]
                 ↓              ↓           ↓
            [统一接口]    [R1集群]    [V3集群]
                 ↓         4台C500     4台C500
            [响应聚合] ← [DeepSeek-R1] [DeepSeek-V3]
```

#### 3.1.2 路由配置示例

```yaml
# API网关路由配置
routes:
  - path: "/v1/chat/completions"
    method: POST
    handler: model_router
    rules:
      - condition: 'model == "deepseek-r1" or model == "deepseek-r1-671b"'
        target: "r1_cluster"
        endpoints:
          - "http://10.0.200.11:8000"
          - "http://10.0.200.12:8000"
          - "http://10.0.200.13:8000"
          - "http://10.0.200.14:8000"
      - condition: 'model == "deepseek-v3" or model == "deepseek-v3-671b"'
        target: "v3_cluster"
        endpoints:
          - "http://10.0.200.15:8000"
          - "http://10.0.200.16:8000"
          - "http://10.0.200.17:8000"
          - "http://10.0.200.18:8000"
```

### 3.3 API网关部署方案

#### 3.3.1 使用Kong网关

```bash
# 在C500-01上部署Kong网关
docker run -d --name kong-gateway \
  --network=host \
  -e "KONG_DATABASE=off" \
  -e "KONG_DECLARATIVE_CONFIG=/kong/kong.yml" \
  -e "KONG_PROXY_ACCESS_LOG=/dev/stdout" \
  -e "KONG_ADMIN_ACCESS_LOG=/dev/stdout" \
  -e "KONG_PROXY_ERROR_LOG=/dev/stderr" \
  -e "KONG_ADMIN_ERROR_LOG=/dev/stderr" \
  -e "KONG_ADMIN_LISTEN=0.0.0.0:8001" \
  -v /opt/kong:/kong \
  kong:3.4
```

#### 3.3.2 自定义路由插件

```python
# 模型路由插件示例
import json
from typing import Dict, List

class ModelRouter:
    def __init__(self):
        self.r1_endpoints = [
            "http://10.0.200.11:8000",
            "http://10.0.200.12:8000", 
            "http://10.0.200.13:8000",
            "http://10.0.200.14:8000"
        ]
        self.v3_endpoints = [
            "http://10.0.200.15:8000",
            "http://10.0.200.16:8000",
            "http://10.0.200.17:8000", 
            "http://10.0.200.18:8000"
        ]
    
    def route_request(self, request_body: Dict) -> List[str]:
        model = request_body.get("model", "").lower()
        
        if "r1" in model:
            return self.r1_endpoints
        elif "v3" in model:
            return self.v3_endpoints
        else:
            # 默认路由到R1集群
            return self.r1_endpoints
```

---

## 4. 模型部署配置

### 4.1 DeepSeek-R1-671B满血版部署

#### 4.1.1 模型下载（R1集群）
```bash
# 在C500-01~04上执行
cd /data/models
git clone https://huggingface.co/deepseek-ai/DeepSeek-R1-671B deepseek-r1-671b

# 验证模型大小（约2.6TB）
du -sh /data/models/deepseek-r1-671b
```

#### 4.1.2 vLLM部署配置（R1集群）
```yaml
# r1-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deepseek-r1-671b
spec:
  replicas: 4
  selector:
    matchLabels:
      app: deepseek-r1
  template:
    spec:
      containers:
      - name: deepseek-r1
        image: vllm/vllm-openai:latest
        command:
          - python
          - -m
          - vllm.entrypoints.openai.api_server
          - --model
          - /models/deepseek-r1-671b
          - --tensor-parallel-size
          - "8"
          - --gpu-memory-utilization
          - "0.95"
          - --max-model-len
          - "32768"
          - --port
          - "8000"
        resources:
          limits:
            nvidia.com/gpu: 8
        volumeMounts:
        - name: model-storage
          mountPath: /models
      volumes:
      - name: model-storage
        hostPath:
          path: /data/models
```

### 4.2 DeepSeek-V3-671B部署

#### 4.2.1 模型下载（V3集群）
```bash
# 在C500-05~08上执行
cd /data/models  
git clone https://huggingface.co/deepseek-ai/DeepSeek-V3 deepseek-v3-671b

# 验证模型大小（约2.6TB）
du -sh /data/models/deepseek-v3-671b
```

#### 4.2.2 vLLM部署配置（V3集群）
```yaml
# v3-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deepseek-v3-671b
spec:
  replicas: 4
  selector:
    matchLabels:
      app: deepseek-v3
  template:
    spec:
      containers:
      - name: deepseek-v3
        image: vllm/vllm-openai:latest
        command:
          - python
          - -m
          - vllm.entrypoints.openai.api_server
          - --model
          - /models/deepseek-v3-671b
          - --tensor-parallel-size
          - "8"
          - --gpu-memory-utilization
          - "0.95"
          - --max-model-len
          - "65536"
          - --port
          - "8000"
        resources:
          limits:
            nvidia.com/gpu: 8
        volumeMounts:
        - name: model-storage
          mountPath: /models
      volumes:
      - name: model-storage
        hostPath:
          path: /data/models
```

---

## 5. 监控系统（8台设备）

### 5.1 监控架构

```
[Prometheus] ← [Node Exporter×8] ← [8台C500服务器指标]
   ↓              ↓
[AlertManager] [沐曦GPU监控×8] ← [64张GPU指标]
   ↓              ↓
[Grafana] ← [Model Metrics×2] ← [R1+V3集群指标]
```

### 5.2 关键监控指标

| 指标类型 | 监控项目 | 告警阈值 | 备注 |
|---------|----------|----------|------|
| **硬件监控** | GPU利用率 | >95% | 64张GPU |
| **硬件监控** | 显存使用率 | >90% | 每张GPU |
| **集群监控** | R1集群可用性 | <3台 | 4台最少3台 |
| **集群监控** | V3集群可用性 | <3台 | 4台最少3台 |
| **性能监控** | API响应时间 | >15秒 | 671B模型 |
| **性能监控** | 吞吐量 | <100 tokens/s | 每个集群 |
| **路由监控** | 路由错误率 | >1% | API网关 |

---

## 6. 验收标准（修订版）

### 6.1 硬件验收
- [ ] **8台C500服务器**正常启动
- [ ] **64张沐曦GPU**正常识别和工作
- [ ] 网络连通性测试通过（8台设备）
- [ ] 存储读写性能达标（16TB×8）

### 6.2 软件验收
- [ ] 沐曦GPU驱动正常工作（64张GPU）
- [ ] Kubernetes集群状态正常（8节点）
- [ ] **DeepSeek-R1-671B满血版**加载成功（4台）
- [ ] **DeepSeek-V3-671B**加载成功（4台）
- [ ] API网关路由正常工作
- [ ] 模型自动识别和路由功能正常

### 6.3 性能验收
- [ ] **R1集群**：每秒200+ tokens（4台8卡）
- [ ] **V3集群**：每秒200+ tokens（4台8卡）
- [ ] API响应时间<15秒（671B模型）
- [ ] 系统稳定运行48小时
- [ ] 监控系统正常工作（8台设备）
- [ ] 负载均衡正常分发（双集群）
- [ ] 故障转移功能正常（集群内）

### 6.4 路由功能验收
- [ ] 通过model参数自动路由到正确集群
- [ ] 支持同时访问R1和V3模型
- [ ] 路由错误率<0.1%
- [ ] 集群间负载均衡正常

---

## 7. 部署时间规划

| 阶段 | 工作内容 | 预计时间 | 备注 |
|------|---------|----------|------|
| **硬件验收** | 8台C500验收、网络配置 | 2天 | 并行进行 |
| **基础环境** | 系统安装、驱动配置 | 2天 | 8台并行 |
| **集群搭建** | K8s集群、存储配置 | 1天 | 8节点集群 |
| **模型下载** | 下载R1+V3模型 | 3天 | 5.2TB总计 |
| **模型部署** | 双集群部署、配置调优 | 2天 | 并行部署 |
| **路由配置** | API网关、路由规则 | 1天 | 关键环节 |
| **监控配置** | 8台设备监控系统 | 1天 | 全面监控 |
| **测试优化** | 性能测试、问题修复 | 2天 | 双集群测试 |
| **文档交付** | 文档整理、培训 | 1天 | 总结交付 |
| **总计** | - | **15天** | 预留缓冲时间 |
