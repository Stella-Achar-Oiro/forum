// ChatListComponent.js - Displays a list of chat conversations
import Component from '../../core/Component.js';
import ComponentRegistry from '../../core/ComponentRegistry.js';
import ChatService from '../../services/ChatService.js';

// Use the global Store if module import fails
let Store;
try {
    Store = (await import('../../core/Store.js')).default;
} catch (e) {
    console.warn('Failed to import Store module, falling back to global', e);
    Store = window.ForumCore?.Store || window.Store;
}

class ChatListComponent extends Component {
    constructor(props) {
        super(props);
        this.state = {
            loading: true,
            error: null,
            chats: [],
            onlineUsers: new Set()
        };

        // Subscribe to store changes
        this.unsubscribeChats = Store.subscribe('chats', (chats) => {
            this.setState({ chats: chats || [] });
        });
        
        this.unsubscribeOnlineUsers = Store.subscribe('onlineUsers', (onlineUsers) => {
            this.setState({ onlineUsers: onlineUsers || new Set() });
        });
    }

    componentWillUnmount() {
        if (this.unsubscribeChats) this.unsubscribeChats();
        if (this.unsubscribeOnlineUsers) this.unsubscribeOnlineUsers();
    }

    async componentDidMount() {
        try {
            const chats = await ChatService.getMessageUsers();
            this.setState({ loading: false, chats });
        } catch (error) {
            this.setState({ error: error.message, loading: false });
        }
    }

    formatLastMessageTime(timestamp) {
        const date = new Date(timestamp);
        const now = new Date();
        const diff = now - date;
        const days = Math.floor(diff / (1000 * 60 * 60 * 24));

        if (days > 7) {
            return date.toLocaleDateString();
        } else if (days > 0) {
            return `${days}d ago`;
        } else {
            const hours = Math.floor(diff / (1000 * 60 * 60));
            if (hours > 0) {
                return `${hours}h ago`;
            } else {
                const minutes = Math.floor(diff / (1000 * 60));
                return minutes > 0 ? `${minutes}m ago` : 'Just now';
            }
        }
    }

    renderChatItem(chat) {
        const { onlineUsers } = this.state;
        const isOnline = chat.user && onlineUsers.has(chat.user.id);
        const lastMessage = chat.lastMessage;

        if (!chat.user) {
            return null;
        }

        return this.createEl('div', {
            className: `chat-item ${chat.unreadCount > 0 ? 'unread' : ''}`,
            onClick: () => this.navigate(`/chat/${chat.user.id}`)
        }, [
            this.createEl('div', { className: 'chat-item-avatar' }, [
                this.createEl('div', { 
                    className: `status-indicator ${isOnline ? 'online' : 'offline'}`
                })
            ]),
            this.createEl('div', { className: 'chat-item-content' }, [
                this.createEl('div', { className: 'chat-item-header' }, [
                    this.createEl('h4', { className: 'chat-item-name' }, [
                        `${chat.user.firstName || ''} ${chat.user.lastName || ''}`
                    ]),
                    lastMessage && this.createEl('span', { className: 'chat-item-time' }, [
                        this.formatLastMessageTime(lastMessage.createdAt)
                    ])
                ]),
                this.createEl('div', { className: 'chat-item-message' }, [
                    lastMessage ? lastMessage.content : 'No messages yet'
                ]),
                chat.unreadCount > 0 && this.createEl('div', { className: 'unread-badge' }, [
                    chat.unreadCount.toString()
                ])
            ])
        ]);
    }

    renderContent() {
        const { chats, loading, error } = this.state;

        if (loading) {
            return this.createEl('div', { className: 'chat-list-loading' }, ['Loading chats...']);
        }

        if (error) {
            return this.createEl('div', { className: 'chat-list-error' }, [
                'Failed to load chats: ',
                error
            ]);
        }

        if (!chats || chats.length === 0) {
            return this.createEl('div', { className: 'chat-list-empty' }, [
                'No conversations yet'
            ]);
        }

        return this.createEl('div', { className: 'chat-list' }, [
            this.createEl('div', { className: 'chat-list-header' }, [
                this.createEl('h2', {}, ['Messages'])
            ]),
            this.createEl('div', { className: 'chat-list-content' }, 
                chats.map(chat => this.renderChatItem(chat)).filter(Boolean)
            )
        ]);
    }
}

// Register component
ComponentRegistry.register('ChatListComponent', ChatListComponent);

// Export the component
export default ChatListComponent; 