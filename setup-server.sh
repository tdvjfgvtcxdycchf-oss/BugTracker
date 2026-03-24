#!/bin/bash
# Первоначальная настройка сервера. Запускать один раз под root.
set -e

DOMAIN="bugtracker.sytes.net"
APP_DIR="/opt/bugtracker"

echo "=== Клонируем репозиторий ==="
mkdir -p $APP_DIR
git clone https://github.com/tdvjfgvtcxdycchf-oss/BugTracker.git $APP_DIR
# Если уже склонирован: cd $APP_DIR && git pull

echo "=== Создаём .env для бэкенда ==="
cat > $APP_DIR/TaskTrackerBackend/docker/.env << 'EOF'
POSTGRES_DB=postgres
POSTGRES_USER=postgres
POSTGRES_PASSWORD=ЗАМЕНИТЕ_НА_СВОЙ_ПАРОЛЬ
POSTGRES_URL=postgres://postgres:ЗАМЕНИТЕ_НА_СВОЙ_ПАРОЛЬ@postgres_db:5432/postgres?sslmode=disable
EOF
echo "ВАЖНО: отредактируйте $APP_DIR/TaskTrackerBackend/docker/.env и замените пароль!"

echo "=== Копируем конфиг Nginx ==="
cp $APP_DIR/nginx/bugtracker.conf /etc/nginx/sites-available/bugtracker
ln -sf /etc/nginx/sites-available/bugtracker /etc/nginx/sites-enabled/bugtracker
nginx -t && systemctl reload nginx

echo "=== Получаем SSL сертификат ==="
certbot --nginx -d $DOMAIN --non-interactive --agree-tos -m admin@$DOMAIN

echo "=== Первый деплой ==="
bash $APP_DIR/deploy.sh

echo "=== Готово! https://$DOMAIN ==="
