import requests

url = "http://localhost:8080"
payload={"url": "https://www.youtube.com"}

response = requests.post(f"{url}/shorten", json=payload)

print(response.status_code, response.text)

# response = requests.get(f"{url}/{response.text}")

# print(response.status_code, response.text)

payload={"url": "https://www.google.com"}

response = requests.post(f"{url}/shorten", json=payload)

print(response.status_code, response.text)

# response = requests.get(f"{url}/{response.text}")

# print(response.status_code, response.text)