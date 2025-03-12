// AuthGuard.js - Protects routes that require authentication

class AuthGuard {
    constructor() {
        this.isAuthenticated = false;
        this.isCheckingAuth = true;
        this.user = null;
        this.authCallbacks = [];
        this.checkAuthentication();

        // Check authentication status periodically to handle session expiration
        setInterval(() => {
            this.checkAuthentication();
        }, 5 * 60 * 1000); // Check every 5 minutes
    }

    // Check if the user is authenticated by calling the API
    async checkAuthentication() {
        try {
            this.isCheckingAuth = true;
            const response = await fetch('/api/user', {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                credentials: 'include'
            });

            if (response.ok) {
                // First check the content type to avoid JSON parsing errors
                const contentType = response.headers.get('content-type');
                if (contentType && contentType.includes('application/json')) {
                    const userData = await response.json();
                    this.isAuthenticated = true;
                    this.user = userData;
                } else {
                    console.error('Expected JSON response but got:', contentType);
                    this.isAuthenticated = false;
                    this.user = null;
                }
            } else {
                this.isAuthenticated = false;
                this.user = null;
                
                // Try to log the error response
                try {
                    const contentType = response.headers.get('content-type');
                    if (contentType && contentType.includes('application/json')) {
                        const errorData = await response.json();
                        console.error('Auth check failed with error:', errorData);
                    } else {
                        console.error('Auth check failed with status:', response.status);
                    }
                } catch (parseError) {
                    console.error('Could not parse error response:', parseError);
                }
            }
        } catch (error) {
            console.error('Authentication check failed:', error);
            this.isAuthenticated = false;
            this.user = null;
        } finally {
            this.isCheckingAuth = false;
            this.notifyCallbacks();
        }
    }

    // Get the current authentication status
    getAuthStatus() {
        return {
            isAuthenticated: this.isAuthenticated,
            isCheckingAuth: this.isCheckingAuth,
            user: this.user
        };
    }

    // Register a callback to be notified when auth status changes
    registerCallback(callback) {
        this.authCallbacks.push(callback);
        return () => {
            this.authCallbacks = this.authCallbacks.filter(cb => cb !== callback);
        };
    }

    // Notify all registered callbacks of auth status changes
    notifyCallbacks() {
        const authStatus = this.getAuthStatus();
        this.authCallbacks.forEach(callback => {
            try {
                callback(authStatus);
            } catch (err) {
                console.error('Error in auth callback:', err);
            }
        });
    }

    // Redirect to login if not authenticated
    guardRoute() {
        if (this.isCheckingAuth) {
            // Show loading state while checking
            return { 
                allowed: false, 
                loading: true 
            };
        }

        if (!this.isAuthenticated) {
            // Redirect to login page
            window.location.href = window.location.origin + '/#/login';
            return { 
                allowed: false, 
                loading: false 
            };
        }

        return { 
            allowed: true, 
            loading: false 
        };
    }

    // Get authentication token for CSRF protection
    async getCSRFToken() {
        const cookies = document.cookie.split(';');
        for (let i = 0; i < cookies.length; i++) {
            const cookie = cookies[i].trim();
            if (cookie.startsWith('forum_csrf_token=')) {
                return cookie.substring('forum_csrf_token='.length, cookie.length);
            }
        }
        return null;
    }

    // Log the user out
    async logout() {
        try {
            const csrfToken = await this.getCSRFToken();
            
            const response = await fetch('/api/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken || ''
                },
                credentials: 'include'
            });

            if (response.ok) {
                this.isAuthenticated = false;
                this.user = null;
                this.notifyCallbacks();
                window.location.href = window.location.origin + '/#/login';
                return true;
            } else {
                console.error('Logout failed:', await response.text());
                return false;
            }
        } catch (error) {
            console.error('Logout error:', error);
            return false;
        }
    }
}

// Create and export a singleton instance
const authGuard = new AuthGuard();
export default authGuard; 