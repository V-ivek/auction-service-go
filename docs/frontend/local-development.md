# Frontend Local Development Guide

## Overview

The auction platform frontend is a vanilla JavaScript application that provides a modern, responsive interface for real-time auction participation. The frontend is served by a lightweight Go HTTP server that handles CORS and serves static files from the `web/` directory.

## Prerequisites

### Required Software
- **Web Browser**: Chrome 70+, Firefox 65+, Safari 12+, or Edge 79+
- **Text Editor/IDE**: VS Code, WebStorm, Sublime Text, or similar
- **Go 1.21+**: For running the frontend server
- **Backend Service**: The auction backend must be running first

### Optional Tools
- **Live Server Extension**: For VS Code users (alternative to Go server)
- **Browser Developer Tools**: For debugging and performance analysis
- **Node.js**: For alternative development servers (optional)

## Quick Start

### 1. Ensure Backend is Running
```bash
# Start the backend service first
go run cmd/backend/main.go
```

Verify backend is accessible:
- Backend API: http://localhost:8081
- WebSocket: ws://localhost:8081/ws
- Health Check: http://localhost:8081/health

### 2. Start Frontend Service
The project includes a built-in Go server for serving static files:

```bash
# Start the frontend server
go run cmd/frontend/main.go
```

The frontend will be available at:
- **Frontend**: http://localhost:3000
- **Static Files**: Served from `web/` directory
- **CORS**: Enabled for development

### 3. Test the Application
1. Open http://localhost:3000 in your browser
2. Log in with one of the sample users:
   - **alice**: `1a2b3c4d-5e6f-4071-8a9b-0c1d2e3f4a5b`
   - **bob**: `2b3c4d5e-6f70-4182-9bac-1d2e3f4a5b6c`
   - **charlie**: `3c4d5e6f-7081-4293-abcd-2e3f4a5b6c7d`
   - **admin**: `4d5e6f70-8192-4b3c-bcde-3f4a5b6c7d8e`
3. View active auctions and place bids
4. Monitor real-time updates via WebSocket

## Development Environment Setup

### VS Code Setup

#### Recommended Extensions
```json
{
  "recommendations": [
    "ms-vscode.vscode-json",
    "bradlc.vscode-tailwindcss",
    "esbenp.prettier-vscode",
    "ms-vscode.live-server",
    "formulahendry.auto-rename-tag",
    "christian-kohler.path-intellisense",
    "ms-vscode.vscode-css-peek"
  ]
}
```

#### VS Code Settings
Create `.vscode/settings.json`:
```json
{
  "liveServer.settings.port": 3000,
  "liveServer.settings.root": "/web",
  "liveServer.settings.CustomBrowser": "chrome",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "css.validate": true,
  "html.autoClosingTags": true,
  "javascript.suggest.autoImports": true
}
```

### Alternative Development Servers

#### Option 1: Python HTTP Server
```bash
# Python 3
cd web
python -m http.server 3000

# Python 2
cd web
python -m SimpleHTTPServer 3000
```

#### Option 2: Node.js Live Server
```bash
# Install globally
npm install -g live-server

# Run in web directory
cd web
live-server --port=3000
```

#### Option 3: Go HTTP Server (Recommended)
```bash
# Use the built-in Go server
go run cmd/frontend/main.go
```

## Project Structure

```
web/
├── index.html          # Main application page
├── css/
│   └── style.css      # Main stylesheet
└── js/
    └── app.js         # Main application logic
```

## Key Features

### Real-Time WebSocket Communication
- Automatic connection management
- Reconnection on connection loss
- Heartbeat ping/pong for connection health
- Real-time bid updates and notifications

### Responsive Design
- Mobile-first approach
- Touch-friendly interface
- Cross-browser compatibility
- Progressive enhancement

### User Authentication
- Simple token-based authentication
- Sample users for testing
- Session persistence via localStorage
- Secure WebSocket connections

## Development Workflow

### 1. Start Both Services
```bash
# Terminal 1: Backend
go run cmd/backend/main.go

# Terminal 2: Frontend
go run cmd/frontend/main.go
```

### 2. Make Changes
- Edit files in the `web/` directory
- Refresh browser to see changes
- Use browser dev tools for debugging

### 3. Test Features
- Log in with different users
- Place bids and verify real-time updates
- Test WebSocket reconnection
- Verify bid validation

## Debugging

### Browser Developer Tools
- **Console**: View WebSocket messages and errors
- **Network**: Monitor HTTP requests and WebSocket connections
- **Elements**: Inspect DOM structure and CSS
- **Application**: Check localStorage and session data

### WebSocket Debugging
```javascript
// In browser console
// Check WebSocket connection status
console.log(auctionApp.ws.readyState);

// Monitor WebSocket messages
auctionApp.ws.onmessage = (event) => {
  console.log('WebSocket message:', event.data);
};
```

### Common Issues

#### WebSocket Connection Failed
- Ensure backend is running on port 8081
- Check CORS settings
- Verify WebSocket endpoint: `ws://localhost:8081/ws`

#### Bids Not Processing
- Check browser console for errors
- Verify user authentication
- Check WebSocket message format

#### Real-Time Updates Not Working
- Verify WebSocket connection status
- Check subscription to auction updates
- Monitor WebSocket message flow

## Testing

### Manual Testing
- Use the sample users for testing
- Test bid placement and validation
- Verify real-time updates across multiple browsers
- Test WebSocket reconnection scenarios

### Automated Testing
- Frontend tests can be added using Jest or similar
- Integration tests with the backend
- E2E tests with Playwright or Cypress

## Performance Considerations

### WebSocket Optimization
- Connection pooling for multiple users
- Message batching for high-frequency updates
- Efficient DOM updates for real-time data

### Asset Optimization
- Minify CSS and JavaScript for production
- Optimize images and assets
- Enable gzip compression

## Deployment

### Production Build
- Minify and bundle JavaScript
- Optimize CSS and assets
- Configure CORS for production domains
- Set up CDN for static assets

### Environment Configuration
- API endpoint configuration
- WebSocket endpoint configuration
- Feature flags and toggles
- Analytics and monitoring setup

## Troubleshooting

### Port Conflicts
```bash
# Check what's using port 3000
lsof -i :3000

# Kill conflicting process
kill -9 <PID>
```

### CORS Issues
- Ensure backend CORS settings match frontend origin
- Check browser console for CORS errors
- Verify WebSocket upgrade headers

### File Permissions
```bash
# Ensure web directory is readable
chmod -R 755 web/
```

## Next Steps

- Add unit tests for JavaScript functions
- Implement error boundaries and fallbacks
- Add performance monitoring
- Implement progressive web app features
- Add accessibility improvements
