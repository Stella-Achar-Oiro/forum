// AuthService.js - Handles all authentication-related API calls

class AuthService {
    // Helper method to get the store instance
    static getStore() {
        // Try window.Store first (exported as window.Store in Store.js)
        if (window.Store && typeof window.Store.get === 'function') {
            return window.Store;
        }
        // Try window.store as a fallback (created as window.store in Store.js)
        if (window.store && typeof window.store.get === 'function') {
            return window.store;
        }
        // If still not found, throw a readable error to help with debugging
        throw new Error('Store not initialized - check script loading order. Make sure Store.js is loaded before AuthService.js');
    }

    static async login(identifier, password) {
        console.log('AuthService.login called with identifier:', identifier);
        try {
            console.log('Sending login request to /api/public/login');
            const response = await fetch('/api/public/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ identifier, password })
            });

            console.log('Login response received, status:', response.status);
            
            if (!response.ok) {
                console.error('Login response not OK:', response.status, response.statusText);
                let errorData;
                try {
                    errorData = await response.json();
                    console.error('Error response data:', errorData);
                    throw new Error(errorData.message || `Login failed with status: ${response.status}`);
                } catch (jsonError) {
                    console.error('Failed to parse error response:', jsonError);
                    throw new Error(`Login failed with status: ${response.status}`);
                }
            }

            console.log('Parsing login response data');
            let data;
            try {
                data = await response.json();
                console.log('Login response data:', data);
            } catch (jsonError) {
                console.error('Failed to parse successful response:', jsonError);
                throw new Error('Failed to parse server response');
            }
            
            console.log('Setting user data in store');
            try {
                this.getStore().set('currentUser', data.data);
                console.log('User data set in store:', data.data);
            } catch (storeError) {
                console.error('Error setting user in store:', storeError);
                // Continue despite store error
            }
            
            return data.data;
        } catch (error) {
            console.error('Login process error:', error);
            throw error;
        }
    }

    static initiateGoogleAuth() {
        // Update to use the correct backend endpoint
        window.location.href = '/api/public/auth/google';
    }

    static initiateGithubAuth() {
        // Update to use the correct backend endpoint
        window.location.href = '/api/public/auth/github';
    }

    static async handleOAuthCallback(provider, code) {
        try {
            const response = await fetch(`/api/public/auth/${provider}/callback`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ code })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || `${provider} authentication failed`);
            }

            const data = await response.json();
            this.getStore().set('currentUser', data.data);
            return data.data;
        } catch (error) {
            console.error(`${provider} OAuth error:`, error);
            throw error;
        }
    }

    static async register(userData) {
        try {
            const response = await fetch('/api/public/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(userData)
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Registration failed');
            }

            const data = await response.json();
            // Store user data if login upon registration
            if (data.data && data.status === 'success') {
                this.getStore().set('currentUser', data.data);
            }
            return data;
        } catch (error) {
            console.error('Registration error:', error);
            throw error;
        }
    }

    static async logout() {
        try {
            const response = await fetch('/api/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Logout failed');
            }

            this.getStore().set('currentUser', null);
            return true;
        } catch (error) {
            console.error('Logout error:', error);
            throw error;
        }
    }

    static async getCurrentUser() {
        try {
            // Using try-catch to handle case when Store is not yet initialized
            try {
                const cachedUser = this.getStore().get('currentUser');
                if (cachedUser) {
                    return cachedUser;
                }
            } catch (error) {
                console.warn('Store not available yet, will attempt to fetch user from API');
                // Continue to API request if store isn't available
            }

            const response = await fetch('/api/user');
            
            if (!response.ok) {
                if (response.status === 401) {
                    try {
                        this.getStore().set('currentUser', null);
                    } catch (error) {
                        console.warn('Could not update store on 401 unauthorized');
                    }
                    return null;
                }
                throw new Error('Failed to get current user');
            }

            const data = await response.json();
            try {
                this.getStore().set('currentUser', data.data);
            } catch (error) {
                console.warn('Could not update store with user data');
            }
            return data.data;
        } catch (error) {
            console.error('Error getting current user:', error);
            try {
                this.getStore().set('currentUser', null);
            } catch (e) {
                console.warn('Could not reset store after error');
            }
            return null;
        }
    }
}

// Export the service
export default AuthService; 