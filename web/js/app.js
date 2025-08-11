// Auction Platform Frontend Application
class AuctionApp {
    constructor() {
        this.ws = null;
        this.currentUser = null;
        this.auctions = new Map();
        this.listings = new Map();
        this.currentAuctionId = null;
        this.reconnectInterval = null;
        this.heartbeatInterval = null;
        
        // Configuration
        this.config = {
            wsUrl: 'ws://localhost:8081/ws',
            apiUrl: 'http://localhost:8081/api',
            reconnectDelay: 3000,
            heartbeatInterval: 30000
        };

        this.init();
    }

    init() {
        this.bindEvents();
        this.showAuthSection();
        this.updateConnectionStatus('Disconnected', false);
        
        // Auto-connect if user is already authenticated
        const savedUser = localStorage.getItem('auctionUser');
        if (savedUser) {
            this.currentUser = JSON.parse(savedUser);
            this.showDashboard();
            this.connectWebSocket();
        }
    }

    bindEvents() {
        // Authentication
        document.getElementById('loginForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleLogin();
        });

        document.getElementById('logoutBtn').addEventListener('click', () => {
            this.handleLogout();
        });

        // Modal controls
        document.getElementById('closeModal').addEventListener('click', () => {
            this.closeModal();
        });

        document.getElementById('auctionModal').addEventListener('click', (e) => {
            if (e.target.id === 'auctionModal') {
                this.closeModal();
            }
        });

        // Bidding
        document.getElementById('bidForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleBid();
        });

        // ESC key to close modal
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeModal();
            }
        });
    }

    // Authentication
    handleLogin() {
        const username = document.getElementById('username').value.trim();
        if (!username) {
            this.showToast('Please enter a username', 'error');
            return;
        }

        // Simple JWT-like token (for demo purposes)
        const token = btoa(JSON.stringify({
            username: username,
            timestamp: Date.now(),
            exp: Date.now() + (24 * 60 * 60 * 1000) // 24 hours
        }));

        this.currentUser = { username, token };
        localStorage.setItem('auctionUser', JSON.stringify(this.currentUser));
        
        this.showDashboard();
        this.connectWebSocket();
    }

    handleLogout() {
        this.currentUser = null;
        localStorage.removeItem('auctionUser');
        
        if (this.ws) {
            this.ws.close();
        }
        
        this.clearIntervals();
        this.showAuthSection();
        this.updateConnectionStatus('Disconnected', false);
        this.showToast('Logged out successfully', 'success');
    }

    showAuthSection() {
        document.getElementById('authSection').classList.remove('hidden');
        document.getElementById('dashboardSection').classList.add('hidden');
        document.getElementById('username').value = '';
    }

    showDashboard() {
        document.getElementById('authSection').classList.add('hidden');
        document.getElementById('dashboardSection').classList.remove('hidden');
        document.getElementById('currentUser').textContent = this.currentUser.username;
        
        this.loadAuctions();
        this.addActivityMessage('Connected to auction platform', 'system');
    }

    // WebSocket Connection
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
                this.clearReconnectInterval();
                this.startHeartbeat();
                this.showToast('Connected to auction platform', 'success');
            };

            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleWebSocketMessage(message);
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };

            this.ws.onclose = (event) => {
                console.log('WebSocket disconnected:', event.code, event.reason);
                this.updateConnectionStatus('Disconnected', false);
                this.clearHeartbeat();
                
                if (this.currentUser && !event.wasClean) {
                    this.scheduleReconnect();
                }
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.updateConnectionStatus('Connection Error', false);
                this.showToast('Connection error occurred', 'error');
            };

        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
            this.updateConnectionStatus('Connection Failed', false);
            this.scheduleReconnect();
        }
    }

    scheduleReconnect() {
        this.clearReconnectInterval();
        this.reconnectInterval = setTimeout(() => {
            if (this.currentUser) {
                console.log('Attempting to reconnect...');
                this.connectWebSocket();
            }
        }, this.config.reconnectDelay);
    }

    clearReconnectInterval() {
        if (this.reconnectInterval) {
            clearTimeout(this.reconnectInterval);
            this.reconnectInterval = null;
        }
    }

    startHeartbeat() {
        this.heartbeatInterval = setInterval(() => {
            if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                this.sendWebSocketMessage({
                    action: 'ping',
                    timestamp: Date.now()
                });
            }
        }, this.config.heartbeatInterval);
    }

    clearHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }
    }

    clearIntervals() {
        this.clearReconnectInterval();
        this.clearHeartbeat();
    }

    updateConnectionStatus(text, isConnected) {
        const statusElement = document.getElementById('connectionStatus');
        const textElement = document.getElementById('connectionText');
        
        statusElement.className = `status-dot ${isConnected ? 'online' : 'offline'}`;
        textElement.textContent = text;
    }

    // WebSocket Message Handling
    handleWebSocketMessage(message) {
        console.log('Received WebSocket message:', message);

        switch (message.event) {
            case 'bid_placed':
                this.handleBidPlacedEvent(message.data);
                break;
            case 'auction_started':
                this.handleAuctionStartedEvent(message.data);
                break;
            case 'auction_ended':
                this.handleAuctionEndedEvent(message.data);
                break;
            case 'auction_updated':
                this.handleAuctionUpdatedEvent(message.data);
                break;
            case 'error':
                this.handleErrorEvent(message.data);
                break;
            case 'pong':
                // Heartbeat response - do nothing
                break;
            default:
                console.log('Unknown message event:', message.event);
        }
    }

    handleBidPlacedEvent(data) {
        const auction = this.auctions.get(data.auction_id);
        if (auction) {
            auction.current_bid = data.amount;
            auction.bid_count = (auction.bid_count || 0) + 1;
            this.auctions.set(data.auction_id, auction);
            
            if (this.currentAuctionId === data.auction_id) {
                this.updateModalAuctionData(auction);
                this.addBidToList(data);
            }
            
            this.updateAuctionCard(data.auction_id);
        }
        
        const isOwnBid = data.user_id === this.currentUser?.username;
        const message = isOwnBid 
            ? `You placed a bid of $${data.amount.toFixed(2)}`
            : `${data.user_id} bid $${data.amount.toFixed(2)}`;
            
        this.addActivityMessage(message, 'bid');
        
        if (!isOwnBid) {
            this.showToast(`New bid: $${data.amount.toFixed(2)}`, 'success');
        }
    }

    handleAuctionStartedEvent(data) {
        this.auctions.set(data.id, data);
        this.addActivityMessage(`Auction started: ${data.listing?.title || data.id}`, 'auction');
        this.loadAuctions(); // Refresh the auction list
        this.showToast('New auction started!', 'success');
    }

    handleAuctionEndedEvent(data) {
        const auction = this.auctions.get(data.id);
        if (auction) {
            auction.status = 'ended';
            this.auctions.set(data.id, auction);
            this.updateAuctionCard(data.id);
        }
        
        this.addActivityMessage(`Auction ended: ${data.listing?.title || data.id}`, 'auction');
        this.showToast('Auction ended', 'warning');
        
        if (this.currentAuctionId === data.id) {
            this.closeModal();
        }
    }

    handleAuctionUpdatedEvent(data) {
        this.auctions.set(data.id, data);
        this.updateAuctionCard(data.id);
        
        if (this.currentAuctionId === data.id) {
            this.updateModalAuctionData(data);
        }
    }

    handleErrorEvent(data) {
        console.error('WebSocket error:', data);
        this.showToast(data.message || 'An error occurred', 'error');
    }

    sendWebSocketMessage(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            console.error('WebSocket is not connected');
            this.showToast('Not connected to server', 'error');
        }
    }

    // API Calls
    async loadAuctions() {
        try {
            const response = await fetch(`${this.config.apiUrl}/auctions/active`);
            const result = await response.json();
            
            if (result.success && result.data.auctions) {
                // Load listings for each auction
                const auctionPromises = result.data.auctions.map(async (auction) => {
                    const listing = await this.loadListing(auction.listing_id);
                    return { ...auction, listing };
                });
                
                const auctionsWithListings = await Promise.all(auctionPromises);
                
                // Store auctions
                auctionsWithListings.forEach(auction => {
                    this.auctions.set(auction.id, auction);
                });
                
                this.renderAuctions(auctionsWithListings);
            }
        } catch (error) {
            console.error('Failed to load auctions:', error);
            this.showToast('Failed to load auctions', 'error');
        }
    }

    async loadListing(listingId) {
        try {
            // Check cache first
            if (this.listings.has(listingId)) {
                return this.listings.get(listingId);
            }
            
            const response = await fetch(`${this.config.apiUrl}/listings/${listingId}`);
            const result = await response.json();
            
            if (result.success && result.data) {
                this.listings.set(listingId, result.data);
                return result.data;
            }
        } catch (error) {
            console.error('Failed to load listing:', error);
            return null;
        }
    }

    // UI Rendering
    renderAuctions(auctions) {
        const container = document.getElementById('auctionsList');
        
        if (!auctions || auctions.length === 0) {
            container.innerHTML = '<div class="no-auctions">No active auctions at the moment.</div>';
            return;
        }

        const auctionCards = auctions.map(auction => this.createAuctionCard(auction)).join('');
        container.innerHTML = auctionCards;
        
        // Start timers for each auction
        auctions.forEach(auction => {
            if (auction.status === 'active') {
                this.startAuctionTimer(auction.id);
            }
        });
    }

    createAuctionCard(auction) {
        const listing = auction.listing || {};
        const currentBid = auction.current_bid || listing.reserve_price || 0;
        const timeLeft = this.formatTimeRemaining(auction.end_time);
        
        return `
            <div class="auction-card" onclick="auctionApp.openAuctionModal('${auction.id}')">
                <h4>${listing.title || 'Unknown Item'}</h4>
                <p>${listing.description || 'No description available'}</p>
                <div class="auction-meta">
                    <span>Current Bid: <span class="price">$${currentBid.toFixed(2)}</span></span>
                    <span>Time Left: <span class="time" id="timer-${auction.id}">${timeLeft}</span></span>
                </div>
                <div class="auction-meta">
                    <span>Bids: ${auction.bid_count || 0}</span>
                    <span>Status: ${auction.status}</span>
                </div>
            </div>
        `;
    }

    updateAuctionCard(auctionId) {
        const auction = this.auctions.get(auctionId);
        if (!auction) return;
        
        // Find and update the card
        const cards = document.querySelectorAll('.auction-card');
        cards.forEach(card => {
            if (card.onclick && card.onclick.toString().includes(auctionId)) {
                const listing = auction.listing || {};
                const currentBid = auction.current_bid || listing.reserve_price || 0;
                
                // Update price
                const priceElement = card.querySelector('.price');
                if (priceElement) {
                    priceElement.textContent = `$${currentBid.toFixed(2)}`;
                }
                
                // Update bid count
                const bidCountElement = card.querySelector('.auction-meta span:last-child');
                if (bidCountElement) {
                    bidCountElement.textContent = `Bids: ${auction.bid_count || 0}`;
                }
            }
        });
    }

    startAuctionTimer(auctionId) {
        const updateTimer = () => {
            const auction = this.auctions.get(auctionId);
            if (!auction || auction.status !== 'active') return;
            
            const timerElement = document.getElementById(`timer-${auctionId}`);
            if (timerElement) {
                const timeLeft = this.formatTimeRemaining(auction.end_time);
                timerElement.textContent = timeLeft;
                
                // Check if auction should end
                if (new Date(auction.end_time) <= new Date()) {
                    timerElement.textContent = 'Ended';
                    auction.status = 'ended';
                    return;
                }
                
                // Continue timer
                setTimeout(updateTimer, 1000);
            }
        };
        
        updateTimer();
    }

    // Modal Management
    async openAuctionModal(auctionId) {
        this.currentAuctionId = auctionId;
        const auction = this.auctions.get(auctionId);
        
        if (!auction) {
            this.showToast('Auction not found', 'error');
            return;
        }

        // Subscribe to auction updates via WebSocket
        this.sendWebSocketMessage({
            action: 'subscribe',
            listing_id: auction.listing_id
        });

        // Populate modal with auction data
        this.updateModalAuctionData(auction);
        
        // Show modal
        document.getElementById('auctionModal').classList.add('show');
        
        // Start modal timer
        this.startModalTimer();
        
        // Clear bid form
        document.getElementById('bidAmount').value = '';
        document.getElementById('bidMessage').innerHTML = '';
    }

    closeModal() {
        const modal = document.getElementById('auctionModal');
        modal.classList.remove('show');
        
        // Unsubscribe from auction updates
        if (this.currentAuctionId) {
            const auction = this.auctions.get(this.currentAuctionId);
            if (auction) {
                this.sendWebSocketMessage({
                    action: 'unsubscribe',
                    listing_id: auction.listing_id
                });
            }
        }
        
        this.currentAuctionId = null;
    }

    updateModalAuctionData(auction) {
        const listing = auction.listing || {};
        
        document.getElementById('modalTitle').textContent = listing.title || 'Auction Details';
        document.getElementById('auctionTitle').textContent = listing.title || 'Unknown Item';
        document.getElementById('auctionDescription').textContent = listing.description || 'No description available';
        document.getElementById('currentBid').textContent = `$${(auction.current_bid || listing.reserve_price || 0).toFixed(2)}`;
        document.getElementById('reservePrice').textContent = `$${(listing.reserve_price || 0).toFixed(2)}`;
        document.getElementById('bidCount').textContent = auction.bid_count || 0;
        
        // Set minimum bid amount
        const minBid = (auction.current_bid || listing.reserve_price || 0) + 1;
        const bidInput = document.getElementById('bidAmount');
        bidInput.min = minBid;
        bidInput.placeholder = `Minimum: $${minBid.toFixed(2)}`;
    }

    startModalTimer() {
        const updateModalTimer = () => {
            if (!this.currentAuctionId) return;
            
            const auction = this.auctions.get(this.currentAuctionId);
            if (!auction || auction.status !== 'active') {
                document.getElementById('timeRemaining').textContent = 'Ended';
                return;
            }
            
            const timeLeft = this.formatTimeRemaining(auction.end_time);
            document.getElementById('timeRemaining').textContent = timeLeft;
            
            if (new Date(auction.end_time) > new Date()) {
                setTimeout(updateModalTimer, 1000);
            }
        };
        
        updateModalTimer();
    }

    // Bidding
    async handleBid() {
        if (!this.currentAuctionId) {
            this.showToast('No auction selected', 'error');
            return;
        }

        const bidAmount = parseFloat(document.getElementById('bidAmount').value);
        if (isNaN(bidAmount) || bidAmount <= 0) {
            this.showBidMessage('Please enter a valid bid amount', 'error');
            return;
        }

        const auction = this.auctions.get(this.currentAuctionId);
        const minBid = (auction.current_bid || auction.listing?.reserve_price || 0) + 1;
        
        if (bidAmount < minBid) {
            this.showBidMessage(`Bid must be at least $${minBid.toFixed(2)}`, 'error');
            return;
        }

        // Send bid via WebSocket
        this.sendWebSocketMessage({
            action: 'bid',
            listing_id: auction.listing_id,
            amount: bidAmount
        });

        this.showBidMessage('Placing bid...', 'success');
        document.getElementById('bidAmount').value = '';
    }

    showBidMessage(message, type) {
        const messageElement = document.getElementById('bidMessage');
        messageElement.textContent = message;
        messageElement.className = `message ${type}`;
        
        setTimeout(() => {
            messageElement.innerHTML = '';
        }, 3000);
    }

    addBidToList(bidData) {
        const bidsList = document.getElementById('bidsList');
        const noBids = bidsList.querySelector('.no-bids');
        
        if (noBids) {
            noBids.remove();
        }
        
        const isOwnBid = bidData.user_id === this.currentUser?.username;
        const bidElement = document.createElement('div');
        bidElement.className = `bid-item ${isOwnBid ? 'own-bid' : ''}`;
        bidElement.innerHTML = `
            <span class="bid-amount">$${bidData.amount.toFixed(2)}</span>
            <span class="bid-user">${bidData.user_id}${isOwnBid ? ' (You)' : ''}</span>
        `;
        
        bidsList.insertBefore(bidElement, bidsList.firstChild);
        
        // Limit to 10 recent bids
        const bidItems = bidsList.querySelectorAll('.bid-item');
        if (bidItems.length > 10) {
            bidItems[bidItems.length - 1].remove();
        }
    }

    // Activity Feed
    addActivityMessage(message, type) {
        const feed = document.getElementById('activityFeed');
        const activityElement = document.createElement('div');
        activityElement.className = `activity-item ${type}`;
        
        const timestamp = new Date().toLocaleTimeString();
        activityElement.innerHTML = `
            <span class="timestamp">${timestamp}</span>
            <span class="message">${message}</span>
        `;
        
        feed.insertBefore(activityElement, feed.firstChild);
        
        // Limit to 50 activity items
        const items = feed.querySelectorAll('.activity-item');
        if (items.length > 50) {
            items[items.length - 1].remove();
        }
        
        // Scroll to top of feed
        feed.scrollTop = 0;
    }

    // Utility Functions
    formatTimeRemaining(endTime) {
        const now = new Date();
        const end = new Date(endTime);
        const diff = end - now;
        
        if (diff <= 0) {
            return 'Ended';
        }
        
        const hours = Math.floor(diff / (1000 * 60 * 60));
        const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = Math.floor((diff % (1000 * 60)) / 1000);
        
        if (hours > 0) {
            return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
        } else {
            return `${minutes}:${seconds.toString().padStart(2, '0')}`;
        }
    }

    showToast(message, type = 'success') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        
        container.appendChild(toast);
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 5000);
        
        // Remove on click
        toast.addEventListener('click', () => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        });
    }
}

// Initialize the application
let auctionApp;
document.addEventListener('DOMContentLoaded', () => {
    auctionApp = new AuctionApp();
});