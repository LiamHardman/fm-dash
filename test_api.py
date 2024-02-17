import http.client

conn = http.client.HTTPSConnection("api-football-v1.p.rapidapi.com")

headers = {
    'X-RapidAPI-Key': "836ab3868amsh6083aa8904482c5p191ba0jsn2895d0424b33",
    'X-RapidAPI-Host': "api-football-v1.p.rapidapi.com"
}

conn.request("GET", "/v3/teams?season=2023", headers=headers)

res = conn.getresponse()
data = res.read()

print(data.decode("utf-8"))