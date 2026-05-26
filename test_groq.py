import urllib.request
import urllib.error
import json

url = "https://api.groq.com/openai/v1/chat/completions"
headers = {
    "Authorization": "Bearer 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",
    "Content-Type": "application/json"
}
data = {
    "model": "llama3-70b-8192",
    "messages": [{"role": "user", "content": "hello"}]
}
req = urllib.request.Request(url, headers=headers, data=json.dumps(data).encode('utf-8'))
try:
    response = urllib.request.urlopen(req)
    print("Success:", response.read().decode('utf-8'))
except urllib.error.HTTPError as e:
    print("HTTP Error:", e.code, e.read().decode('utf-8'))
