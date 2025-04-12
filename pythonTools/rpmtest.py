import http.client
import json

conn = http.client.HTTPSConnection("api.deerapi.com")
payload = json.dumps({
   "model": "claude-3-5-sonnet-20250219",
   "messages": [
      {
         "role": "user",
         "content": "Hello!"
      }
   ],
   "stream": False
})
headers = {
   'Authorization': 'Bearer sk-ynzKdDQifKpgDvOnjX8ZHxhTsPt18FrPiu6OilXBFOJuh9E7',
   'Content-Type': 'application/json'
}
conn.request("POST", "/v1/chat/completions", payload, headers)
res = conn.getresponse()
data = res.read()
print(data.decode("utf-8"))