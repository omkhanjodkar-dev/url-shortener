import time
import requests

url = "http://localhost:8080"
url = "https://url-shortener-h6yx.onrender.com/"
# payload={"url": "https://www.youtube.com", "code": "a"}

# response = requests.post(f"{url}/shorten", json=payload)

# print(response.status_code, response.text)

# response = requests.get(f"{url}/{response.text}")

# print(response.status_code, response.text)

# payload={"url": "https://www.google.com", "code" : "abc"}

# response = requests.post(f"{url}/shorten", json=payload)

# print(response.status_code, response.text)

# response = requests.get(f"{url}/{response.text}")

# print(response.status_code, response.text)

# response = requests.get(f"{url}/a")

# payload = {"code": "a"}

# response = requests.post(f"{url}/clicks", json=payload)

payload = {"url":"https://www.google.co.in"}
payload = {"url":"https://www.youtube.com"}

for i in range(5):
	response = requests.post(f"{url}/shorten", json=payload)
	print("SHORTEN", response.status_code, response.text)
	# response = requests.get(f"{url}/a")
	# response = requests.get(f"{url}/abc")
	# response = requests.get(f"{url}/bNQadu")
	# print("AJD", response.status_code, response.text)
	response = requests.post(f"{url}/shorten", json=payload)
	# time.sleep(1)
	print("SHORTEN", response.status_code, response.text)
