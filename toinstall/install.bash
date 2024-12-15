#!/bin/bash

set -e  # Завершить выполнение при ошибке

# Функция для проверки выполнения шагов
echo_step() {
    echo -e "\n\033[1;34m==> $1...\033[0m\n"
}

# 1. Установка WireGuard
echo_step "Установка WireGuard"
if ! command -v wg > /dev/null; then
    apt-get update
    apt-get install -y wireguard
    echo "WireGuard установлен."
else
    echo "WireGuard уже установлен."
fi

# 2. Установка PostgreSQL
echo_step "Установка PostgreSQL"
if ! command -v psql > /dev/null; then
    apt-get install -y postgresql postgresql-contrib
    echo "PostgreSQL установлен."
else
    echo "PostgreSQL уже установлен."
fi

# 3. Запуск службы PostgreSQL
echo_step "Запуск службы PostgreSQL"
systemctl enable postgresql
systemctl start postgresql

# 4. Загрузка базы данных из дампа
DB_DUMP="vpn.sql"
echo_step "Загрузка базы данных из дампа $DB_DUMP"
DB_NAME="vpn"
DB_USER="vpn"
DB_PASSWORD="vpn"

# Создание пользователя и базы данных
sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" || echo "Пользователь $DB_USER уже существует."
sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" || echo "База данных $DB_NAME уже существует."

# Импорт дампа
sudo +x migrate

sudo migrate

sudo +x main
# Уведомление об успешной установке
echo -e "\n\033[1;32mУстановка завершена успешно. WireGuard и PostgreSQL настроены. Дамп базы данных загружен.\033[0m\n"
