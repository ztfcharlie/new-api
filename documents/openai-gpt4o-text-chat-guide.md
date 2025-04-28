# OpenAI GPT-4o Text Chat Models Guide

## Introduction
This guide shows how to call OpenAI's gpt-4o model through Burncloud's API and provides an overview of OpenAI's text chat models, including gpt-4o and other models. It covers the supported model list, a simple example of calling the gpt-4o model, and a detailed example of using the gpt-4o model for a chatbot application.
 
## Supported Model List
- gpt-4.5-preview
- gpt-4o-mini
- gpt-4.1
- o3
- gpt-4o
  
## Simplest Example: Hello World

### Complete Response Mode / One-shot Response

#### Request
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

#### Response
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

### Stream Generation Mode / Streaming Response Mode
#### Request
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
        "content": "Hello world !"
      }
    ],
    "stream": true
  }'
</pre>
#### Response
<pre class="markdown-custom-code">
data: {"choices":[],"created":0,"id":"","model":"","object":"","prompt_filter_results":[{"prompt_index":0,"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"jailbreak":{"filtered":false,"detected":false},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}}}]}

data: {"choices":[{"content_filter_results":{},"delta":{"content":"","refusal":null,"role":"assistant"},"finish_reason":null,"index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: {"choices":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"Hello"},"finish_reason":null,"index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: {"choices":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"!"},"finish_reason":null,"index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: {"choices":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":" How"},"finish_reason":null,"index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: {"choices":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":" can"},"finish_reason":null,"index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: {"choices":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":" I"},"finish_reason":null,"index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: {"choices":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":" assist"},"finish_reason":null,"index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: {"choices":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":" you"},"finish_reason":null,"index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: {"choices":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":" today"},"finish_reason":null,"index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: {"choices":[{"content_filter_results":{"hate":{"filtered":false,"severity":"safe"},"self_harm":{"filtered":false,"severity":"safe"},"sexual":{"filtered":false,"severity":"safe"},"violence":{"filtered":false,"severity":"safe"}},"delta":{"content":"?"},"finish_reason":null,"index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: {"choices":[{"content_filter_results":{},"delta":{},"finish_reason":"stop","index":0,"logprobs":null}],"created":1745817306,"id":"chatcmpl-BRB7afx5zbyUQlHBAhT55hDVk9aUp","model":"gpt-4o-2024-08-06","object":"chat.completion.chunk","system_fingerprint":"fp_ee1d74bde0","usage":null}

data: [DONE]


</pre>
