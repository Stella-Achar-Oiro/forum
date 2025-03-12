// NavbarComponent.js - Main navigation bar component
import Component from '../../core/Component.js';
import ComponentRegistry from '../../core/ComponentRegistry.js';
import AuthService from '../../services/AuthService.js';

// Use the global Store if module import fails
let Store;
try {
    Store = (await import('../../core/Store.js')).default;
} catch (e) {
    console.warn('Failed to import Store module, falling back to global', e);
    Store = window.ForumCore?.Store || window.Store;
}

class NavbarComponent extends Component {
    constructor(props) {
        super(props);
        this.state = {
            loading: false,
            showUserMenu: false,
            currentUser: null,
            unreadNotifications: 0,
            unreadMessages: 0
        };

        // Try to get current user state safely
        try {
            const store = AuthService.getStore();
            if (store) {
                this.setState({ currentUser: store.get('currentUser') });
            }
        } catch (error) {
            console.error('Error initializing NavbarComponent:', error);
        }
    }

    componentDidMount() {
        // Try to connect to the store safely
        try {
            this.connectStore(
                (state) => ({ currentUser: state.currentUser }),
                (selectedState) => this.setState(selectedState)
            );
            
            // Listen for store changes
            this.storeSubscription = Store.subscribe('currentUser', (user) => {
                this.setState({ currentUser: user });
            });
            
            // Set up polling for notifications and messages (in a real app)
            this.startPollingUnreadCounts();
        } catch (error) {
            console.error('Error connecting NavbarComponent to store:', error);
        }
        
        // Attach event handlers after DOM is ready
        setTimeout(() => this.attachEventHandlers(), 100);
    }
    
    componentWillUnmount() {
        // Clean up subscriptions
        if (this.storeSubscription) {
            Store.unsubscribe(this.storeSubscription);
        }
        
        // Clear any polling intervals
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
        }
    }
    
    // Start polling for unread notifications and messages
    startPollingUnreadCounts() {
        // Poll every 30 seconds
        this.pollingInterval = setInterval(() => {
            if (this.state.currentUser) {
                // Fetch notification counts from API
                this.fetchUnreadCounts();
            }
        }, 30000);
        
        // Do initial fetch
        if (this.state.currentUser) {
            this.fetchUnreadCounts();
        }
    }
    
    // Fetch unread notifications and messages
    async fetchUnreadCounts() {
        try {
            // In a real implementation, these would be actual API calls
            const notificationsResponse = await fetch('/api/notifications/unread/count');
            if (notificationsResponse.ok) {
                const notificationData = await notificationsResponse.json();
                this.setState({ unreadNotifications: notificationData.count || 0 });
            }
            
            // For messages we would also make an API call
            // This is a placeholder for now
        } catch (error) {
            console.error('Error fetching unread counts:', error);
        }
    }

    // Attach event handlers to DOM elements
    attachEventHandlers() {
        // Find logout button by class
        const logoutBtn = document.querySelector('.logout-button');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', (e) => {
                e.preventDefault();
                this.handleLogout();
            });
        }
        
        // Find user dropdown
        const userDropdown = document.querySelector('.user-dropdown');
        if (userDropdown) {
            userDropdown.addEventListener('click', (e) => {
                e.preventDefault();
                this.toggleUserMenu();
            });
        }
    }

    async handleLogout() {
        this.setState({ loading: true });
        try {
            await AuthService.logout();
            // Navigate to login page after successful logout
            window.location.hash = '#/login';
        } catch (error) {
            console.error('Logout failed:', error);
            alert('Logout failed: ' + error.message);
        } finally {
            this.setState({ loading: false, showUserMenu: false });
        }
    }

    toggleUserMenu() {
        this.setState({ showUserMenu: !this.state.showUserMenu });
    }

    renderContent() {
        const { currentUser, loading, showUserMenu, unreadNotifications, unreadMessages } = this.state;

        if (!currentUser) {
            return `
                <nav class="navbar">
                    <div class="navbar-brand">
                        <a href="#/" class="logo">Forum</a>
                    </div>
                    <div class="navbar-menu">
                        <a href="#/login" class="navbar-item">Login</a>
                        <a href="#/register" class="navbar-item btn-primary">Sign Up</a>
                    </div>
                </nav>
            `;
        }

        // User is logged in - show full navbar with logout button
        return `
            <nav class="navbar">
                <div class="navbar-brand">
                    <a href="#/" class="logo">Forum</a>
                </div>
                <div class="navbar-links">
                    <a href="#/" class="navbar-item">Home</a>
                    <a href="#/posts" class="navbar-item">Posts</a>
                    <a href="#/chat" class="navbar-item">
                        Messages
                        ${unreadMessages > 0 ? `<span class="badge">${unreadMessages}</span>` : ''}
                    </a>
                </div>
                <div class="navbar-menu">
                    <div class="user-dropdown">
                        <div class="navbar-user">
                            <span class="user-avatar">
                                ${currentUser && currentUser.nickname ? currentUser.nickname.substring(0, 1).toUpperCase() : '?'}
                            </span>
                            <span class="user-name">
                                ${currentUser && currentUser.nickname ? currentUser.nickname : 'User'}
                            </span>
                            <span class="dropdown-icon">▼</span>
                        </div>
                        ${showUserMenu ? `
                            <div class="user-dropdown-menu">
                                <a href="#/profile" class="dropdown-item">Profile</a>
                                <a href="#/settings" class="dropdown-item">Settings</a>
                                <div class="dropdown-divider"></div>
                                <button class="dropdown-item logout-button" ${loading ? 'disabled' : ''}>
                                    ${loading ? 'Logging out...' : 'Logout'}
                                </button>
                            </div>
                        ` : ''}
                    </div>
                </div>
                
                <!-- Always visible logout button for accessibility -->
                <div class="navbar-logout">
                    <button class="btn btn-outline logout-button" ${loading ? 'disabled' : ''}>
                        ${loading ? 'Logging out...' : 'Logout'}
                    </button>
                </div>
            </nav>
        `;
    }
}

// Register component
ComponentRegistry.register('NavbarComponent', NavbarComponent);

// Export the component
export default NavbarComponent; 