# Frontend Local Development Guide

## Overview

The auction platform frontend is a vanilla JavaScript application that provides a modern, responsive interface for real-time auction participation. This guide covers setting up and developing the frontend locally.

## Prerequisites

### Required Software
- **Web Browser**: Chrome 70+, Firefox 65+, Safari 12+, or Edge 79+
- **Text Editor/IDE**: VS Code, WebStorm, Sublime Text, or similar
- **Local Server**: For serving static files (multiple options available)
- **Backend Service**: The auction backend must be running

### Optional Tools
- **Live Server Extension**: For VS Code users
- **Browser Developer Tools**: For debugging and performance analysis
- **Node.js**: For advanced development tools (optional)

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

### 2. Start Frontend Server
The project includes a built-in Go server for serving static files:

```bash
# Start the frontend server
go run cmd/frontend/main.go
```

The frontend will be available at:
- **Frontend**: http://localhost:3000
- **Static Files**: Served from `web/` directory

### 3. Test the Application
1. Open http://localhost:3000 in your browser
2. Enter a username to join the auction
3. View active auctions and place bids
4. Monitor real-time updates

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

# Start server
cd web
live-server --port=3000
```

#### Option 3: PHP Built-in Server
```bash
cd web
php -S localhost:3000
```

#### Option 4: VS Code Live Server Extension
1. Install "Live Server" extension
2. Right-click on `web/index.html`
3. Select "Open with Live Server"

## Project Structure

```
web/
├── index.html              # Main HTML file
├── css/
│   └── style.css          # Comprehensive stylesheet
└── js/
    └── app.js             # Main application logic

docs/frontend/             # Documentation
├── README.md              # This file
└── local-development.md   # Development guide
```

### File Organization Principles
- **Separation of Concerns**: HTML, CSS, and JS in separate files
- **Component-Based**: Logical grouping of related functionality
- **Progressive Enhancement**: Core functionality works without JavaScript
- **Maintainability**: Clear structure for easy updates

## HTML Development

### Document Structure
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Real-Time Auction Platform</title>
    
    <!-- External fonts -->
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    
    <!-- Local styles -->
    <link rel="stylesheet" href="css/style.css">
</head>
<body>
    <div id="app">
        <!-- Application content -->
    </div>
    
    <!-- Local scripts -->
    <script src="js/app.js"></script>
</body>
</html>
```

### Semantic HTML Best Practices
- Use semantic elements (`header`, `main`, `section`, `article`)
- Proper heading hierarchy (`h1` → `h2` → `h3`)
- Accessible form labels and inputs
- ARIA attributes for dynamic content

### Template Structure
```html
<!-- Header component -->
<header class="header">
    <div class="container">
        <h1 class="logo">🏆 Auction Platform</h1>
        <div class="connection-status">
            <span id="connectionStatus" class="status-dot"></span>
            <span id="connectionText">Connected</span>
        </div>
    </div>
</header>

<!-- Main content area -->
<main class="main">
    <div class="container">
        <!-- Authentication section -->
        <section id="authSection" class="auth-section">
            <!-- Auth form -->
        </section>
        
        <!-- Dashboard section -->
        <section id="dashboardSection" class="dashboard hidden">
            <!-- Auction interface -->
        </section>
    </div>
</main>
```

## CSS Development

### Architecture Overview
The CSS follows a component-based architecture:

```css
/* 1. Reset and base styles */
* { margin: 0; padding: 0; box-sizing: border-box; }

/* 2. CSS Custom Properties (Variables) */
:root {
  --primary-color: #667eea;
  --secondary-color: #764ba2;
  /* ... more variables */
}

/* 3. Base typography and layout */
body { font-family: 'Inter', sans-serif; }

/* 4. Component styles */
.header { /* header styles */ }
.auth-section { /* auth styles */ }
.dashboard { /* dashboard styles */ }

/* 5. Utility classes */
.hidden { display: none !important; }
.btn { /* button component */ }
```

### CSS Custom Properties
```css
:root {
  /* Colors */
  --primary-color: #667eea;
  --secondary-color: #764ba2;
  --success-color: #10b981;
  --error-color: #ef4444;
  --warning-color: #f59e0b;
  
  /* Typography */
  --font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-size-sm: 0.875rem;
  --font-size-base: 1rem;
  --font-size-lg: 1.125rem;
  
  /* Spacing */
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
  
  /* Borders and shadows */
  --border-radius: 8px;
  --border-color: #e5e7eb;
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}
```

### Responsive Design Strategy
```css
/* Mobile-first approach */
.container {
  max-width: 100%;
  padding: 0 1rem;
  margin: 0 auto;
}

/* Tablet breakpoint */
@media (min-width: 768px) {
  .container {
    max-width: 768px;
    padding: 0 2rem;
  }
  
  .auctions-list {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* Desktop breakpoint */
@media (min-width: 1024px) {
  .container {
    max-width: 1200px;
  }
  
  .auctions-list {
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  }
}
```

### Animation and Transitions
```css
/* Smooth transitions */
.btn {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 25px rgba(102, 126, 234, 0.3);
}

/* Modal animations */
.modal {
  opacity: 0;
  transform: scale(0.95);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.modal.show {
  opacity: 1;
  transform: scale(1);
}
```

## JavaScript Development

### Application Architecture
The application uses a class-based architecture without external frameworks:

```javascript
// Main application class
class AuctionApp {
  constructor() {
    // Configuration
    this.config = {
      wsUrl: 'ws://localhost:8081/ws',
      apiUrl: 'http://localhost:8081/api',
      reconnectDelay: 3000,
      heartbeatInterval: 30000
    };
    
    // State management
    this.ws = null;
    this.currentUser = null;
    this.auctions = new Map();
    this.listings = new Map();
    this.currentAuctionId = null;
    
    // Initialize
    this.init();
  }
  
  init() {
    this.bindEvents();
    this.showAuthSection();
    this.updateConnectionStatus('Disconnected', false);
  }
}
```

### Event Management
```javascript
bindEvents() {
  // Authentication events
  document.getElementById('loginForm')
    .addEventListener('submit', (e) => {
      e.preventDefault();
      this.handleLogin();
    });

  // Modal events
  document.getElementById('closeModal')
    .addEventListener('click', () => this.closeModal());

  // Keyboard events
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      this.closeModal();
    }
  });

  // Bidding events
  document.getElementById('bidForm')
    .addEventListener('submit', (e) => {
      e.preventDefault();
      this.handleBid();
    });
}
```

### WebSocket Management
```javascript
connectWebSocket() {
  if (this.ws) {
    this.ws.close();
  }

  this.updateConnectionStatus('Connecting...', false);

  try {
    const wsUrl = `${this.config.wsUrl}?token=${encodeURIComponent(this.currentUser.token)}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.updateConnectionStatus('Connected', true);
      this.startHeartbeat();
      this.showToast('Connected to auction platform', 'success');
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleWebSocketMessage(message);
    };

    this.ws.onclose = (event) => {
      this.updateConnectionStatus('Disconnected', false);
      this.clearHeartbeat();
      
      if (this.currentUser && !event.wasClean) {
        this.scheduleReconnect();
      }
    };

  } catch (error) {
    console.error('WebSocket connection failed:', error);
    this.updateConnectionStatus('Connection Failed', false);
  }
}
```

### State Management
```javascript
// Local state management patterns
class StateManager {
  constructor() {
    this.state = {
      auctions: new Map(),
      listings: new Map(),
      user: null,
      currentAuction: null
    };
    this.subscribers = new Map();
  }
  
  setState(key, value) {
    this.state[key] = value;
    this.notifySubscribers(key, value);
  }
  
  getState(key) {
    return this.state[key];
  }
  
  subscribe(key, callback) {
    if (!this.subscribers.has(key)) {
      this.subscribers.set(key, []);
    }
    this.subscribers.get(key).push(callback);
  }
  
  notifySubscribers(key, value) {
    const callbacks = this.subscribers.get(key) || [];
    callbacks.forEach(callback => callback(value));
  }
}
```

### API Integration
```javascript
// RESTful API helper methods
class APIClient {
  constructor(baseURL) {
    this.baseURL = baseURL;
  }
  
  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    
    try {
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers
        },
        ...options
      });
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      return await response.json();
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }
  
  async get(endpoint) {
    return this.request(endpoint);
  }
  
  async post(endpoint, data) {
    return this.request(endpoint, {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }
}
```

## Debugging and Development Tools

### Browser Developer Tools

#### Console Debugging
```javascript
// Enhanced console logging
class Logger {
  static debug(message, data = null) {
    if (process.env.NODE_ENV === 'development') {
      console.log(`[DEBUG] ${message}`, data);
    }
  }
  
  static error(message, error = null) {
    console.error(`[ERROR] ${message}`, error);
  }
  
  static warn(message, data = null) {
    console.warn(`[WARN] ${message}`, data);
  }
}

// Usage
Logger.debug('WebSocket message received', message);
Logger.error('Bid placement failed', error);
```

#### Network Monitoring
- Use Network tab to monitor API calls
- Check WebSocket connection status
- Monitor real-time message flow
- Analyze response times and payloads

#### Performance Profiling
```javascript
// Performance measurement
function measurePerformance(name, fn) {
  performance.mark(`${name}-start`);
  const result = fn();
  performance.mark(`${name}-end`);
  performance.measure(name, `${name}-start`, `${name}-end`);
  
  const measurements = performance.getEntriesByName(name);
  console.log(`${name} took ${measurements[0].duration}ms`);
  
  return result;
}

// Usage
const auctions = measurePerformance('loadAuctions', () => {
  return this.loadAuctions();
});
```

### Development Configuration

#### Environment Detection
```javascript
// Detect environment
const isDevelopment = window.location.hostname === 'localhost' ||
                     window.location.hostname === '127.0.0.1';

// Configuration based on environment
const config = {
  wsUrl: isDevelopment ? 'ws://localhost:8081/ws' : 'wss://your-domain.com/ws',
  apiUrl: isDevelopment ? 'http://localhost:8081/api' : 'https://your-domain.com/api',
  debug: isDevelopment,
  logLevel: isDevelopment ? 'debug' : 'error'
};
```

#### Debug Mode
```javascript
// Enable debug mode
if (config.debug) {
  // Expose app instance globally for debugging
  window.auctionApp = auctionApp;
  
  // Enhanced logging
  console.log('Debug mode enabled');
  console.log('Configuration:', config);
  
  // WebSocket message logging
  const originalSend = WebSocket.prototype.send;
  WebSocket.prototype.send = function(data) {
    console.log('WebSocket send:', data);
    originalSend.call(this, data);
  };
}
```

## Testing During Development

### Manual Testing Checklist

#### Authentication Flow
- [ ] Enter username and join auction
- [ ] Verify token generation and storage
- [ ] Test logout functionality
- [ ] Check persistent login across refreshes

#### Auction Interface
- [ ] Load and display active auctions
- [ ] Verify real-time timer updates
- [ ] Test auction card interactions
- [ ] Check responsive design on different screen sizes

#### Bidding Functionality
- [ ] Open auction detail modal
- [ ] Place valid bids
- [ ] Test bid validation (minimum amount)
- [ ] Verify real-time bid updates
- [ ] Check bid history display

#### WebSocket Connection
- [ ] Initial connection establishment
- [ ] Connection status indicator accuracy
- [ ] Auto-reconnection after network loss
- [ ] Heartbeat mechanism
- [ ] Message handling and parsing

#### Real-Time Updates
- [ ] Live bid updates across multiple browser tabs
- [ ] Auction status changes
- [ ] Activity feed updates
- [ ] Toast notifications

### Browser Testing

#### Cross-Browser Testing
```bash
# Test in different browsers
open -a "Google Chrome" http://localhost:3000
open -a "Firefox" http://localhost:3000
open -a "Safari" http://localhost:3000
open -a "Microsoft Edge" http://localhost:3000
```

#### Mobile Testing
- Use browser developer tools device emulation
- Test on actual mobile devices
- Verify touch interactions
- Check responsive layout

### Performance Testing

#### Page Load Performance
```javascript
// Measure page load time
window.addEventListener('load', () => {
  const loadTime = performance.timing.loadEventEnd - 
                   performance.timing.navigationStart;
  console.log(`Page loaded in ${loadTime}ms`);
});
```

#### Memory Usage Monitoring
```javascript
// Monitor memory usage (Chrome only)
if ('memory' in performance) {
  setInterval(() => {
    const memory = performance.memory;
    console.log({
      used: Math.round(memory.usedJSHeapSize / 1024 / 1024) + 'MB',
      total: Math.round(memory.totalJSHeapSize / 1024 / 1024) + 'MB',
      limit: Math.round(memory.jsHeapSizeLimit / 1024 / 1024) + 'MB'
    });
  }, 10000);
}
```

## Common Development Issues

### 1. CORS Issues
If accessing backend from different port:
```javascript
// Backend must allow frontend origin
// Check backend CORS configuration
fetch('http://localhost:8081/api/auctions/active')
  .catch(error => {
    if (error.message.includes('CORS')) {
      console.error('CORS error - check backend configuration');
    }
  });
```

### 2. WebSocket Connection Issues
```javascript
// Debug WebSocket connection
this.ws.onerror = (error) => {
  console.error('WebSocket error:', error);
  console.log('WebSocket URL:', this.ws.url);
  console.log('Ready State:', this.ws.readyState);
};
```

### 3. Event Listener Memory Leaks
```javascript
// Proper cleanup
class ComponentManager {
  constructor() {
    this.eventListeners = [];
  }
  
  addEventListener(element, event, handler) {
    element.addEventListener(event, handler);
    this.eventListeners.push({ element, event, handler });
  }
  
  cleanup() {
    this.eventListeners.forEach(({ element, event, handler }) => {
      element.removeEventListener(event, handler);
    });
    this.eventListeners = [];
  }
}
```

### 4. DOM Manipulation Performance
```javascript
// Efficient DOM updates
function updateAuctionList(auctions) {
  // Create document fragment for batch updates
  const fragment = document.createDocumentFragment();
  
  auctions.forEach(auction => {
    const element = createAuctionElement(auction);
    fragment.appendChild(element);
  });
  
  // Single DOM update
  const container = document.getElementById('auctionsList');
  container.innerHTML = '';
  container.appendChild(fragment);
}
```

## Code Organization Best Practices

### 1. Modular Structure
```javascript
// Separate concerns into modules
const AuctionModule = {
  state: new Map(),
  
  add(auction) {
    this.state.set(auction.id, auction);
    this.render();
  },
  
  update(auctionId, data) {
    const auction = this.state.get(auctionId);
    if (auction) {
      Object.assign(auction, data);
      this.render();
    }
  },
  
  render() {
    // Update UI
  }
};
```

### 2. Configuration Management
```javascript
// Centralized configuration
const Config = {
  development: {
    wsUrl: 'ws://localhost:8081/ws',
    apiUrl: 'http://localhost:8081/api',
    debug: true,
    logLevel: 'debug'
  },
  
  production: {
    wsUrl: 'wss://api.yourdomain.com/ws',
    apiUrl: 'https://api.yourdomain.com',
    debug: false,
    logLevel: 'error'
  },
  
  get current() {
    const env = location.hostname === 'localhost' ? 'development' : 'production';
    return this[env];
  }
};
```

### 3. Error Handling Strategy
```javascript
// Global error handling
window.addEventListener('error', (event) => {
  console.error('Global error:', event.error);
  // Report to error tracking service
});

window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason);
  event.preventDefault();
});

// Application-specific error handling
class ErrorHandler {
  static handle(error, context = '') {
    console.error(`Error in ${context}:`, error);
    
    // Show user-friendly message
    this.showUserError(this.getUserMessage(error));
    
    // Report to monitoring service
    if (Config.current.errorReporting) {
      this.reportError(error, context);
    }
  }
  
  static getUserMessage(error) {
    if (error.message.includes('network')) {
      return 'Connection error. Please check your internet connection.';
    }
    return 'An unexpected error occurred. Please try again.';
  }
}
```

## Hot Reload Development

### Live Reload Setup
For automatic browser refresh during development:

```html
<!-- Add to development HTML -->
<script>
  if (location.hostname === 'localhost') {
    const ws = new WebSocket('ws://localhost:35729');
    ws.onopen = () => console.log('Live reload connected');
    ws.onmessage = () => location.reload();
  }
</script>
```

### File Watcher Script
Create `scripts/watch.js` for Node.js environments:
```javascript
const fs = require('fs');
const path = require('path');

const watchDirectory = path.join(__dirname, '../web');
const clients = [];

// Watch for file changes
fs.watch(watchDirectory, { recursive: true }, (eventType, filename) => {
  console.log(`File changed: ${filename}`);
  clients.forEach(client => client.send('reload'));
});

// Simple WebSocket server for reload notifications
const WebSocket = require('ws');
const wss = new WebSocket.Server({ port: 35729 });

wss.on('connection', (ws) => {
  console.log('Live reload client connected');
  clients.push(ws);
  
  ws.on('close', () => {
    const index = clients.indexOf(ws);
    if (index > -1) clients.splice(index, 1);
  });
});
```

This comprehensive guide provides everything needed for effective frontend development of the auction platform.