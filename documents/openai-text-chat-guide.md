# OpenAI Text Chat Models Guide - GPT-4o & Others

## Introduction
This guide shows how to call OpenAI's gpt-4o model through Burncloud's API and provides an overview of OpenAI's text chat models, including gpt-4o and other models. It covers the supported model list, a simple example of calling the gpt-4o model, and a detailed example of using the gpt-4o model for a chatbot application. It also includes information on the content filtering feature and how to use it in your applications.
 
## Supported Model List
- gpt-4.5-preview
- gpt-4o-mini
- gpt-4.1
- o3
- gpt-4o
  
## Simplest Example: Hello World


### Request
<pre class="markdown-custom-code">
curl https://ai.burncloud.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-9DufUZISqD4FrlVzvO8nkioqEwxqjPkJ2OLSAibs7d9OKJLe" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Hello world!"
      }
    ]
  }'
</pre>

### Response
<pre class="markdown-custom-code">
{
    "choices": [
        {
            "content_filter_results": {
                "hate": {
                    "filtered": false,
                    "severity": "safe"
                },
                "protected_material_code": {
                    "filtered": false,
                    "detected": false
                },
                "protected_material_text": {
                    "filtered": false,
                    "detected": false
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
                "content": "Hello! How can I assist you today?",
                "refusal": null,
                "role": "assistant"
            }
        }
    ],
    "created": 1745504517,
    "id": "chatcmpl-BPrkbzVMfZNmwawaKWMjULUyUDFJb",
    "model": "gpt-4o-2024-05-13",
    "object": "chat.completion",
    "prompt_filter_results": [
        {
            "prompt_index": 0,
            "content_filter_results": {
                "hate": {
                    "filtered": false,
                    "severity": "safe"
                },
                "jailbreak": {
                    "filtered": false,
                    "detected": false
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
            }
        }
    ],
    "system_fingerprint": "fp_ee1d74bde0",
    "usage": {
        "completion_tokens": 10,
        "completion_tokens_details": {
            "accepted_prediction_tokens": 0,
            "audio_tokens": 0,
            "reasoning_tokens": 0,
            "rejected_prediction_tokens": 0
        },
        "prompt_tokens": 20,
        "prompt_tokens_details": {
            "audio_tokens": 0,
            "cached_tokens": 0
        },
        "total_tokens": 30
    }
}
</pre>
