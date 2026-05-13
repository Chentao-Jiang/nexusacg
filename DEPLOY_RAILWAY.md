# Railway Deployment Guide

This guide walks through deploying the NexusACG Go backend to [Railway](https://railway.app).

Railway auto-detects the project using `nixpacks.toml` and handles building, deploying, and hosting the Go binary along with PostgreSQL and Redis services.

## Prerequisites

- A GitHub account with the nexusacg repository pushed
- A Railway account (sign in with GitHub)
- Alipay developer account (for payment integration)

## Step-by-Step Deployment

### 1. Push Code to GitHub

Ensure the repository is pushed with all files including `nixpacks.toml` and `railway.json`:

```bash
git add .
git commit -m "add railway deployment config"
git push
```

### 2. Sign In to Railway

Go to [railway.app](https://railway.app) and sign in with your GitHub account.

### 3. Create a New Project

- Click **"New Project"**
- Select **"Deploy from GitHub repo"**
- Authorize Railway to access your repositories if prompted
- Select the **nexusacg** repository

### 4. Add Database and Cache Services

In the Railway project dashboard:

1. Click **"+ New"** → **"Database"** → **"Add PostgreSQL"**
   - Railway provisions a PostgreSQL instance automatically
   - Environment variables (`DATABASE_URL`, `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`) are injected automatically

2. Click **"+ New"** → **"Data"** → **"Add Redis"**
   - Railway provisions a Redis instance automatically
   - The `REDIS_URL` environment variable is injected automatically

### 5. Configure Environment Variables

In the Railway dashboard, go to your service → **Variables** tab. Set the following:

| Variable | Description | Example |
|---|---|---|
| `ENV` | Runtime environment | `production` |
| `PORT` | Port the server listens on (Railway sets this automatically) | `8080` |
| `BASE_URL` | Public URL for payment callback endpoints | `https://nexusacg-production.up.railway.app` |
| `DB_HOST` | PostgreSQL host (Railway auto-injects `PGHOST`) | `$PGHOST` |
| `DB_PORT` | PostgreSQL port (Railway auto-injects `PGPORT`) | `$PGPORT` |
| `DB_NAME` | PostgreSQL database name (Railway auto-injects `PGDATABASE`) | `$PGDATABASE` |
| `DB_USER` | PostgreSQL user (Railway auto-injects `PGUSER`) | `$PGUSER` |
| `DB_PASSWORD` | PostgreSQL password (Railway auto-injects `PGPASSWORD`) | `$PGPASSWORD` |
| `REDIS_HOST` | Redis host | Extract from `REDIS_URL` or set directly |
| `REDIS_PORT` | Redis port | `6379` |
| `JWT_SECRET` | **Required.** Random secret for signing JWT tokens. Generate with `openssl rand -base64 64` | (random string) |
| `WECHAT_OAUTH_APP_ID` | WeChat OAuth AppID from open.weixin.qq.com | |
| `WECHAT_OAUTH_APP_SECRET` | WeChat OAuth AppSecret | |
| `WECHAT_PAY_APP_ID` | WeChat Pay AppID | |
| `WECHAT_PAY_MCH_ID` | WeChat Pay Merchant ID | |
| `WECHAT_PAY_APIV3_KEY` | WeChat Pay API v3 key for callback decryption | |
| `WECHAT_PAY_CERT_SERIAL` | WeChat Pay merchant certificate serial number | |
| `WECHAT_PAY_PRIVATE_KEY_PATH` | Path to WeChat Pay RSA private key PEM file | |
| `ALIPAY_APP_ID` | Alipay application ID | |
| `ALIPAY_APP_PRIVATE_KEY_PATH` | Path to Alipay RSA private key PEM file | |
| `ALIPAY_PUBLIC_KEY_PATH` | Path to Alipay RSA public key PEM file | |
| `ALIPAY_SANDBOX` | Use Alipay sandbox mode (`true` for testing) | `false` |
| `QQ_OAUTH_APP_ID` | QQ OAuth AppID from connect.qq.com | |
| `QQ_OAUTH_APP_KEY` | QQ OAuth AppKey | |
| `MODERATION_API_KEY` | Alibaba Cloud Content Security API key | |
| `MODERATION_API_SECRET` | Alibaba Cloud Content Security API secret | |
| `SMS_ACCESS_KEY_ID` | Aliyun AccessKey ID for SMS | |
| `SMS_ACCESS_KEY_SECRET` | Aliyun AccessKey Secret for SMS | |
| `SMS_SIGN_NAME` | SMS sender name (e.g. "次元链") | |
| `SMS_TEMPLATE_CODE` | SMS template code (e.g. "SMS_123456789") | |
| `ORDER_TIMEOUT_MINUTES` | Minutes before pending orders are auto-cancelled | `30` |

**Railway variable references:** You can use Railway's variable reference syntax. For example, set `DB_HOST` to `$PGHOST` to automatically use the PostgreSQL host that Railway provisions.

**Generating JWT_SECRET:**
```bash
openssl rand -base64 64
```

### 6. Upload PEM Key Files

The Alipay and WeChat Pay configurations require PEM key files. Since these are file paths, you have two options:

**Option A: Railway Secrets (Recommended)**
1. In Railway → Variables, paste the PEM file contents as the value for a variable like `ALIPAY_PRIVATE_KEY_CONTENTS`
2. Update the code to read from the environment variable directly (requires a small code change)

**Option B: Railway Volume Mounts**
1. In Railway, go to your service → **Volumes** → **Add Volume**
2. Name it (e.g., `keys`) and mount it at `/app/keys`
3. Upload the `.pem` files to the volume via the Railway dashboard
4. Set `ALIPAY_APP_PRIVATE_KEY_PATH` to `/app/keys/alipay_app_private_key.pem`
5. Set `ALIPAY_PUBLIC_KEY_PATH` to `/app/keys/alipay_public_key.pem`

### 7. Deploy

Railway automatically triggers a build when it detects changes in the connected GitHub repository. The build process:

1. Detects `nixpacks.toml`
2. Installs Go 1.23
3. Downloads dependencies (`go mod download`)
4. Builds the binary: `CGO_ENABLED=0 go build -o nexusacg ./cmd/server`
5. Starts the service: `./backend/nexusacg`

Monitor the build in the **Deployments** tab.

### 8. Get the Public URL

After a successful deployment, Railway assigns a public URL visible in the **Settings** tab under **Domains**:

```
https://nexusacg-production.up.railway.app
```

### 9. Update BASE_URL

Set the `BASE_URL` environment variable in Railway to the public URL from step 8. This is critical for payment callback URLs to work correctly.

### 10. Configure Alipay Callback URL

In the [Alipay Open Platform](https://open.alipay.com/) developer console:

1. Go to your application settings
2. Set the **Authorization Callback URL** (授权回调地址) to:
   ```
   https://your-railway-url.up.railway.app/api/v1/payment/alipay/callback
   ```
3. If using sandbox mode, configure the sandbox callback URL similarly

### 11. (Optional) Add a Custom Domain

1. In Railway → Settings → **Domains**, click **Generate Domain** or add a custom domain
2. For production, configure your own domain (e.g., `api.nexusacg.com`) via DNS

## Environment Variable Summary (Quick Copy)

Variables that must be set manually in Railway:

```
ENV=production
JWT_SECRET=<openssl rand -base64 64>
BASE_URL=https://your-app.up.railway.app
ALIPAY_APP_ID=
ALIPAY_APP_PRIVATE_KEY_PATH=
ALIPAY_PUBLIC_KEY_PATH=
ALIPAY_SANDBOX=false
WECHAT_OAUTH_APP_ID=
WECHAT_OAUTH_APP_SECRET=
WECHAT_PAY_APP_ID=
WECHAT_PAY_MCH_ID=
WECHAT_PAY_APIV3_KEY=
WECHAT_PAY_CERT_SERIAL=
WECHAT_PAY_PRIVATE_KEY_PATH=
QQ_OAUTH_APP_ID=
QQ_OAUTH_APP_KEY=
MODERATION_API_KEY=
MODERATION_API_SECRET=
SMS_ACCESS_KEY_ID=
SMS_ACCESS_KEY_SECRET=
SMS_SIGN_NAME=次元链
SMS_TEMPLATE_CODE=
ORDER_TIMEOUT_MINUTES=30
```

Railway auto-injects these from the PostgreSQL service (use variable references):

```
DB_HOST=$PGHOST
DB_PORT=$PGPORT
DB_NAME=$PGDATABASE
DB_USER=$PGUSER
DB_PASSWORD=$PGPASSWORD
```

For Redis, set `REDIS_HOST` based on the `REDIS_URL` that Railway provides, or extract the host/port from it.

## Troubleshooting

### Build fails with "go: cannot find main module"

Ensure `go.mod` exists in the `backend/` directory. The `nixpacks.toml` runs `go mod download` from the project root, which should find `backend/go.mod`.

### Build fails because binary not found

The `nixpacks.toml` builds the binary into `backend/nexusacg`. The start command references `./backend/nexusacg`. If the working directory after build is different, adjust the start command path accordingly.

### Database connection fails

Railway auto-injects `PG*` variables. Ensure `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, and `DB_PASSWORD` are set using variable references (`$PGHOST`, etc.) in the Railway dashboard. The DSN in the code uses `sslmode=disable` which works with Railway's PostgreSQL.

### Health check fails

The server must respond to `GET /health` within the healthcheck timeout (10 seconds). Verify the health endpoint is implemented and accessible.

### JWT_SECRET validation error

In production (`ENV=production`), the server will crash if `JWT_SECRET` is empty or contains "change-me". Generate a strong random value:
```bash
openssl rand -base64 64
```

### Payment callbacks not working

1. Verify `BASE_URL` matches the Railway public URL exactly (no trailing slash, correct protocol)
2. Check Alipay developer console for the correct callback URL configuration
3. Review Railway logs for callback endpoint errors

### Redis connection fails

Railway provides `REDIS_URL` (format: `redis://host:port`). Extract the host and port and set `REDIS_HOST` and `REDIS_PORT` accordingly, or set `REDIS_URL` and modify the code to use it directly.

### PEM key files not found

The application reads key files from disk. In Railway's ephemeral filesystem, you must use volume mounts (see Step 6) or modify the code to read key contents from environment variables.
