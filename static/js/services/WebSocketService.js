class WebSocketService {
    constructor() {
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectTimeout = 1000;
        this.connected = false;
    }

    connect() {
        if (this.ws) {
            this.ws.close();
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        try {
            this.ws = new WebSocket(`${protocol}//${window.location.host}/api/ws`);

            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.reconnectAttempts = 0;
                this.connected = true;
            };

            this.ws.onclose = () => {
                console.log('WebSocket disconnected');
                this.connected = false;
                this.handleReconnect();
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.connected = false;
            };

            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleMessage(message);
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };
        } catch (error) {
            console.error('Failed to connect WebSocket:', error);
            this.handleReconnect();
        }
    }

    handleReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            setTimeout(() => this.connect(), this.reconnectTimeout * this.reconnectAttempts);
        }
    }

    handleMessage(message) {
        switch (message.type) {
            case 'user_status':
                const { userId, isOnline } = message.payload;
                const onlineUsers = new Set(Store.get('onlineUsers') || new Set());
                
                if (isOnline) {
                    onlineUsers.add(userId);
                } else {
                    onlineUsers.delete(userId);
                }
                
                Store.set('onlineUsers', onlineUsers);
                break;

            case 'new_message':
                const { message: newMessage } = message.payload;
                const chats = [...(Store.get('chats') || [])];
                const chatIndex = chats.findIndex(c => 
                    (c.user && c.user.id === newMessage.senderId) || 
                    (c.user && c.user.id === newMessage.receiverId)
                );

                if (chatIndex !== -1) {
                    if (!chats[chatIndex].messages) {
                        chats[chatIndex].messages = [];
                    }
                    chats[chatIndex].messages.push(newMessage);
                    Store.set('chats', chats);
                }
                break;

            case 'message_read':
                const { messageId, chatId } = message.payload;
                const updatedChats = (Store.get('chats') || []).map(chat => {
                    if (chat.id === chatId && chat.messages) {
                        return {
                            ...chat,
                            messages: chat.messages.map(msg => 
                                msg.id === messageId ? { ...msg, isRead: true } : msg
                            )
                        };
                    }
                    return chat;
                });
                
                Store.set('chats', updatedChats);
                break;
                
            default:
                console.log('Unhandled WebSocket message type:', message.type);
        }
    }

    send(type, payload) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ type, payload }));
            return true;
        }
        console.warn('WebSocket not connected, cannot send message');
        return false;
    }

    isConnected() {
        return this.connected && this.ws && this.ws.readyState === WebSocket.OPEN;
    }
}

// Create global WebSocket instance
const WebSocket = new WebSocketService();

// Export the instance
window.WebSocket = WebSocket; 