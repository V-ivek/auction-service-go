# 🚀 Step-by-Step Deployment Guide

This guide will walk you through deploying your auction microservice to GitHub and Railway.

## Generated JWT Secret (SAVE THIS!)
```
JWT_SECRET_KEY=c486900a8963b8eee0584ff803db8724232389cf480b33b6959c5feeb2ec6fee
```

## Step 1: Create GitHub Repository

1. **Go to [GitHub.com](https://github.com)** and log in
2. **Click "New repository"** (green "New" button)
3. **Set repository details**:
   - Name: `auction-microservice`
   - Description: `Real-time auction and bidding microservice with WebSocket support`
   - Visibility: Public or Private (your choice)
   - **Important**: Don't initialize with README, .gitignore, or license
4. **Click "Create repository"**

## Step 2: Push Code to GitHub

After creating the repository, run these commands in your terminal:

```bash
# Replace 'yourusername' with your actual GitHub username
git remote add origin https://github.com/yourusername/auction-microservice.git
git branch -M main
git push -u origin main
```

## Step 3: Set Up GitHub Repository Secrets

1. **Go to your repository** on GitHub
2. **Click "Settings" tab**
3. **Click "Secrets and variables" → "Actions"**
4. **Add these repository secrets** (click "New repository secret"):

### Required Secrets:
```
Name: JWT_SECRET_KEY
Value: c486900a8963b8eee0584ff803db8724232389cf480b33b6959c5feeb2ec6fee

Name: RAILWAY_TOKEN
Value: [Get this from Railway dashboard after step 4]
```

### Additional Secrets (will be set after deployment):
```
Name: BACKEND_URL
Value: [Railway backend URL - set after deployment]

Name: FRONTEND_URL  
Value: [Railway frontend URL - set after deployment]

Name: BACKEND_WS_URL
Value: [Railway backend WebSocket URL - set after deployment]
```

## Step 4: Deploy to Railway

### 4.1 Login to Railway
```bash
railway login
```
This will open your browser for authentication.

### 4.2 Create New Project
```bash
railway new auction-microservice
cd auction-microservice  # if it created a new directory
```

### 4.3 Add PostgreSQL Database
```bash
railway add postgresql
```

### 4.4 Deploy Backend Service
```bash
# Create backend service
railway service create auction-backend

# Deploy backend
railway up --service auction-backend --dockerfile Dockerfile.backend
```

### 4.5 Set Backend Environment Variables
```bash
railway variables set --service auction-backend \
  JWT_SECRET_KEY="c486900a8963b8eee0584ff803db8724232389cf480b33b6959c5feeb2ec6fee" \
  ENVIRONMENT="production" \
  USE_MOCK_REPOS="false" \
  LOG_FORMAT="json" \
  METRICS_ENABLED="true" \
  LOG_LEVEL="info"
```

### 4.6 Deploy Frontend Service
```bash
# Create frontend service
railway service create auction-frontend

# Deploy frontend
railway up --service auction-frontend --dockerfile Dockerfile.frontend
```

### 4.7 Configure Service URLs
After both services are deployed, get their URLs:

```bash
# Get backend URL
railway service get auction-backend

# Get frontend URL  
railway service get auction-frontend
```

### 4.8 Update Environment Variables with URLs
```bash
# Set frontend environment (replace with actual URLs)
railway variables set --service auction-frontend \
  BACKEND_API_URL="https://your-backend-url.railway.app" \
  BACKEND_WS_URL="wss://your-backend-url.railway.app/ws"

# Set backend CORS (replace with actual frontend URL)
railway variables set --service auction-backend \
  CORS_ALLOWED_ORIGINS="https://your-frontend-url.railway.app"
```

### 4.9 Initialize Database
```bash
# Connect to your backend service and run database init
railway run --service auction-backend ./scripts/init-db.sh
```

## Step 5: Update GitHub Secrets

After Railway deployment, update these GitHub secrets:

1. **Get Railway Token**:
   - Go to [Railway Dashboard](https://railway.app/dashboard)
   - Click on your profile → "Account Settings" → "Tokens"
   - Create new token and copy it

2. **Add/Update GitHub Secrets**:
   ```
   RAILWAY_TOKEN=[your-railway-token]
   BACKEND_URL=https://your-backend-url.railway.app
   FRONTEND_URL=https://your-frontend-url.railway.app  
   BACKEND_WS_URL=wss://your-backend-url.railway.app/ws
   ```

## Step 6: Test Deployment

### 6.1 Test Backend Health
```bash
curl https://your-backend-url.railway.app/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2025-08-09T17:04:05.240081+04:00",
  "version": "1.0.0"
}
```

### 6.2 Test API Endpoints
```bash
# Test active auctions
curl https://your-backend-url.railway.app/api/auctions/active

# Test listings
curl https://your-backend-url.railway.app/api/listings
```

### 6.3 Test Frontend
Open your frontend URL in multiple browsers and test:
1. **Connection**: Should show "🟢 Connected"
2. **Real-time bidding**: Place bids in one browser, see updates in others
3. **WebSocket**: Check browser dev tools → Network → WS tab

## Step 7: Enable Automatic Deployments

Now when you push to the `main` branch, GitHub Actions will:
1. ✅ Run tests
2. ✅ Build Docker images  
3. ✅ Deploy to Railway
4. ✅ Run health checks
5. ✅ Notify deployment status

## 🎉 Success Criteria

Your deployment is successful when:

- ✅ **Backend Health**: `GET /health` returns 200
- ✅ **API Works**: `GET /api/auctions/active` returns data  
- ✅ **Frontend Loads**: Static files serve correctly
- ✅ **WebSocket Connects**: Real-time connection established
- ✅ **Cross-browser Sync**: Bids appear instantly across browsers
- ✅ **HTTPS**: All connections use SSL/TLS
- ✅ **Database**: Data persists between restarts

## 🔧 Useful Railway Commands

```bash
# View logs
railway logs --service auction-backend
railway logs --service auction-frontend

# Check service status
railway status

# Open service in browser
railway service open auction-frontend
railway service open auction-backend

# Redeploy service
railway redeploy --service auction-backend

# List environment variables
railway variables --service auction-backend

# Connect to database
railway connect postgres
```

## 🐛 Troubleshooting

### Backend Won't Start
```bash
# Check logs
railway logs --service auction-backend

# Verify environment variables
railway variables --service auction-backend

# Redeploy
railway redeploy --service auction-backend
```

### WebSocket Connection Issues
- Check CORS_ALLOWED_ORIGINS includes frontend URL
- Verify WebSocket URL uses `wss://` (not `ws://`)
- Check browser console for connection errors

### Database Connection Issues
- Ensure PostgreSQL service is running
- Check DATABASE_URL is set automatically by Railway
- Run database initialization script

---

**🎯 Need Help?**
- Railway Docs: https://docs.railway.app
- GitHub Actions Logs: Your repository → Actions tab
- Railway Dashboard: https://railway.app/dashboard