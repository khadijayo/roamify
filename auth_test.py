import requests, json, time
base = 'https://roamify-f9v5.onrender.com'
user_email = f'test_{int(time.time())}@example.com'
register = requests.post(base + '/api/v1/auth/register', json={
    'full_name': 'Test User',
    'email': user_email,
    'password': 'password123'
}, timeout=15)
print('REGISTER', register.status_code, register.text)
login = requests.post(base + '/api/v1/auth/login', json={
    'email': user_email,
    'password': 'password123'
}, timeout=15)
print('LOGIN', login.status_code, login.text)
if login.status_code == 200:
    token = login.json().get('data', {}).get('token') if isinstance(login.json(), dict) else None
    if not token:
        print('no token found; check login output')
    else:
        headers = {'Authorization': 'Bearer ' + token}
        for method,path in [('GET','/api/v1/users/me'), ('GET','/api/v1/users/me/vibe'), ('GET','/api/v1/trips/'),('GET','/api/v1/posts/')]:
            r = requests.request(method, base + path, headers=headers, timeout=15)
            print(method, path, r.status_code, r.text[:200])
else:
    print('login failed; skip authenticated checks')
