import requests

url = "https://leaderfactor.app/recruiting"
payload = {"email": "lehigraciaiii@gmail.com"}

response = requests.post(url, json=payload)

print(response.status_code)
print(response.text)
