# MZHN CHATS SERVICE

## DEPLOY

### Docker

1. Склонируйте репозиторий `git clone https://github.com/mzhn-mzhnr/chats.git`
2. Настройте конфиг приложения. Для этого выполните команду

```bash
cp example.env .env
```

Настройте конфиг `.env` в вашем текстовом редакторе

3. Для запуска приложения в docker-контейнере используйте

```bash
docker compose up --build
```
