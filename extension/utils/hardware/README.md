# OADIN 系统硬件监控功能

OADIN系统现已集成了全面的硬件监控功能，可以实时获取系统的RAM和VRAM信息，支持跨平台GPU检测。

## 功能特性

### 内存监控
- **RAM监控**: 实时获取系统内存使用情况
- **VRAM监控**: 支持独立显卡显存监控
- **使用率计算**: 自动计算内存使用百分比
- **多单位显示**: 自动选择合适的单位（B, KB, MB, GB）

### GPU检测
- **多GPU支持**: 检测和监控多个GPU设备
- **跨平台兼容**: 支持Windows、Linux、macOS
- **厂商支持**: 
  - NVIDIA GPU (通过nvidia-smi)
  - AMD GPU (通过sysfs/调试接口)
  - Intel GPU (包括集成显卡和Arc系列)
- **详细信息**: GPU名称、显存容量、使用率、温度等

## API接口

### 1. 获取内存信息
```http
GET /api/system/memory
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "platform": "Microsoft Windows 11 Home China",
    "hostname": "LAPTOP-4SV0FIFP",
    "update_time": "2025-08-12T16:25:21+08:00",
    "ram_total": 33822867456,
    "ram_used": 27487776768,
    "ram_free": 6335090688,
    "ram_used_pct": 81.3,
    "vram_total": 2147479552,
    "vram_used": 0,
    "vram_free": 2147479552,
    "vram_used_pct": 0.0
  }
}
```

### 2. 获取GPU信息
```http
GET /api/system/gpu
```

**响应示例:**
```json
{
  "success": true,
  "count": 1,
  "data": [
    {
      "name": "Intel(R) Arc(TM) 140V GPU (16GB)",
      "memory_total": 2147479552,
      "memory_used": 0,
      "memory_free": 2147479552,
      "utilization": -1,
      "temperature": 0
    }
  ]
}
```

### 3. 获取完整硬件信息
```http
GET /api/system/hardware
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "memory": {
      "platform": "Microsoft Windows 11 Home China",
      "hostname": "LAPTOP-4SV0FIFP",
      "ram_total": 33822867456,
      "ram_used": 27487776768,
      "vram_total": 2147479552,
      "vram_used": 0
    },
    "gpus": [
      {
        "name": "Intel(R) Arc(TM) 140V GPU (16GB)",
        "memory_total": 2147479552,
        "memory_used": 0,
        "memory_free": 2147479552
      }
    ]
  }
}
```

## 代码使用示例

### Go语言调用
```go
import "oadin/extension/utils/hardware"

// 获取内存信息
memInfo, err := hardware.GetMemoryInfo()
if err != nil {
    log.Printf("获取内存信息失败: %v", err)
    return
}

fmt.Printf("RAM使用率: %.1f%%\n", memInfo.RAMUsedPct)
fmt.Printf("总内存: %s\n", hardware.FormatBytes(memInfo.RAMTotal))

// 获取GPU信息
gpuInfo, err := hardware.GetGPUInfo()
if err != nil {
    log.Printf("获取GPU信息失败: %v", err)
    return
}

for i, gpu := range gpuInfo {
    fmt.Printf("GPU %d: %s\n", i+1, gpu.Name)
    if gpu.MemoryTotal > 0 {
        fmt.Printf("  显存: %s\n", hardware.FormatBytes(gpu.MemoryTotal))
    }
}

// 打印详细信息
hardware.PrintMemoryInfo()
```

### 测试程序
项目包含了一个完整的测试程序：
```bash
cd /path/to/oadin
go run extension/utils/hardware/test/main.go
```

**输出示例:**
```
🚀 OADIN 系统硬件监控测试
================================

📊 获取内存信息...
✅ 内存信息获取成功
   系统: Microsoft Windows 11 Home China (LAPTOP-4SV0FIFP)
   RAM: 25.8 GB / 31.5 GB (使用率: 82.0%)
   VRAM: 0 B / 2.0 GB (使用率: 0.0%)

🎮 获取GPU信息...
✅ 检测到 1 个GPU设备:
   GPU 1: Intel(R) Arc(TM) 140V GPU (16GB)
         显存: 0 B / 2.0 GB
```

## 技术实现

### 跨平台支持
- **Windows**: 使用WMI/PowerShell和nvidia-smi
- **Linux**: 使用sysfs、调试接口和nvidia-smi
- **macOS**: 使用system_profiler和Metal框架

### 主要依赖
- `github.com/shirou/gopsutil/v3`: 系统信息获取
- 平台特定工具：nvidia-smi、wmic、PowerShell等

### 文件结构
```
extension/utils/hardware/
├── monitor.go          # 主要监控接口
├── gpu_windows.go      # Windows GPU检测
├── gpu_linux.go        # Linux GPU检测  
├── gpu_darwin.go       # macOS GPU检测
├── gpu_default.go      # 默认实现
└── test/
    └── main.go         # 测试程序
```

## 注意事项

1. **权限要求**: 某些GPU信息获取可能需要管理员权限
2. **兼容性**: 不同平台的GPU检测能力有差异
3. **性能**: GPU温度和使用率获取可能需要额外的系统调用
4. **准确性**: 集成显卡的显存信息可能不够精确

## 未来增强

- [ ] 实时监控和告警功能
- [ ] 历史数据记录和分析
- [ ] 更精确的GPU使用率检测
- [ ] 支持更多GPU厂商
- [ ] WebSocket实时推送
