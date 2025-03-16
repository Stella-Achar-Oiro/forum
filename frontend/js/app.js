// frontend/js/app.js
const App = {
    // Initialize the application
    async init() {
        try {
            // Check if user is logged in
            const user = await AuthService.init();
            
            if (user) {
                // Initialize WebSocket
                WebSocketService.init(user.id);
                
                // Initialize chat component
                ChatComponent.init();
                
                // Render home page
                this.renderHome();
            } else {
                // Render login page
                this.renderLogin();
            }
        } catch (error) {
            console.error('Initialization error:', error);
            this.renderLogin();
        }
    },
    
    // Render login page
    renderLogin() {
        const appContainer = document.getElementById('app');
        appContainer.innerHTML = '';
        
        const loginComponent = AuthComponent.renderLogin();
        appContainer.appendChild(loginComponent);
    },
    
    // Render registration page
    renderRegister() {
        const appContainer = document.getElementById('app');
        appContainer.innerHTML = '';
        
        const registerComponent = AuthComponent.renderRegister();
        appContainer.appendChild(registerComponent);
    },
    
    // Render home page with posts and chat
    async renderHome() {
        const appContainer = document.getElementById('app');
        appContainer.innerHTML = '';
        
        // Add navigation
        const navbar = NavigationComponent.renderNavbar();
        appContainer.appendChild(navbar);
        
        // Create main container
        const mainContainer = document.createElement('div');
        mainContainer.className = 'main-container';
        
        // Load and add posts component
        const postsComponent = await PostsComponent.renderPosts();
        mainContainer.appendChild(postsComponent);
        
        // Load and add chat sidebar
        const chatSidebar = await ChatComponent.renderChatSidebar();
        mainContainer.appendChild(chatSidebar);
        
        appContainer.appendChild(mainContainer);
    },

    // Render user profile
    async renderProfile(userId) {
        const appContainer = document.getElementById('app');
        appContainer.innerHTML = '';
        
        // Add navigation
        const navbar = NavigationComponent.renderNavbar();
        appContainer.appendChild(navbar);
        
        // Create main container
        const mainContainer = document.createElement('div');
        mainContainer.className = 'main-container';
        
        // Load and add profile component
        const profileComponent = await ProfileComponent.renderProfile(userId);
        mainContainer.appendChild(profileComponent);
        
        // Load and add chat sidebar
        const chatSidebar = await ChatComponent.renderChatSidebar();
        mainContainer.appendChild(chatSidebar);
        
        appContainer.appendChild(mainContainer);
    }
};

// Initialize the app when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    App.init();
});