import os
from openai import AzureOpenAI

#endpoint = "https://aron-m6yjpnxl-swedencentral.cognitiveservices.azure.com/"
endpoint = "https://aron-m6yjpnxl-swedencentral.openai.azure.com/"
model_name = "gpt-4.5-preview"
deployment = "gpt-4.5-preview"

#subscription_key = "AtCETkVdqH6FxZjQdN9P53ZSWCUNcfiuvuCTMUyKlbii0e6mYbfTJQQJ99BBACfhMk5XJ3w3AAAAACOGpfHx"
subscription_key = "AtCETkVdqH6FxZjQdN9P53ZSWCUNcfiuvuCTMUyKlbii0e6mYbfTJQQJ99BBACfhMk5XJ3w3AAAAACOGpfHx"
api_version = "2024-12-01-preview"

client = AzureOpenAI(
    api_version=api_version,
    azure_endpoint=endpoint,
    api_key=subscription_key,
)

response = client.chat.completions.create(
    messages=[
        {
            "role": "system",
            "content": "You are a helpful assistant.",
        },
        {
            "role": "user",
            "content": "I am going to Paris, what should I see?",
        }
    ],
    max_tokens=800,
    temperature=1.0,
    top_p=1.0,
    frequency_penalty=0.0,
    presence_penalty=0.0,
    model=deployment
)

print(response.choices[0].message.content)