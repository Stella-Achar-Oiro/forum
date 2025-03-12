class Router {
    constructor(routes, container) {
        this.routes = routes;
        this.container = container || document.getElementById('main-container');
        console.log('Router initialized with container:', this.container);
        
        if (!this.container) {
            console.error('Main container not found. Creating one.');
            this.container = document.createElement('div');
            this.container.id = 'main-container';
            document.getElementById('app').appendChild(this.container);
        }
        
        this.currentComponent = null;
        this.params = {};

        // Handle browser back/forward buttons
        window.addEventListener('popstate', () => this.handleRoute());

        // Handle initial route
        this.handleRoute();
    }

    handleRoute() {
        const path = window.location.pathname;
        console.log('Router handling route:', path);
        
        let route = null;
        let params = {};

        // First check for exact match
        route = this.routes.find(r => r.path === path);
        console.log('Exact match found:', !!route);

        // If no exact match, try matching routes with parameters
        if (!route) {
            for (const r of this.routes) {
                if (r.path.includes(':')) {
                    const routeParts = r.path.split('/');
                    const pathParts = path.split('/');
                    
                    if (routeParts.length === pathParts.length) {
                        let match = true;
                        const tempParams = {};
                        
                        for (let i = 0; i < routeParts.length; i++) {
                            if (routeParts[i].startsWith(':')) {
                                // This is a parameter
                                const paramName = routeParts[i].substring(1);
                                tempParams[paramName] = pathParts[i];
                            } else if (routeParts[i] !== pathParts[i]) {
                                match = false;
                                break;
                            }
                        }
                        
                        if (match) {
                            route = r;
                            params = tempParams;
                            console.log('Param match found:', r.path, 'with params:', params);
                            break;
                        }
                    }
                }
            }
        }

        // Fallback to wildcard route
        if (!route) {
            route = this.routes.find(r => r.path === '*');
            console.log('Using wildcard route:', !!route);
        }

        if (!route) {
            console.error('No route found for path:', path);
            return;
        }

        console.log('Selected route component:', route.component.name || 'Anonymous Component');

        // Save parameters
        this.params = params;

        // Unmount current component if exists
        if (this.currentComponent) {
            console.log('Unmounting current component');
            this.currentComponent.unmount();
        }

        // Check container again before mounting
        if (!this.container) {
            console.error('Container missing before mounting component. Creating one.');
            this.container = document.createElement('div');
            this.container.id = 'main-container';
            document.getElementById('app').appendChild(this.container);
        }

        try {
            // Create and mount new component
            const props = { ...route.props, params };
            console.log('Creating new component instance');
            this.currentComponent = new route.component(props);
            console.log('Mounting component to container:', this.container);
            this.currentComponent.mount(this.container);
        } catch (error) {
            console.error('Error mounting component:', error);
            // Fallback to simple HTML
            this.container.innerHTML = `
                <div class="error-container">
                    <h1>Error rendering page</h1>
                    <p>${error.message}</p>
                    <p>Check the console for more details.</p>
                </div>
            `;
        }
    }

    navigate(path) {
        window.history.pushState({}, '', path);
        this.handleRoute();
    }

    getParams() {
        return this.params;
    }

    static init(routes) {
        console.log('Router.init called');
        
        // Create container if it doesn't exist
        let container = document.getElementById('main-container');
        if (!container) {
            console.log('Creating main-container in Router.init');
            container = document.createElement('div');
            container.id = 'main-container';
            
            // Find app element or create it
            let appElement = document.getElementById('app');
            if (!appElement) {
                console.log('Creating app element in Router.init');
                appElement = document.createElement('div');
                appElement.id = 'app';
                document.body.appendChild(appElement);
            }
            
            appElement.appendChild(container);
        }
        
        // Create router instance
        const router = new Router(routes, container);
        
        // Add it to window for global access
        window.router = router;
        
        return router;
    }
}

// No need to add navigation to Component prototype as it's already in the Component class 

// Export the Router class for ES6 module usage
export default Router; 