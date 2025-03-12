// ComponentRegistry.js
// Handles registration and retrieval of components

class ComponentRegistry {
    static components = {};
    
    /**
     * Register a component with the registry
     * @param {string} name - The name of the component
     * @param {class} componentClass - The component class
     */
    static register(name, componentClass) {
        console.log(`Registering component: ${name}`);
        this.components[name] = componentClass;
    }
    
    /**
     * Get a component from the registry
     * @param {string} name - The name of the component
     * @returns {class|null} - The component class or null if not found
     */
    static get(name) {
        const component = this.components[name];
        if (!component) {
            console.warn(`Component not found: ${name}`);
            return null;
        }
        return component;
    }
    
    /**
     * Check if a component is registered
     * @param {string} name - The name of the component
     * @returns {boolean} - True if the component is registered
     */
    static has(name) {
        return !!this.components[name];
    }
    
    /**
     * Get all registered component names
     * @returns {string[]} - Array of component names
     */
    static getNames() {
        return Object.keys(this.components);
    }
}

// Create global window reference for debugging
window.ComponentRegistry = ComponentRegistry;

// Export the ComponentRegistry for use in other modules
export default ComponentRegistry; 