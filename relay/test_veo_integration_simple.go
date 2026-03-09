package main

import (
	"encoding/json"
	"fmt"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel/task/gemini"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func main() {
	fmt.Println("=========================================")
	fmt.Println("Veo 3.1 集成测试")
	fmt.Println("=========================================")
	fmt.Println()

	testCases := []struct {
		name  string
		req   dto.VideoRequest
	}{
		{
			name: "基础请求 (自动设置generateAudio)",
			req: dto.VideoRequest{
				Model:   "veo-3.1-generate-preview",
				Prompt:  "Test prompt",
				Duration: 8,
			},
		},
		{
			name: "使用n参数 (生成3个视频)",
			req: dto.VideoRequest{
				Model:   "veo-3.1-fast-generate-preview",
				Prompt:  "Test prompt",
				Duration: 6,
				N:       3,
			},
		},
		{
			name: "使用width/height (1080p)",
			req: dto.VideoRequest{
				Model:   "veo-3.1-generate-preview",
				Prompt:  "Test prompt",
				Duration: 8,
				Width:   1920,
				Height:  1080,
			},
		},
		{
			name: "使用metadata.sampleCount",
			req: dto.VideoRequest{
				Model:   "veo-3.1-generate-preview",
				Prompt:  "Test prompt",
				Duration: 8,
				Metadata: map[string]any{
					"sampleCount": 2,
				},
			},
		},
		{
			name: "4K分辨率测试",
			req: dto.VideoRequest{
				Model:   "veo-3.1-generate-preview",
				Prompt:  "Test prompt",
				Duration: 8,
				Metadata: map[string]any{
					"resolution": "4k",
				},
			},
		},
		{
			name: "metadata.generateAudio=false",
			req: dto.VideoRequest{
				Model:   "veo-3.1-generate-preview",
				Prompt:  "Test prompt",
				Duration: 8,
				Metadata: map[string]any{
					"generateAudio": false,
				},
			},
		},
	}

	adaptor := &gemini.TaskAdaptor{}

	for i, tc := range testCases {
		fmt.Printf("\n--- 测试 %d: %s ---\n", i+1, tc.name)

		// 构造TaskSubmitReq
		taskSubmitReq := relaycommon.TaskSubmitReq{
			Model:     tc.req.Model,
			Prompt:     tc.req.Prompt,
			Image:      tc.req.Image,
			Duration:    int(tc.req.Duration),
			Metadata:    tc.req.Metadata,
		}

		if tc.req.N > 0 {
			taskSubmitReq.N = tc.req.N
		}

		// 构造RelayInfo (简化版)
		info := &relaycommon.RelayInfo{
			UpstreamModelName: tc.req.Model,
			PublicTaskID:     fmt.Sprintf("task_%d", i+1),
			OriginModelName:   tc.req.Model,
		}

		adaptor.Init(info)

		// 直接构造VeoRequestPayload来测试逻辑
		params := &gemini.VeoParameters{}

		// 模拟 UnmarshalMetadata
		if taskSubmitReq.Metadata != nil {
			metadataJSON, _ := json.Marshal(taskSubmitReq.Metadata)
			_ = json.Unmarshal(metadataJSON, params)
		}

		// 应用修改后的逻辑
		if params.GenerateAudio == nil {
			defaultAudio := true
			params.GenerateAudio = &defaultAudio
		}

		if params.DurationSeconds == 0 && taskSubmitReq.Duration > 0 {
			params.DurationSeconds = taskSubmitReq.Duration
		}

		// 验证 duration
		if params.DurationSeconds != 0 && params.DurationSeconds != 4 && params.DurationSeconds != 6 && params.DurationSeconds != 8 {
			fmt.Printf("❌ 无效的 duration: %d (必须是4/6/8)\n", params.DurationSeconds)
			continue
		}

		if params.Resolution == "" && tc.req.Width > 0 && tc.req.Height > 0 {
			size := fmt.Sprintf("%dx%d", tc.req.Width, tc.req.Height)
			params.Resolution = gemini.SizeToVeoResolution(size)
			params.AspectRatio = gemini.SizeToVeoAspectRatio(size)
		}

		if params.Resolution == "" && tc.req.Metadata != nil && tc.req.Metadata["resolution"] != nil {
			if res, ok := tc.req.Metadata["resolution"].(string); ok && res != "" {
				params.Resolution = gemini.SizeToVeoResolution(res)
			}
		}

		params.Resolution = ""

		if params.Resolution == "" && tc.req.Width > 0 && tc.req.Height > 0 {
			size := fmt.Sprintf("%dx%d", tc.req.Width, tc.req.Height)
			params.Resolution = gemini.SizeToVeoResolution(size)
			params.AspectRatio = gemini.SizeToVeoAspectRatio(size)
		}
		params.Resolution = gemini.SizeToVeoResolution(params.Resolution)

		// 支持 n 参数
		if taskSubmitReq.N > 0 {
			params.SampleCount = taskSubmitReq.N
		}

		// 验证 sampleCount 范围
		if params.SampleCount == 0 || params.SampleCount < 1 {
			params.SampleCount = 1
		}
		if params.SampleCount > 4 {
			params.SampleCount = 4
		}

		// 构造 VeoRequestPayload
		instance := gemini.VeoInstance{Prompt: taskSubmitReq.Prompt}
		veoReq := gemini.VeoRequestPayload{
			Instances:  []gemini.VeoInstance{instance},
			Parameters: params,
		}

		// 验证生成的请求
		fmt.Printf("✅ 成功构建请求:\n")
		fmt.Printf("  - Model: %s\n", tc.req.Model)
		fmt.Printf("  - Prompt: %s\n", veoReq.Instances[0].Prompt)
		fmt.Printf("  - generateAudio: %v (Veo 3.1必需参数)\n", *veoReq.Parameters.GenerateAudio)
		fmt.Printf("  - sampleCount: %d\n", veoReq.Parameters.SampleCount)
		if veoReq.Parameters.DurationSeconds > 0 {
			fmt.Printf("  - durationSeconds: %d\n", veoReq.Parameters.DurationSeconds)
		}
		if veoReq.Parameters.Resolution != "" {
			fmt.Printf("  - resolution: %s\n", veoReq.Parameters.Resolution)
		}
		if veoReq.Parameters.AspectRatio != "" {
			fmt.Printf("  - aspectRatio: %s\n", veoReq.Parameters.AspectRatio)
		}

		// 构建最终的JSON
		jsonData, _ := json.MarshalIndent(veoReq, "", "  ")
		fmt.Printf("\n最终请求JSON:\n%s\n", string(jsonData))
	}

	fmt.Println("\n=========================================")
	fmt.Println("集成测试完成！")
	fmt.Println("=========================================")
}
