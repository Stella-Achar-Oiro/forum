// CreatePostComponent.js - Component for creating new forum posts
import Component from '../../core/Component.js';
import ComponentRegistry from '../../core/ComponentRegistry.js';

class CreatePostComponent extends Component {
    constructor(props) {
        super(props);
        this.state = {
            title: '',
            content: '',
            categories: [],
            selectedCategories: [],
            loading: false,
            error: null,
            success: false,
            loadingCategories: true
        };
        
        console.log('CreatePostComponent constructed');
    }

    componentDidMount() {
        console.log('CreatePostComponent mounted');
        // Load available categories
        this.loadCategories();
        
        // Check if user is logged in
        if (!window.currentUser) {
            // Redirect to login page
            window.location.hash = '#/login?redirect=' + encodeURIComponent('#/posts/new');
        }
    }

    // Load available categories from the API
    async loadCategories() {
        try {
            const response = await fetch('/api/categories', {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`Failed to load categories: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            console.log('Categories loaded:', data);
            
            this.setState({
                categories: data.categories || [],
                loadingCategories: false
            });
        } catch (error) {
            console.error('Error loading categories:', error);
            this.setState({
                error: 'Failed to load categories. Please try again.',
                loadingCategories: false
            });
        }
    }
    
    // Handle form input changes
    handleInputChange(e) {
        const { name, value } = e.target;
        this.setState({ [name]: value });
    }
    
    // Toggle category selection
    toggleCategory(categoryId) {
        const { selectedCategories } = this.state;
        
        if (selectedCategories.includes(categoryId)) {
            // Remove category
            this.setState({
                selectedCategories: selectedCategories.filter(id => id !== categoryId)
            });
        } else {
            // Add category
            this.setState({
                selectedCategories: [...selectedCategories, categoryId]
            });
        }
    }
    
    // Get CSRF token from cookie
    getCsrfToken() {
        const name = 'csrf_token=';
        const decodedCookie = decodeURIComponent(document.cookie);
        const cookieArray = decodedCookie.split(';');
        
        for (let i = 0; i < cookieArray.length; i++) {
            let cookie = cookieArray[i].trim();
            if (cookie.indexOf(name) === 0) {
                return cookie.substring(name.length, cookie.length);
            }
        }
        return '';
    }
    
    // Submit the new post
    async submitPost(e) {
        e.preventDefault();
        
        const { title, content, selectedCategories } = this.state;
        
        // Validate form inputs
        if (!title.trim()) {
            this.setState({ error: 'Please enter a title for your post.' });
            return;
        }
        
        if (!content.trim()) {
            this.setState({ error: 'Please enter content for your post.' });
            return;
        }
        
        // Show loading state and clear previous errors
        this.setState({
            loading: true,
            error: null,
            success: false
        });
        
        try {
            const response = await fetch('/api/posts', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': this.getCsrfToken()
                },
                credentials: 'include',
                body: JSON.stringify({
                    title,
                    content,
                    categories: selectedCategories
                })
            });
            
            if (!response.ok) {
                throw new Error(`Failed to create post: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            console.log('Post created:', data);
            
            // Show success message and reset form
            this.setState({
                loading: false,
                success: true,
                title: '',
                content: '',
                selectedCategories: []
            });
            
            // Redirect to the new post after a short delay
            setTimeout(() => {
                window.location.hash = `#/posts/${data.post.id}`;
            }, 1500);
        } catch (error) {
            console.error('Error creating post:', error);
            this.setState({
                loading: false,
                error: error.message || 'Failed to create post. Please try again.'
            });
        }
    }
    
    // Attach event handlers after rendering
    afterRender() {
        // Form submit handler
        const form = this._container.querySelector('form');
        if (form) {
            form.addEventListener('submit', e => this.submitPost(e));
        }
        
        // Title and content input handlers
        const titleInput = this._container.querySelector('input[name="title"]');
        const contentInput = this._container.querySelector('textarea[name="content"]');
        
        if (titleInput) {
            titleInput.addEventListener('input', e => this.handleInputChange(e));
        }
        
        if (contentInput) {
            contentInput.addEventListener('input', e => this.handleInputChange(e));
        }
        
        // Category checkbox handlers
        const categoryCheckboxes = this._container.querySelectorAll('.category-checkbox');
        categoryCheckboxes.forEach(checkbox => {
            checkbox.addEventListener('change', () => {
                const categoryId = parseInt(checkbox.dataset.categoryId, 10);
                this.toggleCategory(categoryId);
            });
        });
        
        // Cancel button handler
        const cancelButton = this._container.querySelector('.cancel-button');
        if (cancelButton) {
            cancelButton.addEventListener('click', () => {
                window.history.back();
            });
        }
    }
    
    renderContent() {
        console.log('CreatePostComponent renderContent called');
        
        const { title, content, categories, selectedCategories, loading, error, success, loadingCategories } = this.state;
        
        // Create container
        const container = document.createElement('div');
        container.className = 'create-post-container';
        
        // If not logged in, show message
        if (!window.currentUser) {
            container.innerHTML = `
                <div class="not-logged-in">
                    <h2>Please log in to create a post</h2>
                    <p>You need to be logged in to create a new post.</p>
                    <div class="action-buttons">
                        <a href="#/login" class="btn btn-primary">Log In</a>
                        <a href="#/posts" class="btn btn-secondary">Back to Posts</a>
                    </div>
                </div>
            `;
            return container;
        }
        
        // Show create post form
        container.innerHTML = `
            <div class="create-post">
                <div class="page-header">
                    <h2>Create New Post</h2>
                    <button class="btn btn-secondary cancel-button">Cancel</button>
                </div>
                
                ${error ? `
                    <div class="alert alert-danger">
                        <p>${error}</p>
                    </div>
                ` : ''}
                
                ${success ? `
                    <div class="alert alert-success">
                        <p>Post created successfully! Redirecting...</p>
                    </div>
                ` : ''}
                
                <form class="post-form">
                    <div class="form-group">
                        <label for="title">Title</label>
                        <input 
                            type="text" 
                            id="title" 
                            name="title" 
                            class="form-control" 
                            value="${title}" 
                            placeholder="Enter a title for your post" 
                            required
                            ${loading ? 'disabled' : ''}
                        >
                    </div>
                    
                    <div class="form-group">
                        <label for="content">Content</label>
                        <textarea 
                            id="content" 
                            name="content" 
                            class="form-control" 
                            rows="10" 
                            placeholder="Write your post content here..." 
                            required
                            ${loading ? 'disabled' : ''}
                        >${content}</textarea>
                    </div>
                    
                    <div class="form-group categories-group">
                        <label>Categories</label>
                        
                        ${loadingCategories ? `
                            <div class="loading-categories">
                                <div class="spinner-sm"></div>
                                <span>Loading categories...</span>
                            </div>
                        ` : categories.length === 0 ? `
                            <p class="no-categories">No categories available.</p>
                        ` : `
                            <div class="categories-list">
                                ${categories.map(category => `
                                    <div class="category-item">
                                        <input 
                                            type="checkbox" 
                                            id="category-${category.id}" 
                                            class="category-checkbox" 
                                            data-category-id="${category.id}"
                                            ${selectedCategories.includes(category.id) ? 'checked' : ''}
                                            ${loading ? 'disabled' : ''}
                                        >
                                        <label for="category-${category.id}">${category.name}</label>
                                    </div>
                                `).join('')}
                            </div>
                        `}
                    </div>
                    
                    <div class="form-actions">
                        <button 
                            type="submit" 
                            class="btn btn-primary create-button"
                            ${loading ? 'disabled' : ''}
                        >
                            ${loading ? `
                                <span class="spinner-sm"></span>
                                Creating...
                            ` : 'Create Post'}
                        </button>
                    </div>
                </form>
            </div>
        `;
        
        return container;
    }
}

// Register component
ComponentRegistry.register('CreatePostComponent', CreatePostComponent);

// Export the component
export default CreatePostComponent; 