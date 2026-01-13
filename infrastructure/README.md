# Infrastructure as Code для quiz-sprint TMA

Эта директория содержит всю инфраструктуру для развертывания TMA приложения на любом VPS сервере.

## Структура

```
infrastructure/
├── nginx/
│   ├── sites-available/
│   │   ├── dev.conf           # Dev окружение (через SSH tunnel)
│   │   ├── staging.conf       # Staging окружение
│   │   └── production.conf    # Production окружение
│   └── ssl-params.conf        # Общие SSL параметры
├── scripts/
│   ├── setup-vps.sh          # Первичная настройка VPS
│   ├── setup-ssl.sh          # Настройка SSL сертификатов
│   └── deploy-nginx.sh       # Деплой Nginx конфигурации
└── README.md                 # Эта документация
```

## Быстрый старт

### Настройка нового VPS с нуля

```bash
# 1. Подключиться к VPS
ssh root@your-vps-ip

# 2. Склонировать репозиторий
git clone https://github.com/your-username/quiz-sprint.git
cd quiz-sprint/infrastructure/scripts

# 3. Запустить setup
./setup-vps.sh

# 4. Настроить SSL
./setup-ssl.sh

# 5. Готово! Теперь можно деплоить приложение
```

## Детальная документация

### 1. Первичная настройка VPS (setup-vps.sh)

Этот скрипт выполняет:
- ✅ Обновление системных пакетов
- ✅ Установку Nginx
- ✅ Установку Certbot для Let's Encrypt
- ✅ Создание директорий для статики (`/var/www/tma/staging`, `/var/www/tma/production`)
- ✅ Копирование Nginx конфигураций
- ✅ Включение сайтов
- ✅ Настройку firewall

**Требования:**
- Ubuntu 20.04+ или Debian 10+
- Права root
- Минимум 512MB RAM

**Использование:**
```bash
cd infrastructure/scripts
sudo ./setup-vps.sh
```

### 2. Настройка SSL сертификатов (setup-ssl.sh)

Автоматически получает SSL сертификаты от Let's Encrypt для ваших доменов.

**Требования:**
- DNS записи должны указывать на ваш VPS
- Порты 80 и 443 открыты
- Nginx запущен

**Проверка DNS перед запуском:**
```bash
host staging.quiz-sprint-tma.online
host quiz-sprint-tma.online
```

**Использование:**
```bash
cd infrastructure/scripts
sudo ./setup-ssl.sh
```

Скрипт:
- Проверит DNS записи
- Получит SSL сертификаты для staging
- Опционально получит сертификаты для production
- Настроит auto-renewal

### 3. Деплой обновленной конфигурации (deploy-nginx.sh)

Используется для обновления Nginx конфигурации на уже настроенном VPS.

**Использование:**
```bash
cd infrastructure/scripts

# С дефолтными настройками (144.31.199.226)
./deploy-nginx.sh

# С кастомным VPS
VPS_HOST=your-vps-ip VPS_USER=root ./deploy-nginx.sh
```

Скрипт:
- Копирует конфигурации на VPS
- Тестирует Nginx конфигурацию
- Перезагружает Nginx
- Показывает статус

## Архитектура окружений

### Development
- **URL:** https://dev.quiz-sprint-tma.online
- **Источник:** Локальный Vite dev server (через SSH tunnel)
- **Порт:** localhost:5173
- **HMR:** Поддерживается

### Staging
- **URL:** https://staging.quiz-sprint-tma.online
- **Источник:** Статические файлы
- **Путь:** `/var/www/tma/staging`
- **Деплой:** GitHub Actions (ручной запуск)

### Production
- **URL:** https://quiz-sprint-tma.online
- **Источник:** Статические файлы
- **Путь:** `/var/www/tma/production`
- **Деплой:** GitHub Actions (ручной запуск)

## Конфигурация Nginx

### Общие фичи всех окружений:
- ✅ HTTPS/SSL с Let's Encrypt
- ✅ Auto-renewal сертификатов
- ✅ HTTP → HTTPS редирект
- ✅ Gzip сжатие
- ✅ SPA routing (Vue Router поддержка)
- ✅ Кэширование статики (1 год)
- ✅ Security headers

### SSL параметры (ssl-params.conf)
- TLS 1.2 и 1.3
- Современные cipher suites
- OCSP Stapling
- HSTS (31536000 секунд)

## Обслуживание

### Проверка статуса Nginx
```bash
ssh root@your-vps "systemctl status nginx"
```

### Просмотр логов
```bash
# Staging access logs
ssh root@your-vps "tail -f /var/log/nginx/staging-tma-access.log"

# Staging error logs
ssh root@your-vps "tail -f /var/log/nginx/staging-tma-error.log"

# Production logs
ssh root@your-vps "tail -f /var/log/nginx/production-tma-access.log"
ssh root@your-vps "tail -f /var/log/nginx/production-tma-error.log"
```

### Тест конфигурации
```bash
ssh root@your-vps "nginx -t"
```

### Перезагрузка Nginx
```bash
ssh root@your-vps "systemctl reload nginx"
```

### Проверка SSL сертификатов
```bash
ssh root@your-vps "certbot certificates"
```

### Ручное обновление SSL
```bash
ssh root@your-vps "certbot renew"
```

## Миграция на новый сервер

1. **Подготовка DNS:**
   ```bash
   # Обновить A-записи на новый IP
   staging.quiz-sprint-tma.online → new-vps-ip
   quiz-sprint-tma.online → new-vps-ip
   ```

2. **Запуск setup на новом VPS:**
   ```bash
   ssh root@new-vps-ip
   git clone <repo>
   cd quiz-sprint/infrastructure/scripts
   ./setup-vps.sh
   ./setup-ssl.sh
   ```

3. **Деплой приложения:**
   - Запустить GitHub Actions workflow
   - Или скопировать файлы вручную:
     ```bash
     scp -r dist/* root@new-vps:/var/www/tma/staging/
     ```

4. **Проверка:**
   ```bash
   curl -I https://staging.quiz-sprint-tma.online
   ```

## Troubleshooting

### Ошибка: "nginx: configuration file test failed"
```bash
# Проверить синтаксис
nginx -t

# Проверить существование SSL сертификатов
ls -la /etc/letsencrypt/live/

# Если сертификатов нет - запустить setup-ssl.sh
```

### Ошибка: "Connection refused"
```bash
# Проверить что Nginx запущен
systemctl status nginx

# Запустить если не запущен
systemctl start nginx
```

### Ошибка: "502 Bad Gateway" (для dev окружения)
```bash
# Проверить SSH tunnel
ps aux | grep ssh

# Проверить что Vite dev server запущен локально
lsof -i :5173
```

### SSL сертификат не получается
```bash
# Проверить DNS
host staging.quiz-sprint-tma.online

# Проверить firewall
ufw status

# Проверить порты
netstat -tuln | grep -E ':(80|443)'
```

## Безопасность

### Firewall
Убедитесь что открыты только необходимые порты:
```bash
ufw status
```

Должно быть:
- 22 (SSH)
- 80 (HTTP)
- 443 (HTTPS)

### SSH ключи
Используйте SSH ключи вместо паролей:
```bash
# Генерация ключа
ssh-keygen -t ed25519 -C "your_email@example.com"

# Копирование на VPS
ssh-copy-id root@your-vps
```

### Обновления безопасности
```bash
# Регулярно обновляйте систему
apt update && apt upgrade -y
```

## GitHub Actions Integration

### Secrets необходимые для CI/CD:
- `VPS_SSH_KEY` - SSH приватный ключ
- `VPS_HOST` - IP адрес VPS
- `VPS_USER` - SSH пользователь (обычно root)
- `TELEGRAM_BOT_TOKEN` - для уведомлений
- `TELEGRAM_CHAT_ID` - для уведомлений

### Workflow файлы:
- `.github/workflows/deploy-staging.yml` - деплой на staging
- `.github/workflows/deploy-production.yml` - деплой на production (будущее)

## Дальнейшее развитие

### Планируется добавить:
- [ ] Ansible playbooks для полной автоматизации
- [ ] Docker Compose для backend сервисов
- [ ] Prometheus + Grafana для мониторинга
- [ ] Автоматические бэкапы
- [ ] Blue-Green deployment
- [ ] Rate limiting
- [ ] WAF (Web Application Firewall)

## Контакты

Вопросы по инфраструктуре: barsukov.d@gmail.com

## Лицензия

Приватный проект quiz-sprint TMA
