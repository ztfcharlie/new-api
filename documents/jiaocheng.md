# BurnCloud AI API Key Acquisition and Deployment Tutorial

## 1. Introduction
This tutorial will guide you through the process of acquiring an API key and deploying it for use with the ai.burncloud.com API.

## 2. Create an Account
1. Navigate to the BurnCloud AI API website (ai.burncloud.com).
2. Locate the "Create an account" option on the page.
3. Fill in the required registration information, including but not limited to username, email, and password.
4. Follow the verification steps (usually via email) to complete the account creation process.

## 3. Sign In
1. After successfully creating an account, click on the "Sign in" option.
2. Enter your registered username/email and password to log in to your account.

## 4. Create a New Token (API Key)
1. Once logged in, find the section related to API management. It may be labeled as "API Tokens" or something similar.
2. Click on the "Create New Token" button.
3. In the "Create New Token" form:
    - Give your token a descriptive name (e.g., "my - api - key - for - project - X") in the provided field.
    - Optionally, you can set an expiration date for the token if needed.
    - Review any additional settings or disclaimers presented.
4. Click the "Create token" button to generate your API key.

## 5. Locate and Save Your API Key
1. After creating the token, you will be presented with your API key. It will be displayed in a section like the one labeled "Key" in the provided flowchart.
2. **IMPORTANT**: Copy and securely store this API key. Treat it as a sensitive piece of information, similar to a password, as it gives access to the BurnCloud AI API on behalf of your account.

## 6. API Deployment - Example Usage
The base URL (endpoint) for the BurnCloud AI API is: https://ai.burncloud.com/v1/chat/completions

Here is an example of how to use your API key in an API call (using Python with the `requests` library as an example):

```python
import requests

api_key = "YOUR_API_KEY_HERE"
headers = {
    "Content - type": "application/json",
    "Authorization": f"Bearer {api_key}"
}
data = {
    "model": "default",
    "messages": [
        {"role": "system", "content": "You are an assistant"},
        {"role": "user", "content": "Hello, how are you today?"}
    ]
}

response = requests.post(
    "https://ai.burncloud.com/v1/chat/completions",
    headers = headers,
    json = data
)

print(response.json())
```

Replace `"YOUR_API_KEY_HERE"` with the actual API key you obtained in the previous step.

## 7. Tips and Precautions
- Keep your API key confidential at all times. Do not share it publicly or in insecure channels.
- Be aware of the usage limits and costs associated with the API. Check the BurnCloud AI API documentation for details on pricing and rate limits.
- If you encounter any issues during the API key acquisition or deployment process, refer to the official BurnCloud AI API documentation or reach out to their support team. 