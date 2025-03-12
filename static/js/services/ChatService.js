// ChatService.js - Handles all chat-related API calls
import Store from '../../core/Store.js';

class ChatService {
    static async getMessageUsers() {
        try {
            const response = await fetch('/api/messages/users');
            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Failed to get message users');
            }

            const data = await response.json();
            Store.set('chats', data.data || []);
            return data.data || [];
        } catch (error) {
            console.error('Error getting message users:', error);
            return [];
        }
    }

    static async getMessages(userId, offset = 0, limit = 50) {
        try {
            const response = await fetch(`/api/messages/users/${userId}?offset=${offset}&limit=${limit}`);
            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Failed to get messages');
            }

            const data = await response.json();
            const chats = Store.get('chats') || [];
            const chatIndex = chats.findIndex(c => c.user.id === userId);

            if (chatIndex !== -1) {
                const updatedChats = [...chats];
                updatedChats[chatIndex].messages = data.data || [];
                Store.set('chats', updatedChats);
            }

            return data.data || [];
        } catch (error) {
            console.error('Error getting messages:', error);
            return [];
        }
    }

    static async sendMessage(userId, content) {
        try {
            const response = await fetch(`/api/messages/users/${userId}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ content })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Failed to send message');
            }

            const data = await response.json();
            const chats = Store.get('chats') || [];
            const chatIndex = chats.findIndex(c => c.user.id === userId);

            if (chatIndex !== -1) {
                const updatedChats = [...chats];
                updatedChats[chatIndex].messages = data.data || [];
                Store.set('chats', updatedChats);
            }

            return data.data || [];
        } catch (error) {
            console.error('Error sending message:', error);
            throw error;
        }
    }

    static markMessageAsRead(messageId, chatId) {
        try {
            // When WebSocket is implemented
            // window.ws.send('mark_read', { messageId, chatId });
            console.log('Marking message as read:', messageId, 'in chat:', chatId);
        } catch (error) {
            console.error('Error marking message as read:', error);
        }
    }
}

// Export the service
export default ChatService; 