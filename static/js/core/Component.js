// Component.js - Base component class for the application
// This is the foundation for all UI components

class Component {
    constructor() {
        this._isMounted = false;
        this._container = null;
        this.state = {};
        console.log(`Component ${this.constructor.name} created`);
    }

    setState(newState) {
        this.state = { ...this.state, ...newState };
        this.render();
    }

    mount(container) {
        console.log(`Mounting component ${this.constructor.name} to container:`, container);
        this._container = container;
        this._isMounted = true;
        
        // Store component reference on the container
        if (container) {
            container._component = this;
        }
        
        this.componentDidMount();
        this.render();
        
        return this;
    }

    unmount() {
        console.log(`Unmounting component ${this.constructor.name}`);
        this.componentWillUnmount();
        this._container = null;
        this._isMounted = false;
    }
    
    componentDidMount() {
        // Lifecycle method to be overridden by subclasses
        console.log(`componentDidMount default implementation for ${this.constructor.name}`);
    }
    
    componentWillUnmount() {
        // Lifecycle method to be overridden by subclasses
        console.log(`componentWillUnmount default implementation for ${this.constructor.name}`);
    }
    
    afterRender() {
        // Lifecycle method to be overridden by subclasses
        console.log(`afterRender default implementation for ${this.constructor.name}`);
    }
    
    render() {
        if (!this._isMounted) {
            console.warn('Attempted to render unmounted component', this.constructor.name);
            return;
        }
        
        if (!this._container) {
            console.error('No container to render into for', this.constructor.name);
            return;
        }
        
        try {
            console.log(`Rendering component ${this.constructor.name} into container:`, this._container);
            const content = this.renderContent();
            
            if (typeof content === 'string') {
                console.log(`Setting innerHTML for ${this.constructor.name}`);
                this._container.innerHTML = content;
            } else if (content instanceof HTMLElement) {
                console.log(`Appending HTMLElement for ${this.constructor.name}`);
                this._container.innerHTML = '';
                this._container.appendChild(content);
                console.log(`DOM updated for ${this.constructor.name}, container now has children:`, this._container.children.length);
            } else {
                console.error('Invalid content returned from renderContent:', content);
                this._container.innerHTML = '<div>Error: Component did not return valid content</div>';
                return;
            }
            
            // Enhanced: Explicitly store reference to the component on the container
            this._container._component = this;
            
            // Store component reference on all form elements and links
            // This allows us to get back to the component from event handlers if needed
            const elements = this._container.querySelectorAll('input, select, button, textarea, a, form');
            console.log(`Found ${elements.length} elements to attach component reference to`);
            elements.forEach(el => {
                el._component = this;
            });
            
            console.log(`Calling afterRender for ${this.constructor.name}`);
            
            // Call afterRender with a small delay to ensure DOM is updated
            setTimeout(() => {
                try {
                    this.afterRender();
                    console.log(`afterRender completed for ${this.constructor.name}`);
                } catch (error) {
                    console.error(`Error in afterRender for ${this.constructor.name}:`, error);
                }
            }, 50); // Increased delay to ensure DOM is fully updated
        } catch (error) {
            console.error(`Error rendering component ${this.constructor.name}:`, error);
            this._container.innerHTML = `<div class="error">Error rendering component: ${error.message}</div>`;
        }
    }
    
    renderContent() {
        // Abstract method to be implemented by subclasses
        console.error('renderContent method not implemented for component', this.constructor.name);
        return '<div>Component content not implemented</div>';
    }
    
    getRootComponent() {
        if (!this._container) {
            console.warn('Trying to get root component but container is null');
            return null;
        }
        
        return this._container._component || this;
    }

    createEl(tag, attributes = {}, children = []) {
        const element = document.createElement(tag);
        
        // Set attributes
        Object.entries(attributes).forEach(([key, value]) => {
            if (key.startsWith('on') && typeof value === 'function') {
                // Event listeners
                element.addEventListener(key.toLowerCase().slice(2), value);
            } else if (key === 'className') {
                // Class names
                element.setAttribute('class', value);
            } else {
                // Regular attributes
                element.setAttribute(key, value);
            }
        });

        // Add children
        children.forEach(child => {
            if (typeof child === 'string') {
                element.appendChild(document.createTextNode(child));
            } else if (child instanceof HTMLElement) {
                element.appendChild(child);
            }
        });

        return element;
    }

    // Navigation helper
    navigate(path) {
        if (window.router) {
            window.router.navigate(path);
        } else {
            console.error('Router not initialized');
        }
    }

    // Add store helper to Component class
    connectStore(selector, callback) {
        if (this._unsubscribe) {
            this._unsubscribe();
        }

        try {
            // Check if window.store is available
            if (!window.store || typeof window.store.subscribe !== 'function') {
                console.error('Store not available for connectStore or missing subscribe method');
                return;
            }
            
            this._unsubscribe = window.store.subscribe((state) => {
                const selectedState = selector(state);
                if (callback) {
                    callback(selectedState);
                } else {
                    this.setState(selectedState);
                }
            });

            // Initial state
            const initialState = selector(window.store.getState());
            if (callback) {
                callback(initialState);
            } else {
                this.setState(initialState);
            }
        } catch (error) {
            console.error('Error in connectStore:', error);
        }
    }
}

// Make the getRootComponent method available on HTMLElement prototypes
// This allows elements to access their component through this method
HTMLElement.prototype.getRootComponent = function() {
    return this._component || null;
};

// Add a static helper to directly test component rendering
Component.testRender = function(componentClass) {
    console.log('Testing render of', componentClass.name);
    const testContainer = document.createElement('div');
    testContainer.id = 'test-container';
    testContainer.style = 'position: fixed; top: 0; left: 0; background: white; padding: 20px; border: 2px solid red; z-index: 10000; max-height: 80vh; overflow: auto;';
    document.body.appendChild(testContainer);
    
    const component = new componentClass();
    component.mount(testContainer);
    
    return {
        component,
        remove: () => {
            component.unmount();
            document.body.removeChild(testContainer);
        }
    };
};

// Export the Component class for use in other modules
export default Component; 