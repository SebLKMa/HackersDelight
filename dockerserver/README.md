1. `docker build --tag cv-targetserver:latest .`
2. `docker run -dp 9000:80 --name cv-targetserver-container cv-targetserver`
3. Browse to `http://localhost:9000/landing.html`