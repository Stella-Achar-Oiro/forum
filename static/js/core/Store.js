// Store.js - Global state management
class Store {
    constructor(initialState = {}) {
        this.state = initialState;
        this.listeners = new Map(); // Changed to Map for key-based subscriptions
        this.keyListeners = new Map(); // For subscribing to specific keys
        console.log('Store constructor called with initial state:', initialState);
    }

    getState() {
        return this.state;
    }

    setState(newState) {
        console.log('Store.setState called with:', newState);
        this.state = { ...this.state, ...newState };
        this.notify();
    }

    // Get a specific key from state
    get(key) {
        console.log(`Store.get('${key}') called, current value:`, this.state[key]);
        return this.state[key];
    }

    // Set a specific key in state
    set(key, value) {
        console.log(`Store.set('${key}') called with value:`, value);
        const oldValue = this.state[key];
        
        // Ensure we're not causing unnecessary re-renders for non-changes
        if (JSON.stringify(oldValue) === JSON.stringify(value)) {
            console.log(`Store.set('${key}'): Value unchanged, skipping update`);
            return;
        }
        
        try {
            this.state = { ...this.state, [key]: value };
            console.log(`Store.set('${key}'): State updated, new state:`, this.state);
            
            // Notify specific key listeners
            if (this.keyListeners.has(key)) {
                const listeners = this.keyListeners.get(key);
                console.log(`Store.set('${key}'): Notifying ${listeners.size} key-specific listeners`);
                listeners.forEach(listener => {
                    try {
                        listener(value, oldValue);
                    } catch (e) {
                        console.error(`Error in key listener for '${key}':`, e);
                    }
                });
            }
            
            // Notify general listeners
            this.notify();
        } catch (e) {
            console.error(`Error in Store.set('${key}'):`, e);
        }
    }

    // Subscribe to all state changes
    subscribe(listenerOrKey, listener) {
        console.log('Store.subscribe called with:', listenerOrKey, !!listener);
        
        // If two parameters, it's a key-specific subscription
        if (typeof listener === 'function') {
            const key = listenerOrKey;
            if (!this.keyListeners.has(key)) {
                this.keyListeners.set(key, new Set());
            }
            this.keyListeners.get(key).add(listener);
            console.log(`Subscribed to key '${key}', total listeners: ${this.keyListeners.get(key).size}`);
            
            // Return unsubscribe function
            return () => {
                const listeners = this.keyListeners.get(key);
                if (listeners) {
                    listeners.delete(listener);
                    console.log(`Unsubscribed from key '${key}', remaining listeners: ${listeners.size}`);
                    if (listeners.size === 0) {
                        this.keyListeners.delete(key);
                    }
                }
            };
        }
        
        // One parameter - it's a general subscription
        const generalListener = listenerOrKey;
        this.listeners.set(generalListener, true);
        console.log(`Added general listener, total listeners: ${this.listeners.size}`);
        return () => {
            this.listeners.delete(generalListener);
            console.log(`Removed general listener, remaining listeners: ${this.listeners.size}`);
        };
    }

    notify() {
        console.log(`Store.notify: Notifying ${this.listeners.size} general listeners`);
        this.listeners.forEach((_, listener) => {
            try {
                listener(this.state);
            } catch (e) {
                console.error('Error in general listener:', e);
            }
        });
    }
    
    // Helper method to reset the store - useful for testing or logout
    reset() {
        console.log('Store.reset called');
        this.state = {
            currentUser: null,
            chats: [],
            currentChat: null,
            onlineUsers: new Set()
        };
        this.notify();
    }
}

// Create global store instance if it doesn't exist
if (!window.store) {
    console.log('Creating global store instance');
    window.store = new Store({
        currentUser: null,
        chats: [],
        currentChat: null,
        onlineUsers: new Set()
    });

    // Export the store for easier access (make it available globally)
    window.Store = window.store;

    // Log store methods for debugging
    console.log('Store initialized with methods:', {
        get: typeof window.Store.get === 'function',
        set: typeof window.Store.set === 'function',
        subscribe: typeof window.Store.subscribe === 'function',
        getState: typeof window.Store.getState === 'function',
        setState: typeof window.Store.setState === 'function'
    });
} else {
    console.log('Store already exists, not creating a new one');
    // Ensure global Store variable is consistent
    window.Store = window.store;
}

// Export the store instance
export default window.Store;

// NOTE: connectStore has been moved to the Component class.
// This avoids duplication and potential conflicts. 