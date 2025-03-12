// app.js - Main application entry point
import Component from './core/Component.js';
import Router from './core/Router.js';
import Store from './core/Store.js';
import authGuard from './components/Auth/AuthGuard.js';
import PostsComponent from './components/Posts/PostsComponent.js';
import PostDetailComponent from './components/Posts/PostDetailComponent.js';
import CreatePostComponent from './components/Posts/CreatePostComponent.js';
import LoginComponent from './components/Auth/LoginComponent.js';
import RegisterComponent from './components/Auth/RegisterComponent.js';
import ProfileComponent from './components/User/ProfileComponent.js';

// First, prevent duplicate initialization
if (window.appInitialized) {
    console.log('App already initialized, preventing duplicate initialization');
} else {
    window.appInitialized = true;

    // Auth configuration
    const AUTH_CONFIG = {
        sessionTimeoutMinutes: 30, // Session timeout after 30 minutes of inactivity
        publicRoutes: ['/login', '/register', '/forgot-password', '/reset-password'],
        loginRoute: '/login',
        defaultRoute: '/'
    };

    // Define routes with authentication requirements
    const routes = [
        {
            path: '/',
            component: class HomeComponent extends Component {
                renderContent() {
                    return '<div class="home-container"><h1>Welcome to Forum</h1><p>This is the home page of the application.</p></div>';
                }
            },
            requiresAuth: true
        },
        {
            path: '/login',
            component: LoginComponent,
            requiresAuth: false
        },
        {
            path: '/register',
            component: RegisterComponent,
            requiresAuth: false
        },
        {
            path: '/posts',
            component: PostsComponent,
            requiresAuth: false
        },
        {
            path: '/posts/new',
            component: CreatePostComponent,
            requiresAuth: true
        },
        {
            path: '/posts/:id',
            component: PostDetailComponent,
            requiresAuth: false
        },
        {
            path: '/profile',
            component: ProfileComponent,
            requiresAuth: true
        },
        {
            path: '/forgot-password',
            component: class ForgotPasswordComponent extends Component {
                renderContent() {
                    return '<div class="auth-container"><h1>Forgot Password</h1><p>Reset password functionality coming soon.</p><a href="#/login">Back to Login</a></div>';
                }
            },
            requiresAuth: false
        },
        {
            path: '*', // Fallback route
            component: class NotFoundComponent extends Component {
                renderContent() {
                    return '<div class="not-found"><h1>404 - Page Not Found</h1><p>The page you are looking for does not exist.</p><a href="#/">Go to Home</a></div>';
                }
            },
            requiresAuth: false
        }
    ];

    // Make routes available globally
    window.routes = routes;

    class App extends Component {
        constructor(props) {
            super(props);
            this.state = {
                loading: true,
                error: null,
                currentUser: null
            };
            console.log('App instance created');
            
            // Subscribe to auth changes
            this.unsubscribeAuth = authGuard.registerCallback(this.handleAuthChange.bind(this));
        }
        
        // Handle changes in authentication status
        handleAuthChange(authStatus) {
            this.setState({ 
                loading: authStatus.isCheckingAuth,
                currentUser: authStatus.user
            });
            
            // Update user in the global store
            if (window.ForumCore && window.ForumCore.Store) {
                window.ForumCore.Store.set('currentUser', authStatus.user);
            } else if (window.Store) {
                window.Store.set('currentUser', authStatus.user);
            } else {
                console.error('Store not available for setting currentUser');
            }
            
            // Refresh UI components that depend on auth status
            this.refreshAuthDependentComponents();
        }
        
        // Refresh components that depend on authentication status
        refreshAuthDependentComponents() {
            // Update navigation UI
            const navbar = document.querySelector('.navbar-component');
            if (navbar && navbar.instance) {
                navbar.instance.setState({ currentUser: this.state.currentUser });
                navbar.instance.render();
            }
        }

        async componentDidMount() {
            try {
                console.log('App componentDidMount started');
                
                // First ensure Store is available
                await this.ensureStoreAvailable();
                
                // Initialize router with authentication check
                this.initializeRouter();
                
                // Check authentication status
                await authGuard.checkAuthentication();
                
            } catch (error) {
                console.error('App initialization error:', error);
                this.setState({ loading: false, error: error.message });
            }
        }
        
        componentWillUnmount() {
            // Unsubscribe from auth changes when component unmounts
            if (this.unsubscribeAuth) {
                this.unsubscribeAuth();
            }
        }
        
        // Initialize router with authentication guard
        initializeRouter() {
            console.log("Initializing router with routes", routes);
            const router = Router.init(routes);
            
            // Handle hash-based navigation
            window.addEventListener('hashchange', () => {
                const hash = window.location.hash;
                console.log('Hash change detected:', hash);
                if (hash && hash.startsWith('#/')) {
                    const path = hash.substring(1); // Remove the # character
                    console.log('Navigating to hash path:', path);
                    this.navigateToPath(path);
                }
            });
            
            // Override the handleRoute method to add authentication guard
            const originalHandleRoute = router.handleRoute.bind(router);
            router.handleRoute = () => {
                // Check if we're already on a direct server path
                const path = window.location.pathname;
                const directServerPaths = ['/register', '/login', '/posts', '/profile'];
                
                console.log('Current location:', window.location.href);
                console.log('Routing to path:', path);
                
                // If we're already on a direct server path, redirect to hash-based equivalent
                if (directServerPaths.includes(path) && !window.location.hash) {
                    console.log('Detected direct server path, redirecting to hash-based route');
                    window.location.href = '/#' + path;
                    return;
                }
                
                // Check if this is a hash-based route
                if (window.location.hash && window.location.hash.startsWith('#/')) {
                    const hashPath = window.location.hash.substring(1);
                    console.log('Hash-based route detected:', hashPath);
                    this.navigateToPath(hashPath);
                    return;
                }
                
                // Otherwise, handle the path normally
                this.navigateToPath(path);
            };
            
            // Set up navigation helper on the router
            router.navigateTo = (path) => {
                console.log('Router navigateTo called with path:', path);
                this.navigateToPath(path);
            };
            
            // Store router reference globally
            window.appRouter = router;
            
            // Initial route handling
            router.handleRoute();
        }
        
        // Handle navigation with auth checks
        navigateToPath(path) {
            console.log(`Navigating to: ${path}`);
            
            // Find matching route
            const route = routes.find(r => {
                if (r.path === '*') return false; // Skip wildcard route for now
                return path === r.path || path.startsWith(`${r.path}/`);
            }) || routes.find(r => r.path === '*'); // Fallback to wildcard route
            
            if (!route) {
                console.error(`No route found for path: ${path}`);
                return;
            }
            
            // Check authentication for protected routes
            if (route.requiresAuth) {
                const guardResult = authGuard.guardRoute();
                
                if (guardResult.loading) {
                    // Show loading indicator while checking auth
                    this.showLoadingIndicator();
                    return;
                }
                
                if (!guardResult.allowed) {
                    // Already redirected in guardRoute
                    return;
                }
            }
            
            // Mount the component for this route
            this.mountRouteComponent(route, path);
        }
        
        // Show loading indicator
        showLoadingIndicator() {
            const appContainer = document.getElementById('app');
            if (appContainer) {
                appContainer.innerHTML = '<div class="loading-indicator"><div class="spinner"></div><p>Loading...</p></div>';
            }
        }
        
        // Mount a component for a route
        mountRouteComponent(route, path) {
            try {
                console.log(`Mounting component for route: ${route.path}`);
                
                // Create an instance of the component
                const componentInstance = new route.component({
                    path,
                    params: this.extractRouteParams(route.path, path),
                    user: this.state.currentUser
                });
                
                // Mount it to the app container
                const appContainer = document.getElementById('app');
                if (appContainer) {
                    // Clear previous content
                    appContainer.innerHTML = '';
                    
                    // Create wrapper element for the component
                    const componentElement = document.createElement('div');
                    componentElement.className = `route-component ${route.path.substring(1) || 'home'}-component`;
                    
                    // Store component instance on the element for future reference
                    componentElement.instance = componentInstance;
                    
                    // Add to DOM
                    appContainer.appendChild(componentElement);
                    
                    // Render the component
                    componentInstance.mount(componentElement);
                    console.log(`Component for ${route.path} mounted successfully`);
                } else {
                    console.error('App container not found');
                }
            } catch (error) {
                console.error(`Error mounting component for ${route.path}:`, error);
                this.showErrorMessage(`Failed to load page: ${error.message}`);
            }
        }
        
        // Extract route parameters from path
        extractRouteParams(routePath, actualPath) {
            const params = {};
            const routeParts = routePath.split('/');
            const actualParts = actualPath.split('/');
            
            routeParts.forEach((part, index) => {
                if (part.startsWith(':')) {
                    const paramName = part.substring(1);
                    params[paramName] = actualParts[index] || '';
                }
            });
            
            return params;
        }
        
        // Show error message
        showErrorMessage(message) {
            const appContainer = document.getElementById('app');
            if (appContainer) {
                appContainer.innerHTML = `<div class="error-message"><h3>Error</h3><p>${message}</p></div>`;
            }
        }

        // Ensure the Store is available
        async ensureStoreAvailable() {
            return new Promise((resolve, reject) => {
                if (window.Store) {
                    resolve();
                    return;
                }
                
                // Simple implementation of a store if not already available
                class SimpleStore {
                    constructor() {
                        this.state = {};
                        this.listeners = {};
                    }
                    
                    get(key) {
                        return this.state[key];
                    }
                    
                    set(key, value) {
                        this.state[key] = value;
                        if (this.listeners[key]) {
                            this.listeners[key].forEach(listener => listener(value));
                        }
                    }
                    
                    subscribe(key, listener) {
                        if (!this.listeners[key]) {
                            this.listeners[key] = [];
                        }
                        this.listeners[key].push(listener);
                        
                        return () => {
                            this.listeners[key] = this.listeners[key].filter(l => l !== listener);
                        };
                    }
                }
                
                window.Store = new SimpleStore();
                resolve();
            });
        }

        renderContent() {
            if (this.state.loading) {
                return `
                    <div class="app-loading">
                        <div class="spinner"></div>
                        <p>Loading application...</p>
                    </div>
                `;
            }
            
            if (this.state.error) {
                return `
                    <div class="app-error">
                        <h2>Application Error</h2>
                        <p>${this.state.error}</p>
                        <button class="btn btn-primary refresh-btn">Refresh Page</button>
                    </div>
                `;
            }
            
            return `
                <div class="app-container">
                    <div id="app-content"></div>
                </div>
            `;
        }
        
        // Helper: Wait for an element to appear in the DOM
        async waitForElement(selector, timeoutMs = 2000) {
            return new Promise((resolve, reject) => {
                if (document.querySelector(selector)) {
                    return resolve(document.querySelector(selector));
                }
                
                const observer = new MutationObserver(mutations => {
                    if (document.querySelector(selector)) {
                        observer.disconnect();
                        resolve(document.querySelector(selector));
                    }
                });
                
                observer.observe(document.body, {
                    childList: true,
                    subtree: true
                });
                
                // Set timeout
                setTimeout(() => {
                    observer.disconnect();
                    reject(new Error(`Timeout waiting for element: ${selector}`));
                }, timeoutMs);
            });
        }

        afterRender() {
            // Handle refresh button click
            if (this._container) {
                const refreshBtn = this._container.querySelector('.refresh-btn');
                if (refreshBtn) {
                    refreshBtn.addEventListener('click', () => {
                        window.location.reload();
                    });
                }
            }
            
            // Initialize the router if we're past the loading state
            if (!this.state.loading && !this.state.error) {
                // Route to appropriate component based on URL
                if (window.appRouter) {
                    window.appRouter.handleRoute();
                }
            }
        }
    }

    // Add test mode for development
    function setupTestMode() {
        if (window.location.search.includes('test=true')) {
            console.log('Application running in test mode');
            
            // Add test controls
            const testPanel = document.createElement('div');
            testPanel.className = 'test-panel';
            testPanel.innerHTML = `
                <div class="test-panel-header">Test Controls</div>
                <div class="test-panel-body">
                    <button id="test-login">Simulate Login</button>
                    <button id="test-logout">Simulate Logout</button>
                    <button id="test-error">Simulate Error</button>
                </div>
            `;
            document.body.appendChild(testPanel);
            
            // Test button handlers
            document.getElementById('test-login').addEventListener('click', () => {
                Store.set('currentUser', { id: 1, name: 'Test User' });
                alert('Simulated login successful');
            });
            
            document.getElementById('test-logout').addEventListener('click', () => {
                Store.set('currentUser', null);
                alert('Logged out successfully');
            });
            
            document.getElementById('test-error').addEventListener('click', () => {
                throw new Error('Test error');
            });
        }
    }

    // Initialize the application
    document.addEventListener('DOMContentLoaded', () => {
        console.log('DOM content loaded, initializing application');
        
        // Create and mount the main App component
        const appInstance = new App();
        const appContainer = document.createElement('div');
        appContainer.id = 'app-container';
        document.body.appendChild(appContainer);
        appInstance.mount(appContainer);
        
        // Setup test mode if needed
        setupTestMode();
        
        console.log('Application initialized successfully');
    });
} 