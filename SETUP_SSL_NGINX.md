# SSL & Nginx Setup Guide for Multiple Sites

⚡ **Quick Reference**
```bash
# Install requirements
sudo apt update && sudo apt install -y nginx certbot python3-certbot-nginx

# Generate SSL certificate
sudo certbot --nginx -d yourdomain.com

# Test nginx config
sudo nginx -t && sudo systemctl reload nginx
```

## Overview

This guide shows how to set up nginx as a reverse proxy with SSL certificates for multiple personal projects on a single VPS. Perfect for hosting multiple applications with proper SSL termination.

## Prerequisites

- Ubuntu/Debian VPS with root access
- Domain names pointing to your VPS IP
- Applications running on different ports (e.g., 3000, 3001, 3002)

## Step 1: Install Required Software

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install nginx
sudo apt install -y nginx

# Install certbot for Let's Encrypt SSL certificates
sudo apt install -y certbot python3-certbot-nginx

# Start and enable nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

## Step 2: Basic Nginx Configuration Structure

Create the main nginx configuration that supports multiple sites:

```bash
# Backup original config
sudo cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup

# Create main configuration
sudo tee /etc/nginx/nginx.conf > /dev/null << 'NGINX_CONF'
user www-data;
worker_processes auto;
pid /run/nginx.pid;
include /etc/nginx/modules-enabled/*.conf;

events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    # Basic settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    server_tokens off;

    # MIME types
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # SSL settings (global)
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:50m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;

    # Security headers (global)
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    # Include site configurations
    include /etc/nginx/sites-enabled/*;
}
NGINX_CONF
```

## Step 3: Create Site-Specific Configurations

Create individual configuration files for each project:

### Site 1: Go Auth Application (Port 3000)

```bash
sudo tee /etc/nginx/sites-available/go-auth.aldas.dev > /dev/null << 'SITE1_CONF'
# Upstream for Go auth app
upstream go_auth_app {
    server 127.0.0.1:3000;
    keepalive 32;
}

# HTTP redirect to HTTPS
server {
    listen 80;
    server_name go-auth.aldas.dev;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name go-auth.aldas.dev;

    # SSL certificates (will be configured by certbot)
    ssl_certificate /etc/letsencrypt/live/go-auth.aldas.dev/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/go-auth.aldas.dev/privkey.pem;

    # Proxy to Go application
    location / {
        proxy_pass http://go_auth_app;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
        proxy_cache_bypass $http_upgrade;
        proxy_redirect off;
    }

    # Security headers specific to this site
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';" always;
}
SITE1_CONF
```

### Site 2: Example Project (Port 3001)

```bash
sudo tee /etc/nginx/sites-available/project2.aldas.dev > /dev/null << 'SITE2_CONF'
# Upstream for project 2
upstream project2_app {
    server 127.0.0.1:3001;
    keepalive 32;
}

# HTTP redirect to HTTPS
server {
    listen 80;
    server_name project2.aldas.dev;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name project2.aldas.dev;

    # SSL certificates
    ssl_certificate /etc/letsencrypt/live/project2.aldas.dev/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/project2.aldas.dev/privkey.pem;

    # Proxy to application
    location / {
        proxy_pass http://project2_app;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_redirect off;
    }

    # Static files (if needed)
    location /static/ {
        alias /var/www/project2/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
SITE2_CONF
```

### Site 3: Static Website

```bash
sudo tee /etc/nginx/sites-available/portfolio.aldas.dev > /dev/null << 'SITE3_CONF'
# HTTP redirect to HTTPS
server {
    listen 80;
    server_name portfolio.aldas.dev;
    return 301 https://$server_name$request_uri;
}

# HTTPS server for static site
server {
    listen 443 ssl http2;
    server_name portfolio.aldas.dev;

    # SSL certificates
    ssl_certificate /etc/letsencrypt/live/portfolio.aldas.dev/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/portfolio.aldas.dev/privkey.pem;

    # Document root
    root /var/www/portfolio;
    index index.html index.htm;

    # Static file serving
    location / {
        try_files $uri $uri/ =404;
    }

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|svg|ico|woff|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
SITE3_CONF
```

## Step 4: Enable Sites

```bash
# Enable the sites you want to use
sudo ln -s /etc/nginx/sites-available/go-auth.aldas.dev /etc/nginx/sites-enabled/
sudo ln -s /etc/nginx/sites-available/project2.aldas.dev /etc/nginx/sites-enabled/
sudo ln -s /etc/nginx/sites-available/portfolio.aldas.dev /etc/nginx/sites-enabled/

# Remove default site if it exists
sudo rm -f /etc/nginx/sites-enabled/default

# Test configuration
sudo nginx -t
```

## Step 5: Generate SSL Certificates

### For each domain, generate SSL certificates using Let's Encrypt:

```bash
# For go-auth.aldas.dev
sudo certbot --nginx -d go-auth.aldas.dev

# For additional domains
sudo certbot --nginx -d project2.aldas.dev
sudo certbot --nginx -d portfolio.aldas.dev

# Or generate all at once
sudo certbot --nginx -d go-auth.aldas.dev -d project2.aldas.dev -d portfolio.aldas.dev
```

### Manual certificate generation (if certbot nginx plugin doesn't work):

```bash
# Generate certificate only
sudo certbot certonly --standalone -d yourdomain.com

# Then manually update nginx config with certificate paths:
# ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
# ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
```

## Step 6: Setup Auto-Renewal

```bash
# Test auto-renewal
sudo certbot renew --dry-run

# Add cron job for auto-renewal (optional, certbot usually sets this up)
echo "0 12 * * * /usr/bin/certbot renew --quiet" | sudo crontab -
```

## Step 7: Start Your Applications

Make sure your applications are running on their respective ports:

```bash
# Example: Start Go auth app on port 3000
cd /root/auth
go run main.go &

# Example: Start another project on port 3001
cd /path/to/project2
npm start &  # or whatever command starts your app

# Or use systemd services for production
```

## Step 8: Reload Nginx

```bash
# Test configuration
sudo nginx -t

# Reload nginx to apply changes
sudo systemctl reload nginx

# Check status
sudo systemctl status nginx
```

## 🔧 Troubleshooting Checklist

### Common Issues:

1. **Certificate generation fails:**
   ```bash
   # Check if port 80 is accessible
   sudo ufw allow 80
   sudo ufw allow 443
   
   # Stop nginx temporarily for standalone mode
   sudo systemctl stop nginx
   sudo certbot certonly --standalone -d yourdomain.com
   sudo systemctl start nginx
   ```

2. **Nginx fails to start:**
   ```bash
   # Check syntax
   sudo nginx -t
   
   # Check error logs
   sudo tail -f /var/log/nginx/error.log
   
   # Check if ports are in use
   sudo netstat -tlnp | grep :80
   sudo netstat -tlnp | grep :443
   ```

3. **Application not accessible:**
   ```bash
   # Check if app is running on correct port
   sudo netstat -tlnp | grep :3000
   
   # Test direct connection to app
   curl http://localhost:3000
   
   # Check nginx access logs
   sudo tail -f /var/log/nginx/access.log
   ```

4. **SSL certificate issues:**
   ```bash
   # Check certificate status
   sudo certbot certificates
   
   # Force renewal if needed
   sudo certbot renew --force-renewal -d yourdomain.com
   
   # Check certificate details
   openssl x509 -in /etc/letsencrypt/live/yourdomain.com/fullchain.pem -text -noout
   ```

## 🔒 Security Considerations

### 1. Firewall Configuration
```bash
# Enable UFW firewall
sudo ufw enable

# Allow essential ports
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Block direct access to application ports
sudo ufw deny 3000
sudo ufw deny 3001
```

### 2. Nginx Security Headers
Add these to your server blocks for enhanced security:

```nginx
# Security headers
add_header X-Frame-Options DENY;
add_header X-Content-Type-Options nosniff;
add_header X-XSS-Protection "1; mode=block";
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header Referrer-Policy "strict-origin-when-cross-origin";
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';";

# Hide nginx version
server_tokens off;
```

### 3. Rate Limiting
```nginx
# Add to http block
limit_req_zone $binary_remote_addr zone=login:10m rate=5r/m;

# Add to location blocks that need protection
location /login {
    limit_req zone=login burst=5 nodelay;
    proxy_pass http://go_auth_app;
    # ... other proxy settings
}
```

## 📋 Complete Setup Script

Here's a one-liner script to set up everything:

```bash
#!/bin/bash
# Quick setup script for nginx + SSL

# Variables - CHANGE THESE
DOMAIN1="go-auth.aldas.dev"
DOMAIN2="project2.aldas.dev"
DOMAIN3="portfolio.aldas.dev"
APP1_PORT="3000"
APP2_PORT="3001"

# Install dependencies
sudo apt update && sudo apt install -y nginx certbot python3-certbot-nginx

# Create site configs (you'll need to customize these)
echo "Creating nginx configurations..."
echo "Please customize the domain names and ports in the generated configs!"

# Generate SSL certificates
sudo certbot --nginx -d $DOMAIN1 -d $DOMAIN2 -d $DOMAIN3

# Test and reload
sudo nginx -t && sudo systemctl reload nginx

echo "Setup complete! Don't forget to:"
echo "1. Update domain names in configs"
echo "2. Start your applications on the correct ports"
echo "3. Configure your firewall"
```

## 🌐 Adding New Sites

To add a new site:

1. **Create new site config:**
   ```bash
   sudo nano /etc/nginx/sites-available/newsite.aldas.dev
   ```

2. **Use this template:**
   ```nginx
   upstream newsite_app {
       server 127.0.0.1:PORT_NUMBER;
   }

   server {
       listen 80;
       server_name newsite.aldas.dev;
       return 301 https://$server_name$request_uri;
   }

   server {
       listen 443 ssl http2;
       server_name newsite.aldas.dev;

       ssl_certificate /etc/letsencrypt/live/newsite.aldas.dev/fullchain.pem;
       ssl_certificate_key /etc/letsencrypt/live/newsite.aldas.dev/privkey.pem;

       location / {
           proxy_pass http://newsite_app;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
       }
   }
   ```

3. **Enable site:**
   ```bash
   sudo ln -s /etc/nginx/sites-available/newsite.aldas.dev /etc/nginx/sites-enabled/
   ```

4. **Generate SSL certificate:**
   ```bash
   sudo certbot --nginx -d newsite.aldas.dev
   ```

5. **Test and reload:**
   ```bash
   sudo nginx -t && sudo systemctl reload nginx
   ```

## 🔧 Advanced Configuration

### Load Balancing Multiple Instances
```nginx
upstream app_cluster {
    least_conn;
    server 127.0.0.1:3000 weight=3;
    server 127.0.0.1:3001 weight=2;
    server 127.0.0.1:3002 weight=1;
    keepalive 32;
}
```

### WebSocket Support
```nginx
location /ws {
    proxy_pass http://app_backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_cache_bypass $http_upgrade;
}
```

### API Rate Limiting
```nginx
# In http block
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

# In server block
location /api/ {
    limit_req zone=api burst=20 nodelay;
    proxy_pass http://api_backend;
}
```

## 📊 Monitoring & Logs

### Check logs:
```bash
# Nginx access logs
sudo tail -f /var/log/nginx/access.log

# Nginx error logs
sudo tail -f /var/log/nginx/error.log

# Certbot logs
sudo tail -f /var/log/letsencrypt/letsencrypt.log
```

### Monitor certificate expiration:
```bash
# Check all certificates
sudo certbot certificates

# Check specific certificate
openssl x509 -in /etc/letsencrypt/live/yourdomain.com/fullchain.pem -text -noout | grep "Not After"
```

## 🚀 Production Tips

1. **Use systemd services** for your applications instead of running them manually
2. **Set up monitoring** with tools like htop, netstat, or proper monitoring solutions
3. **Regular backups** of nginx configs and Let's Encrypt certificates
4. **Keep certbot updated** for security patches
5. **Use fail2ban** to protect against brute force attacks
6. **Regular security updates** with `sudo apt update && sudo apt upgrade`

## 📁 File Structure

Your nginx setup should look like this:
```
/etc/nginx/
├── nginx.conf                 # Main configuration
├── sites-available/           # Available site configs
│   ├── go-auth.aldas.dev
│   ├── project2.aldas.dev
│   └── portfolio.aldas.dev
├── sites-enabled/             # Enabled sites (symlinks)
│   ├── go-auth.aldas.dev -> ../sites-available/go-auth.aldas.dev
│   ├── project2.aldas.dev -> ../sites-available/project2.aldas.dev
│   └── portfolio.aldas.dev -> ../sites-available/portfolio.aldas.dev
└── conf.d/                    # Additional configs

/etc/letsencrypt/
└── live/
    ├── go-auth.aldas.dev/
    ├── project2.aldas.dev/
    └── portfolio.aldas.dev/
```

---

**Ready to deploy!** This guide contains everything needed to get multiple personal projects running securely with SSL certificates and nginx proxy on your VPS. Each section is designed to be followed step-by-step on a fresh Ubuntu server.
