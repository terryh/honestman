# Honestman
爬加樂富，打潤髮商品及價格

# Dependence
PostgreSQL

# DEPLOY
sudo docker run -d --restart=always -e DBHOST=web -p 80:3000 -p 443:443 -v /usr/src/app/api:/usr/src/app --name api goapp

sudo docker run -d --restart=always -e DBHOST=web  -v /usr/src/app/crawler:/usr/src/app --name crawler goapp

# Maybe
1. Index and search (elastic).
2. Better user interface.
3. If need test (benchmark) case ?

