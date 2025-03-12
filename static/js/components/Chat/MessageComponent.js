// MessageComponent.js - Individual chat message component
import Component from '../../core/Component.js';
import ComponentRegistry from '../../core/ComponentRegistry.js';

// Use the global Store if module import fails
let Store;
try {
    Store = (await import('../../core/Store.js')).default;
} catch (e) {
    console.warn('Failed to import Store module, falling back to global', e);
    Store = window.ForumCore?.Store || window.Store;
}

class MessageComponent extends Component {
    constructor(props) {
        super(props);
        this.state = {
            message: props.message || {}
        };
    }

    render() {
        const { message } = this.state;
        const currentUser = Store?.get('currentUser');
        const isOwnMessage = message.senderId === currentUser?.id;

        return `
            <div class="message ${isOwnMessage ? 'message-own' : 'message-other'}">
                <div class="message-content">
                    <div class="message-text">${this.escapeHtml(message.content || '')}</div>
                    <div class="message-time">${this.formatTime(message.createdAt)}</div>
                </div>
            </div>
        `;
    }

    escapeHtml(unsafe) {
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }

    formatTime(timestamp) {
        if (!timestamp) return '';
        const date = new Date(timestamp);
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
}

// Register component
ComponentRegistry.register('MessageComponent', MessageComponent);

// Export the component
export default MessageComponent; 