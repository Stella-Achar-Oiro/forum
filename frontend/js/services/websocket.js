// frontend/js/services/websocket.js
const WebSocketService = {
    socket: null,
    messageHandlers: [],
    typingHandlers: [],
    onlineStatusHandlers: [],
    postHandlers: [],
    commentHandlers: [],
    reconnectInterval: null,
    messageQueue: [],
    processingQueue: false,
    isConnected: false,
    reconnectAttempts: 0,
    maxReconnectAttempts: 10,
    reconnectDelay: 1000, // start with 1 second delay
    
    // Initialize WebSocket connection
    init(userId) {
        this.userId = userId;
        this.connect();
        
        // Set up reconnection logic
        window.addEventListener('online', () => {
            if (!this.isConnected) {
                this.reconnectAttempts = 0;
                this.reconnectDelay = 1000;
                this.connect();
            }
        });
    },
    
    // Connect to WebSocket server
    connect() {
        // Close existing connection if any
        if (this.socket) {
            this.socket.close();
        }
        
        // Create new connection
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        this.socket = new WebSocket(`${protocol}//${window.location.host}/ws?userId=${this.userId}`);
        
        // Setup event handlers
        this.socket.onopen = () => {
            console.log('WebSocket connection established');
            this.isConnected = true;
            this.reconnectAttempts = 0;
            this.reconnectDelay = 1000;
            
            // Process any queued messages
            this.processMessageQueue();
            
            // Clear reconnect interval if active
            if (this.reconnectInterval) {
                clearInterval(this.reconnectInterval);
                this.reconnectInterval = null;
            }
        };
        
        this.socket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        };
        
        this.socket.onclose = () => {
            console.log('WebSocket connection closed');
            this.isConnected = false;
            
            // Set up exponential backoff reconnection
            if (this.reconnectAttempts < this.maxReconnectAttempts) {
                setTimeout(() => {
                    if (navigator.onLine) {
                        this.reconnectAttempts++;
                        this.reconnectDelay = Math.min(30000, this.reconnectDelay * 2); // Max 30 seconds
                        this.connect();
                    }
                }, this.reconnectDelay);
            }
        };
        
        this.socket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    },
    
    // Send a message over WebSocket
    send(type, payload) {
        const message = JSON.stringify({
            type,
            payload
        });
        
        if (!this.isConnected) {
            // Queue message for later
            this.messageQueue.push(message);
            return false;
        }
        
        try {
            this.socket.send(message);
            return true;
        } catch (error) {
            console.error('Error sending WebSocket message:', error);
            this.messageQueue.push(message);
            return false;
        }
    },
    
    // Process queued messages
    async processMessageQueue() {
        if (this.processingQueue || this.messageQueue.length === 0 || !this.isConnected) {
            return;
        }
        
        this.processingQueue = true;
        
        while (this.messageQueue.length > 0 && this.isConnected) {
            const message = this.messageQueue.shift();
            try {
                this.socket.send(message);
                // Small delay to prevent flooding
                await new Promise(resolve => setTimeout(resolve, 10));
            } catch (error) {
                console.error('Error sending queued message:', error);
                // Put message back in queue if sending fails
                this.messageQueue.unshift(message);
                break;
            }
        }
        
        this.processingQueue = false;
    },
    
    // Handle incoming WebSocket messages
    handleMessage(message) {
        switch (message.type) {
            case 'chat_message':
                this.messageHandlers.forEach(handler => handler(message.payload));
                break;
                
            case 'typing':
                this.typingHandlers.forEach(handler => handler(message.payload));
                break;
                
            case 'online_status':
                this.onlineStatusHandlers.forEach(handler => handler(message.payload));
                break;
                
            case 'new_post':
                this.postHandlers.forEach(handler => handler(message.payload));
                break;
                
            case 'new_comment':
                this.commentHandlers.forEach(handler => handler(message.payload));
                break;
        }
    },
    
    // Register message handler
    onMessage(handler) {
        this.messageHandlers.push(handler);
    },
    
    // Register typing handler
    onTyping(handler) {
        this.typingHandlers.push(handler);
    },
    
    // Register online status handler
    onOnlineStatus(handler) {
        this.onlineStatusHandlers.push(handler);
    },
    
    // Register new post handler
    onNewPost(handler) {
        this.postHandlers.push(handler);
    },
    
    // Register new comment handler
    onNewComment(handler) {
        this.commentHandlers.push(handler);
    },
    
    // Send a chat message
    sendChatMessage(receiverId, content, imageUrl = '') {
        return this.send('chat_message', {
            receiverId,
            content,
            imageUrl,
            senderId: AuthService.user.id,
            senderName: AuthService.user.nickname,
            createdAt: new Date()
        });
    },
    
    // Send typing status
    sendTypingStatus(receiverId, isTyping) {
        return this.send('typing', {
            senderId: AuthService.user.id,
            receiverId,
            isTyping
        });
    },
    
    // Send new post notification
    sendNewPostNotification(postId) {
        return this.send('new_post', {
            postId
        });
    },
    
    // Send new comment notification
    sendNewCommentNotification(postId, commentId) {
        return this.send('new_comment', {
            postId,
            commentId
        });
    }
};