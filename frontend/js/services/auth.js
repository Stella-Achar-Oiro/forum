// frontend/js/services/auth.js
const AuthService = {
    user: null,
    
    // Initialize auth state
    async init() {
        try {
            console.log("Initializing auth service...");
            this.user = await API.auth.getCurrentUser();
            console.log("User authenticated:", this.user);
            return this.user;
        } catch (error) {
            // If it's just a 401, that's expected before login
            if (error.message && error.message.includes('401')) {
                console.log('No active session found, proceeding to login page');
            } else {
                console.error('Auth initialization error:', error);
            }
            this.user = null;
            return null;
        }
    },
    
    // Register a new user
    async register(userData) {
        console.log("Registering new user:", userData.nickname);
        const result = await API.auth.register(userData);
        this.user = result.user;
        console.log("Registration successful:", this.user);
        return result;
    },
    
    // Login a user
    async login(credentials) {
        console.log("Logging in user:", credentials.identifier);
        const result = await API.auth.login(credentials);
        this.user = result.user;
        console.log("Login successful:", this.user);
        return result;
    },
    
    // Logout the current user
    async logout() {
        await API.auth.logout();
        this.user = null;
        
        // Redirect to login page
        window.location.href = '/';
    },
    
    // Check if the user is logged in
    isLoggedIn() {
        return !!this.user;
    }
};