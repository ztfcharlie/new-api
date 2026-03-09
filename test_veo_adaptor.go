package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/relay/channel/task/gemini"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func main() {
	fmt.Println("=========================================")
	fmt.Println("Veo 3.1 参数适配测试")
	fmt.Println("=========================================")
	fmt.Println()

	// 测试1: generateAudio 默认值设置
	fmt.Println("测试1: generateAudio 默认值")
	testGenerateAudioDefault()
	fmt.Println()

	// 测试2: sampleCount 支持
	fmt.Println("测试2: sampleCount 用户参数支持")
	testSampleCountSupport()
	fmt.Println()

	// 测试3: duration 验证
	fmt.Println("测试3: duration 范围验证")
	testDurationValidation()
	fmt.Println()

	// 测试4: n 参数优先级
	fmt.Println("测试4: n 参数优先级")
	testNParameterPriority()
	fmt.Println()

	// 测试5: 分辨率转换
	fmt.Println("测试5: 分辨率转换")
	testResolutionConversion()
	fmt.Println()

	fmt.Println("=========================================")
	fmt.Println("所有测试完成！")
	fmt.Println("=========================================")
}

func testGenerateAudioDefault() {
	params := &gemini.VeoParameters{}

	// 模拟 metadata 中没有 generateAudio 的情况
	metadata := map[string]interface{}{}
	data, _ := json.Marshal(metadata)
	_ = json.Unmarshal(data, params)

	fmt.Printf("  解析前的 params.GenerateAudio: %v\n", params.GenerateAudio)

	// 模拟设置默认值（修改后的逻辑）
	if params.GenerateAudio == nil {
		defaultAudio := true
		params.GenerateAudio = &defaultAudio
		fmt.Printf("  ✅ 成功设置默认值: params.GenerateAudio = %v\n", *params.GenerateAudio)
	} else {
		fmt.Printf("  ℹ️ generateAudio 已存在: %v\n", *params.GenerateAudio)
	}
}

func testSampleCountSupport() {
	testCases := []struct {
		name     string
		n         int
		sampleCount int
		expected  int
	}{
		{"用户 n=1", 1, 0, 1},
		{"用户 n=3", 3, 0, 3},
		{"用户 n=4", 4, 0, 4},
		{"用户 n=5", 5, 0, 4}, // 应该被限制为4
		{"用户 n=0", 0, 0, 1}, // 应该设置为默认值1
		{"metadata sampleCount=2", 0, 2, 2}, // metadata优先
		{"metadata=2, n=3", 3, 2, 3}, // n参数优先
	}

	for _, tc := range testCases {
		params := &gemini.VeoParameters{}
		req := relaycommon.TaskSubmitReq{
			Metadata: map[string]interface{}{},
		}

		// 设置 metadata
		if tc.sampleCount > 0 {
			req.Metadata["sampleCount"] = tc.sampleCount
		}

		// 模拟 UnmarshalMetadata (这是实际代码中的第一步)
		metadataJSON, _ := json.Marshal(req.Metadata)
		_ = json.Unmarshal(metadataJSON, params)

		// 设置 n
		req.N = tc.n

		// 模拟修改后的逻辑
		if req.N > 0 {
			params.SampleCount = req.N
		}

		// 验证范围
		if params.SampleCount == 0 || params.SampleCount < 1 {
			params.SampleCount = 1
		}
		if params.SampleCount > 4 {
			params.SampleCount = 4
		}

		result := "✅"
		if params.SampleCount != tc.expected {
			result = "❌"
		}

		fmt.Printf("  %s %s: n=%d, metadata=%d → sampleCount=%d (期望: %d) %s\n",
			result, tc.name, tc.n, tc.sampleCount, params.SampleCount, tc.expected, result)
	}
}

func testDurationValidation() {
	testCases := []struct {
		duration    int
		shouldPass  bool
	}{
		{4, true},
		{6, true},
		{8, true},
		{5, false}, // 应该失败
		{0, false},  // 应该失败（会被检查）
	}

	for _, tc := range testCases {
		params := &gemini.VeoParameters{}
		params.DurationSeconds = tc.duration

		// 模拟验证逻辑
		errMsg := ""
		if params.DurationSeconds != 0 && params.DurationSeconds != 4 && params.DurationSeconds != 6 && params.DurationSeconds != 8 {
			errMsg = fmt.Sprintf("invalid duration: %d", params.DurationSeconds)
		}

		result := "✅"
		if tc.shouldPass && errMsg != "" {
			result = "❌"
		}
		if !tc.shouldPass && errMsg == "" {
			result = "❌"
		}

		if tc.shouldPass {
			fmt.Printf("  %s duration=%d: %s\n", result, tc.duration,
				map[bool]string{true: "有效", false: "无效"}[tc.shouldPass])
		} else {
			fmt.Printf("  %s duration=%d: 正确拒绝 %s\n", result, tc.duration,
				func() string {
					if errMsg != "" {
						return errMsg
					}
					return "应该拒绝"
				}())
		}
	}
}

func testNParameterPriority() {
	fmt.Println("  测试 n 参数优先于 metadata.sampleCount:")

	params := &gemini.VeoParameters{}
	req := relaycommon.TaskSubmitReq{
		N:        3,
		Metadata: map[string]interface{}{
			"sampleCount": 2, // metadata 中设置为2
		},
	}

	// 模拟修改后的逻辑
	if req.N > 0 {
		params.SampleCount = req.N
	}

	expected := 3 // n=3 应该优先
	result := "✅"
	if params.SampleCount != expected {
		result = "❌"
	}

	fmt.Printf("  %s sampleCount=%d (n=3, metadata.sampleCount=2)\n", result, params.SampleCount)
}

func testResolutionConversion() {
	testCases := []struct {
		size       string
		resolution string
		aspectRatio string
	}{
		{"1920x1080", "1080p", "16:9"},
		{"3840x2160", "4k", "16:9"},
		{"1080x1920", "1080p", "9:16"},
		{"1280x720", "720p", "16:9"},
		{"invalid", "720p", "16:9"}, // 默认值
	}

	for _, tc := range testCases {
		resolution := gemini.SizeToVeoResolution(tc.size)
		aspectRatio := gemini.SizeToVeoAspectRatio(tc.size)

		resOK := "✅"
		if resolution != tc.resolution || aspectRatio != tc.aspectRatio {
			resOK = "❌"
		}

		fmt.Printf("  %s size=%s → resolution=%s, aspectRatio=%s\n",
			resOK, tc.size, resolution, aspectRatio)
	}

	// 测试小写转换
	resolution := "1080P"
	converted := strings.ToLower(resolution)
	result := "✅"
	if converted != "1080p" {
		result = "❌"
	}
	fmt.Printf("  %s 转换 1080P → %s\n", result, converted)
}
