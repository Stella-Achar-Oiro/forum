// frontend/js/components/chat.js
const ChatComponent = {
    activeChat: null,
    users: [],
    messages: {},
    typingUsers: {},
    loadMoreThrottled: null,
    
    // Initialize chat component
    init() {
        try {
            // Set up throttled version of loadMoreMessages
            this.loadMoreThrottled = throttle(this.loadMoreMessages.bind(this), 1000);
            
            // Register WebSocket handlers if WebSocketService is available
            if (WebSocketService && typeof WebSocketService.onMessage === 'function') {
                WebSocketService.onMessage(this.handleNewMessage.bind(this));
            }
            
            if (WebSocketService && typeof WebSocketService.onTyping === 'function') {
                WebSocketService.onTyping(this.handleTypingIndicator.bind(this));
            }
            
            if (WebSocketService && typeof WebSocketService.onOnlineStatus === 'function') {
                WebSocketService.onOnlineStatus(this.handleOnlineStatus.bind(this));
            }
            
            console.log("Chat component initialized successfully");
        } catch (error) {
            console.error("Error initializing chat component:", error);
        }
    },
    
    // Render the chat sidebar
    async renderChatSidebar() {
        const container = document.createElement('div');
        container.className = 'sidebar';
        
        try {
            // Load chat users
            this.users = await API.messages.getChats();
            
            container.innerHTML = `
                <h3>Chat</h3>
                <div class="chat-list">
                    ${this.renderUserList()}
                </div>
            `;
            
            // Add event listeners to chat items
            setTimeout(() => {
                const chatItems = container.querySelectorAll('.chat-list-item');
                chatItems.forEach(item => {
                    item.addEventListener('click', () => {
                        const userId = parseInt(item.dataset.userId);
                        this.setActiveChat(userId);
                    });
                });
            }, 0);
            
        } catch (error) {
            container.innerHTML = `
                <h3>Chat</h3>
                <div class="error-message">
                    <p>Error loading chats: ${error.message}</p>
                </div>
            `;
        }
        
        return container;
    },
    
    // Render the user list
    renderUserList() {
        if (!this.users || this.users.length === 0) {
            return '<p>No users to chat with.</p>';
        }
        
        return this.users.map(user => {
            const isActive = this.activeChat === user.id;
            const isOnline = user.online || false;
            const isTyping = this.typingUsers[user.id] || false;
            
            // frontend/js/components/chat.js (continued)
            return `
                <div class="chat-list-item ${isActive ? 'active' : ''}" data-user-id="${user.id}">
                    <div class="user-avatar">${user.nickname.charAt(0).toUpperCase()}</div>
                    <div class="user-info">
                        <div class="user-name">${user.nickname}</div>
                        ${isTyping ? '<div class="typing-status">typing...</div>' : ''}
                    </div>
                    <div class="online-indicator ${isOnline ? 'online' : 'offline'}"></div>
                </div>
            `;
        }).join('');
    },
    
    // Render the chat content
    renderChatContent() {
        const container = document.createElement('div');
        container.className = 'chat-content';
        container.id = 'chat-content';
        
        if (!this.activeChat) {
            container.innerHTML = `
                <div class="no-chat-selected">
                    <p>Select a chat to start messaging</p>
                </div>
            `;
            return container;
        }
        
        // Find user in chat list
        const activeUser = this.users.find(user => user.id === this.activeChat);
        if (!activeUser) {
            container.innerHTML = `
                <div class="no-chat-selected">
                    <p>User not found</p>
                </div>
            `;
            return container;
        }
        
        container.innerHTML = `
            <div class="chat-header">
                <div class="chat-user-info">
                    <img src="${activeUser.avatar || 'img/default-avatar.png'}" alt="${activeUser.nickname}" class="avatar">
                    <span class="chat-username">${activeUser.nickname}</span>
                </div>
            </div>
            <div class="messages-container" id="messages-container">
                <div id="messages-list" class="messages-list">
                    <div class="loading-messages">Loading messages...</div>
                </div>
                <div id="load-more-container" class="load-more-container" style="display: none;">
                    <button id="load-more-btn">Load More</button>
                </div>
            </div>
            <div class="chat-input">
                <div id="image-preview-container" class="image-preview-container" style="display: none;">
                    <img id="image-preview" src="" alt="Image Preview">
                    <button id="remove-image-btn" class="remove-image-btn">×</button>
                </div>
                <div class="input-container">
                    <textarea id="message-input" placeholder="Type a message..."></textarea>
                    <div class="chat-actions">
                        <label for="image-upload" class="image-upload-label">
                            <img src="img/image-icon.svg" alt="Upload Image" class="image-icon">
                        </label>
                        <input type="file" id="image-upload" accept="image/*" style="display: none;">
                        <button id="send-message-btn">Send</button>
                    </div>
                </div>
            </div>
        `;

        try {
            // Image upload handlers
            const imageUpload = document.getElementById('image-upload');
            const imagePreviewContainer = document.getElementById('image-preview-container');
            const imagePreview = document.getElementById('image-preview');
            const removeImageBtn = document.getElementById('remove-image-btn');
            
            if (imageUpload && imagePreviewContainer && imagePreview && removeImageBtn) {
                imageUpload.addEventListener('change', (e) => {
                    const file = e.target.files[0];
                    if (file) {
                        const reader = new FileReader();
                        reader.onload = function(e) {
                            imagePreview.src = e.target.result;
                            imagePreviewContainer.style.display = 'block';
                        }
                        reader.readAsDataURL(file);
                    }
                });
                
                removeImageBtn.addEventListener('click', () => {
                    imageUpload.value = '';
                    imagePreviewContainer.style.display = 'none';
                });
            }
        } catch (error) {
            console.error("Error setting up image upload handlers:", error);
        }
        
        return container;
    },
    
    // Render messages for the active chat
    renderMessages(messages) {
        if (!messages || messages.length === 0) {
            return '<p class="no-messages">No messages yet. Start the conversation!</p>';
        }
        
        const currentUserId = AuthService.user.id;
        
        return messages.map(message => {
            const isSent = message.senderId === currentUserId;
            const messageClass = isSent ? 'sent' : 'received';
            const userName = isSent ? 'You' : message.sender.nickname;
            
            let content = `<div class="message-content">${message.content}</div>`;
            
            // Add image if present
            if (message.imageUrl) {
                content += `
                    <div class="message-image">
                        <img src="${message.imageUrl}" alt="Image" class="chat-image" onclick="showImageFullscreen('${message.imageUrl}')">
                    </div>
                `;
            }
            
            return `
                <div class="message ${messageClass}">
                    ${content}
                    <div class="message-meta">
                        <span class="message-sender">${userName}</span>
                        <span class="message-time">${new Date(message.createdAt).toLocaleString()}</span>
                    </div>
                </div>
            `;
        }).join('');
    },
    
    // Set the active chat user
    async setActiveChat(userId) {
        this.activeChat = userId;
        
        // Update UI
        const chatContainer = document.getElementById('chat-container');
        if (!chatContainer) return;
        
        // Clear previous content
        const oldContent = document.getElementById('chat-content');
        if (oldContent) {
            chatContainer.removeChild(oldContent);
        }
        
        // Add new content
        const chatContent = this.renderChatContent();
        chatContainer.appendChild(chatContent);
        
        // If no active chat, we're done
        if (!userId) return;
        
        try {
            // Load messages
            const messages = await API.messages.getMessages(userId);
            
            // Update messages container
            const messagesContainer = document.getElementById('messages-list');
            if (!messagesContainer) return;
            
            // Store messages
            this.messages[userId] = messages || [];
            
            // Render messages
            if (messages && messages.length > 0) {
                messagesContainer.innerHTML = this.renderMessages(messages);
                
                // Scroll to bottom
                const container = document.getElementById('messages-container');
                if (container) {
                    container.scrollTop = container.scrollHeight;
                }
            } else {
                messagesContainer.innerHTML = '<div class="no-messages">No messages yet. Start the conversation!</div>';
            }
            
            // Add event listeners
            this.attachChatEventListeners();
            
        } catch (error) {
            console.log('Error loading messages:', error);
            const messagesContainer = document.getElementById('messages-list');
            if (messagesContainer) {
                messagesContainer.innerHTML = `<div class="error-message">Error loading messages. <button id="retry-btn">Retry</button></div>`;
                
                setTimeout(() => {
                    const retryBtn = document.getElementById('retry-btn');
                    if (retryBtn) {
                        retryBtn.addEventListener('click', () => {
                            this.setActiveChat(userId);
                        });
                    }
                }, 0);
            }
        }
    },
    
    // Attach event listeners for chat functionality
    attachChatEventListeners() {
        try {
            const messageInput = document.getElementById('message-input');
            const sendBtn = document.getElementById('send-message-btn');
            
            if (messageInput) {
                messageInput.addEventListener('keypress', (e) => {
                    if (e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault();
                        this.sendMessage();
                    }
                });
            }
            
            if (sendBtn) {
                sendBtn.addEventListener('click', () => {
                    this.sendMessage();
                });
            }
            
            const loadMoreBtn = document.getElementById('load-more-btn');
            if (loadMoreBtn) {
                loadMoreBtn.addEventListener('click', () => {
                    this.loadMoreMessages(this.activeChat);
                });
            }
        } catch (error) {
            console.error("Error attaching chat event listeners:", error);
        }
    },
    
    // Send a message
    async sendMessage() {
        if (!this.activeChat) return;
        
        const messageInput = document.getElementById('message-input');
        const imageUpload = document.getElementById('image-upload');
        const content = messageInput.value.trim();
        
        // Must have either text or image
        if (content === '' && !imageUpload.files.length) return;
        
        try {
            let imageUrl = '';
            
            // Upload image if present
            if (imageUpload.files.length) {
                const formData = new FormData();
                formData.append('image', imageUpload.files[0]);
                
                const response = await fetch('/api/upload-image', {
                    method: 'POST',
                    body: formData
                });
                
                if (!response.ok) {
                    throw new Error('Image upload failed');
                }
                
                const data = await response.json();
                imageUrl = data.url;
            }
            
            // Send via API
            const message = await API.messages.sendMessage(this.activeChat, content, imageUrl);
            
            // Send via WebSocket for real-time delivery
            WebSocketService.sendChatMessage(this.activeChat, content, imageUrl);
            
            // Update local messages
            if (!this.messages[this.activeChat]) {
                this.messages[this.activeChat] = [];
            }
            this.messages[this.activeChat].push(message);
            
            // Update UI
            const messagesContainer = document.getElementById('chat-messages');
            const noMessages = messagesContainer.querySelector('.no-messages');
            if (noMessages) {
                messagesContainer.removeChild(noMessages);
            }
            
            let messageContent = `<div class="message-content">${message.content}</div>`;
            if (message.imageUrl) {
                messageContent += `
                    <div class="message-image">
                        <img src="${message.imageUrl}" alt="Image" class="chat-image" onclick="showImageFullscreen('${message.imageUrl}')">
                    </div>
                `;
            }
            
            const messageHTML = `
                <div class="message sent">
                    ${messageContent}
                    <div class="message-meta">
                        <span class="message-sender">You</span>
                        <span class="message-time">${new Date(message.createdAt).toLocaleString()}</span>
                    </div>
                </div>
            `;
            
            messagesContainer.insertAdjacentHTML('beforeend', messageHTML);
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
            
            // Clear input and image
            messageInput.value = '';
            imageUpload.value = '';
            document.getElementById('image-preview-container').style.display = 'none';
            
            // Reset typing status
            WebSocketService.sendTypingStatus(this.activeChat, false);
            
        } catch (error) {
            alert('Error sending message: ' + error.message);
        }
    },
    
    // Load more messages (for pagination)
    async loadMoreMessages(userId) {
        if (!userId) return;
        
        const currentMessages = this.messages[userId] || [];
        const offset = currentMessages.length;
        
        try {
            const moreMessages = await API.messages.getMessages(userId, 10, offset);
            
            if (moreMessages.length === 0) {
                // No more messages
                const loadMoreBtn = document.getElementById('load-more-messages');
                if (loadMoreBtn) {
                    loadMoreBtn.textContent = 'No more messages';
                    loadMoreBtn.disabled = true;
                }
                return;
            }
            
            // Add to existing messages
            this.messages[userId] = [...moreMessages, ...currentMessages];
            
            // Update UI
            const messagesContainer = document.getElementById('chat-messages');
            const oldHeight = messagesContainer.scrollHeight;
            
            // Insert new messages at the beginning
            const messagesHTML = moreMessages.map(message => {
                const isSent = message.senderId === AuthService.user.id;
                const messageClass = isSent ? 'sent' : 'received';
                const userName = isSent ? 'You' : message.sender.nickname;
                
                return `
                <div class="message ${messageClass}">
                    <div class="message-content">${message.content}</div>
                    <div class="message-meta">
                        <span class="message-sender">${userName}</span>
                        <span class="message-time">${new Date(message.createdAt).toLocaleString()}</span>
                    </div>
                </div>
            `;
            }).join('');
            
            messagesContainer.insertAdjacentHTML('beforeend', messagesHTML);
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
            
        } catch (error) {
            console.error('Error loading more messages:', error);
        }
    },
    
    // Handle new message received
    handleNewMessage(messageData) {
        console.log('New message received:', messageData);
        // This is a stub implementation - will be fully implemented later
    },
    
    // Handle typing indicator
    handleTypingIndicator(typingData) {
        console.log('Typing indicator:', typingData);
        // This is a stub implementation - will be fully implemented later
    },
    
    // Handle online status update
    handleOnlineStatus(statusData) {
        console.log('Online status update:', statusData);
        // This is a stub implementation - will be fully implemented later
    },
};