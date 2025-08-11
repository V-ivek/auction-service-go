# 🚄 Railway Deployment Commands

Copy and paste these commands in order after completing the GitHub setup and Railway login.

## Prerequisites ✅
- [x] GitHub repository created and code pushed
- [ ] Railway login completed (`railway login`)

## Your Generated JWT Secret
```bash
export JWT_SECRET_KEY=""
```

## Railway Deployment Commands

### 1. Create Railway Project
```bash
railway new auction-microservice
```

### 2. Add PostgreSQL Database
```bash
railway add postgresql
```

### 3. Deploy Backend Service
```bash
# Create backend service
railway service create auction-backend

# Deploy backend with specific dockerfile
railway up --service auction-backend --dockerfile Dockerfile.backend
```

### 4. Set Backend Environment Variables
```bash
railway variables set --service auction-backend \
  JWT_SECRET_KEY="c486900a8963b8eee0584ff803db8724232389cf480b33b6959c5feeb2ec6fee" \
  ENVIRONMENT="production" \
  USE_MOCK_REPOS="false" \
  LOG_FORMAT="json" \
  METRICS_ENABLED="true" \
  LOG_LEVEL="info" \
  SERVER_HOST="0.0.0.0" \
  SERVER_PORT="8081"
```

### 5. Deploy Frontend Service
```bash
# Create frontend service
railway service create auction-frontend

# Deploy frontend with specific dockerfile  
railway up --service auction-frontend --dockerfile Dockerfile.frontend
```

### 6. Get Service URLs
```bash
# Get backend URL
echo "Backend URL:"
railway service get auction-backend

echo "Frontend URL:"
railway service get auction-frontend
```

### 7. Configure Cross-Service Communication
```bash
# Replace these URLs with the actual URLs from step 6
BACKEND_URL="https://auction-backend-production-xxxx.up.railway.app"
FRONTEND_URL="https://auction-frontend-production-xxxx.up.railway.app"

# Set frontend environment variables
railway variables set --service auction-frontend \
  BACKEND_API_URL="$BACKEND_URL" \
  BACKEND_WS_URL="${BACKEND_URL/https:/wss:}/ws" \
  SERVER_HOST="0.0.0.0" \
  SERVER_PORT="3000"

# Set backend CORS
railway variables set --service auction-backend \
  CORS_ALLOWED_ORIGINS="$FRONTEND_URL"
```

### 8. Initialize Database (Optional)
```bash
# If you want to add sample data, run this after deployment
# railway run --service auction-backend ./scripts/init-db.sh
```

## Verification Commands

### Check Service Status
```bash
railway status
```

### View Logs
```bash
# Backend logs
railway logs --service auction-backend

# Frontend logs
railway logs --service auction-frontend
```

### Test Deployment
```bash
# Test backend health (replace URL)
curl https://your-backend-url.railway.app/health

# Test API (replace URL)
curl https://your-backend-url.railway.app/api/auctions/active
```

## Quick Redeploy Commands
```bash
# Redeploy backend
railway redeploy --service auction-backend

# Redeploy frontend
railway redeploy --service auction-frontend
```

## Environment Variables Reference
```bash
# View all backend variables
railway variables --service auction-backend

# View all frontend variables  
railway variables --service auction-frontend
```

## After Successful Deployment

1. **Note your service URLs** from step 6
2. **Test both services** work correctly
3. **Set up GitHub secrets** with the URLs for CI/CD
4. **Test real-time bidding** across multiple browsers

---

**🎉 Once deployed, your auction system will be live at:**
- **Backend API**: `https://your-backend-url.railway.app`
- **Frontend UI**: `https://your-frontend-url.railway.app`
- **Health Check**: `https://your-backend-url.railway.app/health`
- **API Docs**: Available in the repository