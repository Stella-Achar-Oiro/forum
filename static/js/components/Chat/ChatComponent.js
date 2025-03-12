// ChatComponent.js - Chat conversation interface
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

class ChatComponent extends Component {
    constructor(props) {
        super(props);
        this.state = {
            loading: true,
            error: null,
            message: '',
            sending: false,
            loadingMore: false,
            hasMore: true,
            offset: 0,
            limit: 50,
            chat: null,
            userId: null,
            onlineUsers: new Set()
        };

        // Get user ID from URL params
        const pathParts = window.location.pathname.split('/');
        const userId = parseInt(pathParts[pathParts.length - 1]);
        if (!isNaN(userId)) {
            this.state.userId = userId;
        }

        // Subscribe to store changes
        this.unsubscribeChats = Store.subscribe('chats', (chats) => {
            if (chats && Array.isArray(chats) && this.state.userId) {
                const chat = chats.find(c => c.user && c.user.id === this.state.userId);
                this.setState({ chat });
            }
        });
        
        this.unsubscribeOnlineUsers = Store.subscribe('onlineUsers', (onlineUsers) => {
            this.setState({ onlineUsers: onlineUsers || new Set() });
        });

        this.messageListRef = null;
    }

    componentWillUnmount() {
        if (this.unsubscribeChats) this.unsubscribeChats();
        if (this.unsubscribeOnlineUsers) this.unsubscribeOnlineUsers();
    }

    async componentDidMount() {
        const { userId } = this.state;
        if (!userId) {
            this.setState({ error: "Invalid user ID", loading: false });
            return;
        }

        try {
            const messages = await ChatService.getMessages(userId);
            const chats = Store.get('chats') || [];
            const chat = chats.find(c => c.user && c.user.id === userId);
            
            this.setState({ 
                loading: false,
                chat
            });
            
            this.scrollToBottom();
        } catch (error) {
            this.setState({ error: error.message, loading: false });
        }
    }

    componentDidUpdate(prevProps, prevState) {
        const { chat } = this.state;
        const prevMessages = prevState.chat && prevState.chat.messages ? prevState.chat.messages.length : 0;
        const currentMessages = chat && chat.messages ? chat.messages.length : 0;
        
        if (currentMessages !== prevMessages) {
            this.scrollToBottom();
        }
    }

    scrollToBottom() {
        if (this.messageListRef) {
            this.messageListRef.scrollTop = this.messageListRef.scrollHeight;
        }
    }

    async handleLoadMore() {
        const { userId, offset, limit, loadingMore } = this.state;
        if (loadingMore || !userId) return;

        this.setState({ loadingMore: true });

        try {
            const messages = await ChatService.getMessages(userId, offset + limit, limit);
            this.setState(state => ({
                offset: state.offset + limit,
                hasMore: messages.length === limit
            }));
        } catch (error) {
            this.setState({ error: error.message });
        } finally {
            this.setState({ loadingMore: false });
        }
    }

    handleScroll = (e) => {
        const { hasMore, loadingMore } = this.state;
        if (!hasMore || loadingMore) return;

        const { scrollTop } = e.target;
        if (scrollTop === 0) {
            this.handleLoadMore();
        }
    };

    async handleSendMessage(e) {
        e.preventDefault();
        const { message, sending, userId } = this.state;
        if (!message.trim() || sending || !userId) return;

        this.setState({ sending: true });

        try {
            await ChatService.sendMessage(userId, message.trim());
            this.setState({ message: '' });
        } catch (error) {
            this.setState({ error: error.message });
        } finally {
            this.setState({ sending: false });
        }
    }

    renderMessage(message) {
        const currentUser = Store.get('currentUser');
        const isMine = currentUser && message.senderId === currentUser.id;

        return this.createEl('div', {
            className: `message ${isMine ? 'message-mine' : 'message-other'}`
        }, [
            this.createEl('div', { className: 'message-content' }, [
                this.createEl('div', { className: 'message-text' }, [message.content]),
                this.createEl('div', { className: 'message-time' }, [
                    new Date(message.createdAt).toLocaleTimeString([], { 
                        hour: '2-digit', 
                        minute: '2-digit' 
                    })
                ]),
                isMine && this.createEl('div', { 
                    className: `message-status ${message.isRead ? 'read' : 'sent'}`
                })
            ])
        ]);
    }

    renderContent() {
        const { chat, onlineUsers, loading, error, message, sending, loadingMore, userId } = this.state;

        if (loading) {
            return this.createEl('div', { className: 'chat-loading' }, ['Loading chat...']);
        }

        if (error) {
            return this.createEl('div', { className: 'chat-error' }, [
                'Failed to load chat: ',
                error
            ]);
        }

        if (!chat || !chat.user) {
            return this.createEl('div', { className: 'chat-not-found' }, [
                'Chat not found'
            ]);
        }

        const isOnline = onlineUsers.has(chat.user.id);
        const messages = chat.messages || [];

        return this.createEl('div', { className: 'chat' }, [
            this.createEl('div', { className: 'chat-header' }, [
                this.createEl('div', { className: 'chat-user-info' }, [
                    this.createEl('h3', {}, [`${chat.user.firstName || ''} ${chat.user.lastName || ''}`]),
                    this.createEl('div', { 
                        className: `status-indicator ${isOnline ? 'online' : 'offline'}`
                    }, [isOnline ? 'Online' : 'Offline'])
                ]),
                this.createEl('button', {
                    className: 'back-button',
                    onClick: () => this.navigate('/chat')
                }, ['Back to Messages'])
            ]),
            this.createEl('div', {
                className: 'message-list',
                onScroll: this.handleScroll,
                ref: el => this.messageListRef = el
            }, [
                loadingMore && this.createEl('div', { className: 'loading-more' }, [
                    'Loading more messages...'
                ]),
                messages.length > 0 
                    ? messages.map(msg => this.renderMessage(msg))
                    : this.createEl('div', { className: 'no-messages' }, ['No messages yet. Start a conversation!'])
            ]),
            this.createEl('form', {
                className: 'message-form',
                onSubmit: (e) => this.handleSendMessage(e)
            }, [
                this.createEl('input', {
                    type: 'text',
                    className: 'message-input',
                    placeholder: 'Type a message...',
                    value: message,
                    onInput: (e) => this.setState({ message: e.target.value }),
                    disabled: sending
                }),
                this.createEl('button', {
                    type: 'submit',
                    className: 'send-button',
                    disabled: sending || !message.trim()
                }, [sending ? 'Sending...' : 'Send'])
            ])
        ]);
    }
}

// Register component
ComponentRegistry.register('ChatComponent', ChatComponent);

// Export the component
export default ChatComponent; 