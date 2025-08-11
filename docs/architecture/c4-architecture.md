# C4 Architecture Diagrams

This document provides C4 model architecture diagrams for the Real-Time Auction Microservice platform. The C4 model consists of Context, Container, Component, and Code diagrams that provide different levels of abstraction.

## System Context Diagram (Level 1)

```mermaid
C4Context
    title System Context Diagram - Real-Time Auction Platform

    Person(user, "Auction User", "Participates in auctions, places bids")
    Person(seller, "Seller", "Creates listings and auctions")
    Person(admin, "Administrator", "Manages platform and monitors system")

    System(auction_platform, "Auction Platform", "Real-time auction and bidding system")
    
    System_Ext(email_service, "Email Service", "Sends notifications to users")
    System_Ext(payment_gateway, "Payment Gateway", "Processes payments for won auctions")
    System_Ext(monitoring, "Monitoring System", "Application monitoring and alerting")
    
    Rel(user, auction_platform, "Views auctions, places bids", "HTTPS/WSS")
    Rel(seller, auction_platform, "Creates listings, starts auctions", "HTTPS/WSS")
    Rel(admin, auction_platform, "Manages system, views metrics", "HTTPS")
    
    Rel(auction_platform, email_service, "Sends email notifications", "SMTP")
    Rel(auction_platform, payment_gateway, "Processes payments", "HTTPS")
    Rel(auction_platform, monitoring, "Sends metrics and logs", "HTTPS")
```

## Container Diagram (Level 2)

```mermaid
C4Container
    title Container Diagram - Auction Platform

    Person(user, "User", "Auction participants")
    
    Container_Boundary(auction_system, "Auction Platform") {
        Container(web_app, "Web Application", "JavaScript, HTML, CSS", "Provides auction interface via web browser")
        Container(api_gateway, "API Gateway", "Go HTTP Server", "Handles HTTP requests, authentication, rate limiting")
        Container(websocket_service, "WebSocket Service", "Go WebSocket Server", "Real-time bidding and notifications")
        Container(auction_service, "Auction Service", "Go Application", "Core auction business logic")
        Container(bid_service, "Bidding Service", "Go Application", "Bid processing and validation")
        Container(notification_service, "Notification Hub", "Go Application", "Real-time event broadcasting")
    }
    
    ContainerDb(database, "Database", "PostgreSQL", "Stores users, listings, auctions, bids")
    ContainerDb(redis_cache, "Redis Cache", "Redis", "Caching and session storage")
    Container(metrics_service, "Metrics Service", "Prometheus", "Application metrics collection")
    
    System_Ext(email_service, "Email Service", "External notification service")
    System_Ext(payment_service, "Payment Service", "External payment processing")

    Rel(user, web_app, "Uses", "HTTPS")
    Rel(web_app, api_gateway, "Makes API calls", "HTTPS/JSON")
    Rel(web_app, websocket_service, "Establishes real-time connection", "WSS")
    
    Rel(api_gateway, auction_service, "Delegates requests", "Go function calls")
    Rel(api_gateway, bid_service, "Delegates requests", "Go function calls")
    
    Rel(websocket_service, notification_service, "Publishes events", "Go channels")
    Rel(auction_service, notification_service, "Publishes events", "Go channels")
    Rel(bid_service, notification_service, "Publishes events", "Go channels")
    
    Rel(auction_service, database, "Reads/writes auction data", "SQL")
    Rel(bid_service, database, "Reads/writes bid data", "SQL")
    
    Rel(api_gateway, redis_cache, "Stores session data", "Redis protocol")
    Rel(websocket_service, redis_cache, "Stores connection data", "Redis protocol")
    
    Rel(auction_service, metrics_service, "Sends metrics", "HTTP")
    Rel(bid_service, metrics_service, "Sends metrics", "HTTP")
    Rel(websocket_service, metrics_service, "Sends metrics", "HTTP")
    
    Rel(auction_service, email_service, "Sends notifications", "SMTP/HTTP")
    Rel(bid_service, payment_service, "Processes payments", "HTTPS")
```

## Component Diagram (Level 3) - Backend Services

```mermaid
C4Component
    title Component Diagram - Backend Services

    Container_Boundary(api_gateway, "API Gateway") {
        Component(auth_middleware, "Authentication Middleware", "Go Middleware", "JWT token validation")
        Component(rate_limiter, "Rate Limiter", "Go Middleware", "Prevents API abuse")
        Component(cors_handler, "CORS Handler", "Go Middleware", "Handles cross-origin requests")
        Component(http_router, "HTTP Router", "Go HTTP Handler", "Routes requests to services")
    }

    Container_Boundary(auction_service, "Auction Service") {
        Component(auction_handler, "Auction Handler", "Go HTTP Handler", "Handles auction HTTP requests")
        Component(auction_domain, "Auction Domain", "Go Domain Model", "Auction business rules and entities")
        Component(auction_repository, "Auction Repository", "Go Interface", "Auction data access")
        Component(timeout_manager, "Timeout Manager", "Go Service", "Manages auction timeouts")
    }

    Container_Boundary(bid_service, "Bidding Service") {
        Component(bid_handler, "Bid Handler", "Go HTTP Handler", "Handles bid HTTP requests")
        Component(bid_domain, "Bid Domain", "Go Domain Model", "Bidding business rules and validation")
        Component(bid_repository, "Bid Repository", "Go Interface", "Bid data access")
        Component(bid_validator, "Bid Validator", "Go Service", "Validates bid rules and constraints")
    }

    Container_Boundary(websocket_service, "WebSocket Service") {
        Component(ws_handler, "WebSocket Handler", "Go WebSocket Handler", "Manages WebSocket connections")
        Component(connection_hub, "Connection Hub", "Go Service", "Manages active connections")
        Component(message_router, "Message Router", "Go Service", "Routes WebSocket messages")
        Component(subscription_manager, "Subscription Manager", "Go Service", "Manages auction subscriptions")
    }

    Container_Boundary(notification_service, "Notification Hub") {
        Component(event_publisher, "Event Publisher", "Go Service", "Publishes domain events")
        Component(event_subscriber, "Event Subscriber", "Go Service", "Handles domain events")
        Component(notification_dispatcher, "Notification Dispatcher", "Go Service", "Dispatches notifications to clients")
    }

    ContainerDb(database, "PostgreSQL Database", "Database", "Persistent data storage")
    ContainerDb(redis, "Redis Cache", "Cache", "Session and temporary data")

    Rel(http_router, auction_handler, "Routes auction requests")
    Rel(http_router, bid_handler, "Routes bid requests")
    
    Rel(auction_handler, auction_domain, "Uses domain logic")
    Rel(auction_domain, auction_repository, "Persists data")
    Rel(auction_domain, timeout_manager, "Manages timeouts")
    
    Rel(bid_handler, bid_domain, "Uses domain logic")
    Rel(bid_domain, bid_repository, "Persists data")
    Rel(bid_domain, bid_validator, "Validates bids")
    
    Rel(ws_handler, connection_hub, "Manages connections")
    Rel(ws_handler, message_router, "Routes messages")
    Rel(message_router, subscription_manager, "Manages subscriptions")
    
    Rel(auction_domain, event_publisher, "Publishes auction events")
    Rel(bid_domain, event_publisher, "Publishes bid events")
    Rel(event_subscriber, notification_dispatcher, "Processes events")
    Rel(notification_dispatcher, connection_hub, "Sends notifications")
    
    Rel(auction_repository, database, "Reads/writes data", "SQL")
    Rel(bid_repository, database, "Reads/writes data", "SQL")
    Rel(connection_hub, redis, "Stores connection state", "Redis")
    Rel(subscription_manager, redis, "Stores subscriptions", "Redis")
```

## Component Diagram (Level 3) - Frontend Application

```mermaid
C4Component
    title Component Diagram - Frontend Application

    Container_Boundary(web_app, "Web Application") {
        Component(app_controller, "Application Controller", "JavaScript Class", "Main application orchestration")
        Component(auth_manager, "Authentication Manager", "JavaScript Module", "Handles user authentication and tokens")
        Component(websocket_client, "WebSocket Client", "JavaScript Class", "Manages WebSocket connection and messaging")
        Component(api_client, "API Client", "JavaScript Class", "Handles REST API communication")
        
        Component(auction_view, "Auction View", "JavaScript Module", "Displays auction listings and details")
        Component(bid_interface, "Bidding Interface", "JavaScript Module", "Handles bid placement interface")
        Component(activity_feed, "Activity Feed", "JavaScript Module", "Displays real-time activity updates")
        Component(notification_system, "Notification System", "JavaScript Module", "Shows toast notifications")
        
        Component(state_manager, "State Manager", "JavaScript Class", "Manages application state")
        Component(event_system, "Event System", "JavaScript Module", "Handles DOM and custom events")
        Component(ui_components, "UI Components", "JavaScript Modules", "Reusable UI components")
        Component(utils, "Utilities", "JavaScript Functions", "Helper functions and utilities")
    }

    Container(backend_api, "Backend API", "Go HTTP Server", "Provides REST API endpoints")
    Container(websocket_service, "WebSocket Service", "Go WebSocket Server", "Real-time communication")
    Container(browser_storage, "Browser Storage", "LocalStorage/SessionStorage", "Client-side data persistence")

    Rel(app_controller, auth_manager, "Manages authentication")
    Rel(app_controller, websocket_client, "Establishes real-time connection")
    Rel(app_controller, api_client, "Makes API requests")
    Rel(app_controller, state_manager, "Manages global state")
    
    Rel(auth_manager, browser_storage, "Stores auth tokens")
    Rel(websocket_client, websocket_service, "WebSocket connection", "WSS")
    Rel(api_client, backend_api, "HTTP requests", "HTTPS/JSON")
    
    Rel(auction_view, state_manager, "Reads auction state")
    Rel(bid_interface, websocket_client, "Sends bid messages")
    Rel(activity_feed, websocket_client, "Receives real-time updates")
    Rel(notification_system, event_system, "Listens for events")
    
    Rel(auction_view, ui_components, "Uses UI components")
    Rel(bid_interface, ui_components, "Uses UI components")
    Rel(ui_components, utils, "Uses utility functions")
    
    Rel(websocket_client, state_manager, "Updates state from messages")
    Rel(api_client, state_manager, "Updates state from API responses")
```

## Data Flow Diagram

```mermaid
flowchart TB
    subgraph "Client Browser"
        A[User Interface] --> B[JavaScript App]
        B --> C[WebSocket Client]
        B --> D[HTTP Client]
    end
    
    subgraph "Load Balancer"
        E[Nginx/HAProxy]
    end
    
    subgraph "Application Layer"
        F[HTTP Server]
        G[WebSocket Server]
        H[Auction Service]
        I[Bidding Service]
        J[Notification Hub]
    end
    
    subgraph "Data Layer"
        K[(PostgreSQL)]
        L[(Redis Cache)]
        M[File Storage]
    end
    
    subgraph "External Services"
        N[Email Service]
        O[Payment Gateway]
        P[Monitoring]
    end
    
    C -.->|WSS| E
    D -->|HTTPS| E
    E --> F
    E --> G
    
    F --> H
    F --> I
    G --> J
    
    H --> K
    I --> K
    H --> L
    I --> L
    J --> L
    
    H --> N
    I --> O
    F --> P
    G --> P
    
    J -.->|Events| G
    G -.->|Messages| C
    
    style A fill:#e1f5fe
    style K fill:#fff3e0
    style L fill:#fce4ec
    style N fill:#f3e5f5
    style O fill:#f3e5f5
    style P fill:#e8f5e8
```

## Deployment Architecture

```mermaid
C4Deployment
    title Deployment Diagram - Production Environment

    Deployment_Node(cdn, "CDN", "CloudFlare") {
        Container(static_assets, "Static Assets", "HTML, CSS, JS", "Frontend assets")
    }
    
    Deployment_Node(load_balancer, "Load Balancer", "AWS ALB") {
        Container(lb, "Application Load Balancer", "AWS ALB", "Distributes traffic")
    }
    
    Deployment_Node(web_tier, "Web Tier", "AWS ECS Fargate") {
        Deployment_Node(web_cluster, "Web Cluster", "ECS Cluster") {
            Container(frontend_service, "Frontend Service", "Go HTTP Server", "Serves web application")
            Container(api_service, "API Service", "Go HTTP Server", "REST API endpoints")
            Container(websocket_service, "WebSocket Service", "Go WebSocket Server", "Real-time communication")
        }
    }
    
    Deployment_Node(app_tier, "Application Tier", "AWS ECS Fargate") {
        Deployment_Node(app_cluster, "Application Cluster", "ECS Cluster") {
            Container(auction_service, "Auction Service", "Go Microservice", "Auction business logic")
            Container(bid_service, "Bidding Service", "Go Microservice", "Bid processing")
            Container(notification_service, "Notification Service", "Go Microservice", "Event handling")
        }
    }
    
    Deployment_Node(data_tier, "Data Tier", "AWS RDS & ElastiCache") {
        ContainerDb(postgres, "PostgreSQL", "AWS RDS PostgreSQL", "Primary database")
        ContainerDb(redis, "Redis", "AWS ElastiCache", "Caching and sessions")
    }
    
    Deployment_Node(monitoring, "Monitoring", "AWS CloudWatch") {
        Container(metrics, "Metrics", "CloudWatch Metrics", "Application metrics")
        Container(logs, "Logs", "CloudWatch Logs", "Centralized logging")
        Container(alarms, "Alarms", "CloudWatch Alarms", "Monitoring alerts")
    }

    Rel(cdn, load_balancer, "Routes requests", "HTTPS")
    Rel(load_balancer, frontend_service, "Load balances", "HTTP")
    Rel(load_balancer, api_service, "Load balances", "HTTP")
    Rel(load_balancer, websocket_service, "Load balances", "WebSocket")
    
    Rel(api_service, auction_service, "Service calls", "HTTP/gRPC")
    Rel(api_service, bid_service, "Service calls", "HTTP/gRPC")
    Rel(websocket_service, notification_service, "Event streaming", "Message Queue")
    
    Rel(auction_service, postgres, "Database queries", "PostgreSQL")
    Rel(bid_service, postgres, "Database queries", "PostgreSQL")
    Rel(websocket_service, redis, "Session storage", "Redis")
    
    Rel(auction_service, metrics, "Metrics", "CloudWatch API")
    Rel(bid_service, metrics, "Metrics", "CloudWatch API")
    Rel(websocket_service, logs, "Logs", "CloudWatch Logs")
```

## Technology Stack Overview

```mermaid
graph TB
    subgraph "Frontend"
        A[HTML5/CSS3]
        B[Vanilla JavaScript]
        C[WebSocket API]
        D[Fetch API]
    end
    
    subgraph "Backend"
        E[Go 1.21+]
        F[Gorilla WebSocket]
        G[net/http]
        H[database/sql]
    end
    
    subgraph "Data Storage"
        I[PostgreSQL 13+]
        J[Redis 6+]
        K[File System]
    end
    
    subgraph "Infrastructure"
        L[Docker]
        M[Railway/AWS]
        N[Nginx]
        O[CloudFlare]
    end
    
    subgraph "Monitoring"
        P[Prometheus]
        Q[Grafana]
        R[Zap Logger]
        S[CloudWatch]
    end
    
    A --> E
    B --> F
    C --> G
    E --> I
    E --> J
    
    E --> L
    L --> M
    M --> N
    N --> O
    
    E --> P
    P --> Q
    E --> R
    M --> S
```

## Security Architecture

```mermaid
flowchart TB
    subgraph "Client Security"
        A[HTTPS Enforcement]
        B[JWT Token Storage]
        C[Input Validation]
        D[XSS Protection]
    end
    
    subgraph "Network Security"
        E[TLS 1.3]
        F[CORS Policy]
        G[Rate Limiting]
        H[DDoS Protection]
    end
    
    subgraph "Application Security"
        I[JWT Authentication]
        J[Authorization Middleware]
        K[SQL Injection Prevention]
        L[Data Validation]
    end
    
    subgraph "Infrastructure Security"
        M[VPC/Network Isolation]
        N[Security Groups]
        O[Database Encryption]
        P[Secrets Management]
    end
    
    A --> E
    B --> I
    C --> L
    D --> K
    
    E --> F
    F --> G
    G --> H
    
    I --> J
    J --> K
    K --> L
    
    M --> N
    N --> O
    O --> P
    
    style A fill:#ffebee
    style I fill:#e8f5e8
    style M fill:#fff3e0
```

## Event-Driven Architecture

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant W as WebSocket
    participant A as Auction Service
    participant B as Bid Service
    participant N as Notification Hub
    participant D as Database
    
    U->>F: Place Bid
    F->>W: Send Bid Message
    W->>B: Process Bid
    B->>D: Validate & Store Bid
    D-->>B: Confirm Storage
    B->>N: Publish Bid Event
    N->>W: Broadcast to Subscribers
    W->>F: Send Real-time Update
    F->>U: Update UI
    
    par Auction Timeout
        A->>A: Check Auction Timer
        A->>N: Publish Auction End Event
        N->>W: Broadcast Auction End
        W->>F: Notify All Clients
    and
        A->>D: Update Auction Status
    end
```

This C4 architecture documentation provides a comprehensive view of the auction platform's structure, from high-level system context down to detailed component interactions, helping stakeholders understand the system at different levels of abstraction.