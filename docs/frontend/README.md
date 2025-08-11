# Frontend Documentation

This directory contains documentation for the auction platform frontend.

## Overview

The frontend is a vanilla JavaScript application that provides a modern, responsive interface for real-time auction participation. It communicates with the backend via REST API and WebSocket connections.

## Documentation

- **[Local Development Guide](local-development.md)**: Complete guide for setting up and developing the frontend locally

## Key Features

- **Real-Time Updates**: WebSocket-powered live auction updates
- **Responsive Design**: Mobile-first, touch-friendly interface
- **User Authentication**: Simple token-based auth with sample users
- **Bid Management**: Real-time bid placement and validation
- **Activity Feeds**: Live updates for all auction activities

## Technology Stack

- **HTML5**: Semantic markup and modern HTML features
- **CSS3**: Flexbox, Grid, animations, and responsive design
- **Vanilla JavaScript**: ES6+ features, no frameworks
- **WebSocket API**: Real-time communication
- **Fetch API**: HTTP requests to backend

## Project Structure

```
web/
├── index.html          # Main application page
├── css/
│   └── style.css      # Main stylesheet
└── js/
    └── app.js         # Main application logic
```

## Quick Start

1. **Start Backend**: `go run cmd/backend/main.go`
2. **Start Frontend**: `go run cmd/frontend/main.go`
3. **Open Browser**: Navigate to http://localhost:3000
4. **Login**: Use sample user credentials from the main README

## Development

See the [Local Development Guide](local-development.md) for detailed development setup and workflow information.
