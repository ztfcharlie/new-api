# OpenAI GPT-4o Image Chat Models Guide

## Introduction
This guide shows how to call OpenAI's gpt-4o model through Burncloud's API and provides an overview of OpenAI's text chat models, including gpt-4o and other models. It covers the supported model list, a simple example of calling the gpt-4o model, and a detailed example of using the gpt-4o model for a chatbot application.
 
## Supported Model List
- gpt-4.5-preview
- gpt-4o-mini
- gpt-4.1
- o3
- gpt-4o
  
## Simplest Example: What is in this picture?

### Complete Response Mode / One-shot Response

#### Request
```bash
curl https://ai.burncloud.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-9DufUZISqD4FrlVzvO8nkioqEwxqjPkJ2OLSAibs7d9OKJLe" \
  -d '{
        "model": "gpt-4o",
        "messages": [
            {
                "role": "user",
                "content": [
                    {
                        "type": "text",
                        "text": "这张图片里有什么?"
                    },
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": "http://devopenai.burncloud.com/uploads/cat.jpg"
                        }
                    }
                ]
            }
        ]
    }'
```

#### Response
```json
{
    "choices": [
        {
            "content_filter_results": {
                "hate": {
                    "filtered": false,
                    "severity": "safe"
                },
                "self_harm": {
                    "filtered": false,
                    "severity": "safe"
                },
                "sexual": {
                    "filtered": false,
                    "severity": "safe"
                },
                "violence": {
                    "filtered": false,
                    "severity": "safe"
                }
            },
            "finish_reason": "stop",
            "index": 0,
            "logprobs": null,
            "message": {
                "content": "这张图片里有两只猫。靠近镜头的是一只橙色的猫，正张嘴大叫或哈欠，戴着一条粉色带心形图案的围巾。背景中的猫是灰色的虎斑猫，舌头伸出来，看上去好像在舔嘴巴，戴着一条格子花纹和白色荷叶边的围巾。后面的背景是一些绿色植物。",
                "refusal": null,
                "role": "assistant"
            }
        }
    ],
    "created": 1745827762,
    "id": "chatcmpl-BRDqEB8LYFTuTYCK4CL5WgLzxdmsN",
    "model": "gpt-4o-2024-08-06",
    "object": "chat.completion",
    "prompt_filter_results": [
        {
            "prompt_index": 0,
            "content_filter_result": {
                "jailbreak": {
                    "filtered": false,
                    "detected": false
                },
                "custom_blocklists": {
                    "filtered": false,
                    "details": []
                }
            }
        },
        {
            "prompt_index": 1,
            "content_filter_result": {
                "sexual": {
                    "filtered": false,
                    "severity": "safe"
                },
                "violence": {
                    "filtered": false,
                    "severity": "safe"
                },
                "hate": {
                    "filtered": false,
                    "severity": "safe"
                },
                "self_harm": {
                    "filtered": false,
                    "severity": "safe"
                },
                "custom_blocklists": {
                    "filtered": false,
                    "details": []
                }
            }
        }
    ],
    "system_fingerprint": "fp_ee1d74bde0",
    "usage": {
        "completion_tokens": 101,
        "completion_tokens_details": {
            "accepted_prediction_tokens": 0,
            "audio_tokens": 0,
            "reasoning_tokens": 0,
            "rejected_prediction_tokens": 0
        },
        "prompt_tokens": 268,
        "prompt_tokens_details": {
            "audio_tokens": 0,
            "cached_tokens": 0
        },
        "total_tokens": 369
    }
}
```
