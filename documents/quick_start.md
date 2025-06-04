# Quick Start with Burncloud: A Practical Guide to One-Stop Access to Multiple AI Models

## How Burncloud Works
Burncloud is an integrated platform that aggregates and distributes popular AI models from various brands. Users only need the platform's base_url and a token to access the desired model, and the interface returns data in a unified OpenAI standard format. This allows for quick integration and easy switching between target models, eliminating the tedious processes of applying for accounts, quotas, and binding credit cards for different brands. Excitingly, Burncloud offers discounted prices, at least 20% cheaper than using the official platforms, achieving an 80% discount.

## Information Flow Diagram

<img style="max-width:500px;" src="https://ai.Burncloud.com/uploads/burncloud_request_flow.svg">

## Supported Model List
Currently, mainstream models are supported:
- OpenAI Series: gpt-4.5-preview, gpt-4o-mini, gpt-4.1, o3, gpt-4o
- Gemini Series: gemini-2.0-flash, gemini-1.5-pro, gemini-1.5-flash, gemini-pro-vision
- Claude Series: claude-3.5-sonnet, claude-3.7-sonnet
- Deepseek Series: deepseek-reasoner, deepseek-chat
- Others

The above list includes the mainstream models at the time of writing. For more models, please see the [available model list](https://ai.burncloud.com/price).

## Simplest Example: Hello World
Below is how to call OpenAI's gpt-4o model through Burncloud's API:

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
In the example, there is an "Authorization: Bearer sk-9DufUZISqD4FrlVzvO8nkioqEwxqjPkJ2OLSAibs7d9OKJLe". How can I obtain this token? Check here [Get Token](https://ai.burncloud.com/#Quick_Start_Guide).
## Bonus
We have implemented the request methods for available models on the platform in Postman. You can download and import them into Postman to simulate model requests. Download link: [Postman AI Request Toolkit](https://ai.burncloud.com/uploads/postman_burncloud_ai_api.json).