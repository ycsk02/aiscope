sudo docker run -d --restart always --name redis -e REDIS_PASSWORD=password123 -p 6379:6379 bitnami/redis:latest
