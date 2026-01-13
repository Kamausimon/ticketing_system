# Separating Backend and Demo App

## Project Structure

```
ticketing_system/
├── Dockerfile                    # Backend API Docker build
├── .dockerignore                 # Excludes demo-app from backend
├── cmd/api-server/              # Backend entry point
├── internal/                     # Backend business logic
├── demo-app/                     # Separate demo frontend
│   ├── Dockerfile               # Demo app Docker build
│   ├── nginx.conf               # Web server config
│   ├── docker-entrypoint.sh     # API URL injection
│   ├── index.html               # Demo UI
│   ├── app.js                   # Demo logic
│   └── styles.css               # Demo styles
└── ...
```

## Key Separation Points

### 1. Different Dockerfiles
- **Backend**: `/Dockerfile` - Builds Go API server
- **Demo**: `/demo-app/Dockerfile` - Serves static HTML/JS with nginx

### 2. .dockerignore Exclusion
The backend's `.dockerignore` excludes `demo-app/` so it's not included in backend builds.

### 3. Independent Deployments

#### Backend on Railway
```bash
# Root directory: /
# Uses: Dockerfile
# Port: 8080
```

#### Demo on Railway (Separate Project)
```bash
# Root directory: /demo-app
# Uses: demo-app/Dockerfile
# Port: 80
```

## Deployment Options

### Option 1: Both on Railway (Separate Projects)

**Backend Project:**
1. Create project from GitHub repo
2. Railway detects `/Dockerfile`
3. Deploy with PostgreSQL + Redis
4. Generate domain: `https://backend-xxx.railway.app`

**Demo Project:**
1. Create NEW separate project
2. Link same GitHub repo
3. Set "Root Directory" to `/demo-app`
4. Set environment variable: `API_BASE_URL=https://backend-xxx.railway.app`
5. Generate domain: `https://demo-xxx.railway.app`

### Option 2: Backend on Railway, Demo on Vercel

**Backend on Railway:**
- Same as above

**Demo on Vercel:**
```bash
cd demo-app
vercel
# When prompted, set root directory to ./
```

Then update `demo-app/app.js`:
```javascript
const API_URL = 'https://your-backend.railway.app';
```

### Option 3: Backend on Railway, Demo on Netlify

```bash
cd demo-app
netlify deploy
```

Update API URL as needed.

## Configuration

### Backend Environment Variables
```env
PORT=8080
CORS_ALLOWED_ORIGINS=https://your-demo.railway.app,https://your-prod-frontend.com
```

### Demo Environment Variables
```env
API_BASE_URL=https://your-backend.railway.app
```

## Local Development

### Run Backend
```bash
cd /workspaces/ticketing_system
go run cmd/api-server/main.go
# Runs on http://localhost:8080
```

### Run Demo
```bash
cd demo-app
python3 -m http.server 3000
# Or open index.html directly
# Demo calls http://localhost:8080 by default
```

### Run Both with Docker
```bash
# Terminal 1: Backend
docker build -t ticketing-backend .
docker run -p 8080:8080 --env-file .env ticketing-backend

# Terminal 2: Demo
cd demo-app
docker build -t ticketing-demo .
docker run -p 3000:80 -e API_BASE_URL=http://localhost:8080 ticketing-demo
```

## Sharing the Demo

### Public Demo (Frontend Only)
1. Deploy demo to Vercel/Netlify (free hosting)
2. Point to your Railway backend API
3. Share URL: `https://your-demo.vercel.app`

Users can:
- ✅ View the demo interface
- ✅ Interact with your backend API
- ✅ See the ticketing system in action
- ❌ Cannot see your backend code (it's separate)

### Private Backend Access
Your Railway backend stays private:
- Requires authentication (JWT)
- Environment variables kept secret
- Database credentials not exposed
- Only API endpoints are public

## Benefits of Separation

1. **Independent Scaling**
   - Scale backend based on API load
   - Demo is just static files (minimal resources)

2. **Security**
   - Backend code not exposed in demo
   - Demo can't access backend internals
   - Separate authentication contexts

3. **Flexibility**
   - Update demo without backend redeploy
   - Multiple frontends can use same backend
   - Different deployment platforms

4. **Cost Optimization**
   - Demo on free tier (Vercel/Netlify)
   - Backend on Railway (only pay for compute)

## Testing Separation

### Verify Backend is Standalone
```bash
docker build -t test-backend .
docker run -p 8080:8080 test-backend

# Check if demo-app files are NOT in container
docker run test-backend ls -la
# Should not see demo-app/ folder
```

### Verify Demo is Standalone
```bash
cd demo-app
docker build -t test-demo .
docker run -p 3000:80 -e API_BASE_URL=http://localhost:8080 test-demo

# Visit http://localhost:3000
# Should see demo interface
```

## Common Issues

### CORS Errors
**Problem**: Demo can't connect to backend

**Solution**: Add demo URL to backend CORS settings
```env
CORS_ALLOWED_ORIGINS=https://your-demo.railway.app
```

### API URL Not Updating
**Problem**: Demo still points to localhost

**Solution**: 
1. Set `API_BASE_URL` environment variable in Railway
2. Or hardcode in `demo-app/app.js`:
```javascript
const API_URL = 'https://your-backend.railway.app';
```

### Demo Not Loading
**Problem**: Nginx returns 404

**Solution**: Check nginx.conf is properly copied in Dockerfile

## Recommended Setup for Production

```
┌─────────────────┐
│   Users/Public  │
└────────┬────────┘
         │
         │ HTTPS
         │
    ┌────▼─────┐
    │  Demo    │  (Vercel - Free)
    │ Frontend │  https://demo.yoursite.com
    └────┬─────┘
         │
         │ API Calls
         │ (with JWT)
         │
    ┌────▼─────┐
    │ Backend  │  (Railway - $5/mo)
    │   API    │  https://api.yoursite.com
    └────┬─────┘
         │
    ┌────▼─────┐
    │PostgreSQL│  (Railway - Included)
    └──────────┘
```

## Quick Commands

```bash
# Build backend locally
docker build -t backend .

# Build demo locally
cd demo-app && docker build -t demo .

# Run backend
docker run -p 8080:8080 --env-file .env backend

# Run demo
docker run -p 3000:80 -e API_BASE_URL=http://localhost:8080 demo

# Deploy backend to Railway
git push origin main  # Auto-deploys if connected

# Deploy demo to Vercel
cd demo-app && vercel --prod
```
