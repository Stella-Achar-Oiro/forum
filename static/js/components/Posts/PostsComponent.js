// PostsComponent.js - Displays a list of posts
import Component from '../../core/Component.js';
import ComponentRegistry from '../../core/ComponentRegistry.js';

class PostsComponent extends Component {
    constructor(props) {
        super(props);
        this.state = {
            posts: [],
            loading: true,
            error: null,
            currentPage: 1,
            totalPages: 1,
            category: props?.category || null,
            searchQuery: ''
        };
        
        console.log('PostsComponent constructed with props:', props);
    }

    componentDidMount() {
        console.log('PostsComponent mounted');
        this.loadPosts();
    }

    // Load posts from the API
    async loadPosts() {
        try {
            this.setState({ loading: true, error: null });
            
            // Determine the API endpoint based on the category
            let url = '/api/posts';
            if (this.state.category) {
                url = `/api/posts/category/${this.state.category}`;
            }
            
            // Add search and pagination parameters
            const params = new URLSearchParams();
            if (this.state.searchQuery) {
                params.append('q', this.state.searchQuery);
            }
            params.append('page', this.state.currentPage.toString());
            
            const fullUrl = `${url}?${params.toString()}`;
            console.log('Loading posts from:', fullUrl);
            
            // Fetch posts
            const response = await fetch(fullUrl, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`Failed to load posts: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            console.log('Posts loaded:', data);
            
            this.setState({
                posts: data.posts || [],
                totalPages: data.totalPages || 1,
                loading: false
            });
        } catch (error) {
            console.error('Error loading posts:', error);
            this.setState({
                error: error.message || 'Failed to load posts. Please try again.',
                loading: false
            });
        }
    }
    
    // Handle page change
    handlePageChange(page) {
        this.setState({ currentPage: page }, () => {
            this.loadPosts();
        });
    }
    
    // Handle search input change
    handleSearchChange(e) {
        this.setState({ searchQuery: e.target.value });
    }
    
    // Handle search form submit
    handleSearchSubmit(e) {
        e.preventDefault();
        this.setState({ currentPage: 1 }, () => {
            this.loadPosts();
        });
    }
    
    // Format date for display
    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }
    
    // Render pagination controls
    renderPagination() {
        const { currentPage, totalPages } = this.state;
        
        if (totalPages <= 1) {
            return '';
        }
        
        let pages = [];
        
        // Always show first page
        pages.push(1);
        
        // Show dots if needed
        if (currentPage > 3) {
            pages.push('...');
        }
        
        // Show current page and neighbors
        for (let i = Math.max(2, currentPage - 1); i <= Math.min(totalPages - 1, currentPage + 1); i++) {
            if (pages.indexOf(i) === -1) {
                pages.push(i);
            }
        }
        
        // Show dots if needed
        if (currentPage < totalPages - 2) {
            pages.push('...');
        }
        
        // Always show last page
        if (totalPages > 1) {
            pages.push(totalPages);
        }
        
        return `
            <div class="pagination">
                <button 
                    class="pagination-btn prev ${currentPage === 1 ? 'disabled' : ''}" 
                    ${currentPage === 1 ? 'disabled' : ''} 
                    data-page="${currentPage - 1}">
                    &laquo; Previous
                </button>
                
                <div class="pagination-pages">
                    ${pages.map(page => {
                        if (page === '...') {
                            return '<span class="pagination-ellipsis">...</span>';
                        }
                        return `
                            <button 
                                class="pagination-btn page ${page === currentPage ? 'active' : ''}" 
                                data-page="${page}">
                                ${page}
                            </button>
                        `;
                    }).join('')}
                </div>
                
                <button 
                    class="pagination-btn next ${currentPage === totalPages ? 'disabled' : ''}" 
                    ${currentPage === totalPages ? 'disabled' : ''} 
                    data-page="${currentPage + 1}">
                    Next &raquo;
                </button>
            </div>
        `;
    }
    
    // Attach event handlers after rendering
    afterRender() {
        // Attach pagination event handlers
        const paginationButtons = this._container.querySelectorAll('.pagination-btn');
        paginationButtons.forEach(button => {
            if (!button.classList.contains('disabled')) {
                button.addEventListener('click', () => {
                    const page = parseInt(button.dataset.page, 10);
                    this.handlePageChange(page);
                });
            }
        });
        
        // Attach search form handlers
        const searchForm = this._container.querySelector('.search-form');
        if (searchForm) {
            searchForm.addEventListener('submit', e => this.handleSearchSubmit(e));
            
            const searchInput = searchForm.querySelector('input[type="search"]');
            if (searchInput) {
                searchInput.addEventListener('input', e => this.handleSearchChange(e));
            }
        }
        
        // Attach new post button handler
        const newPostButton = this._container.querySelector('.new-post-btn');
        if (newPostButton) {
            newPostButton.addEventListener('click', () => {
                window.location.hash = '#/posts/new';
            });
        }
    }
    
    renderContent() {
        console.log('PostsComponent renderContent called');
        
        const { posts, loading, error, searchQuery } = this.state;
        
        // Create container
        const container = document.createElement('div');
        container.className = 'posts-container';
        
        // Handle loading state
        if (loading) {
            container.innerHTML = `
                <div class="loading-container">
                    <div class="spinner"></div>
                    <p>Loading posts...</p>
                </div>
            `;
            return container;
        }
        
        // Handle error state
        if (error) {
            container.innerHTML = `
                <div class="error-container">
                    <h3>Error</h3>
                    <p>${error}</p>
                    <button class="btn btn-primary retry-btn">Retry</button>
                </div>
            `;
            // Attach retry handler
            setTimeout(() => {
                const retryBtn = container.querySelector('.retry-btn');
                if (retryBtn) {
                    retryBtn.addEventListener('click', () => this.loadPosts());
                }
            }, 0);
            return container;
        }
        
        // Handle empty state
        if (posts.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <h3>No posts found</h3>
                    <p>There are no posts to display. ${searchQuery ? 'Try a different search query.' : 'Be the first to create a post!'}</p>
                    <button class="btn btn-primary new-post-btn">Create New Post</button>
                </div>
            `;
            return container;
        }
        
        // Render posts
        container.innerHTML = `
            <div class="posts-header">
                <h2>Forum Posts</h2>
                <div class="posts-actions">
                    <form class="search-form">
                        <input 
                            type="search" 
                            placeholder="Search posts..." 
                            value="${searchQuery}"
                        >
                        <button type="submit" class="btn btn-primary">Search</button>
                    </form>
                    <button class="btn btn-primary new-post-btn">New Post</button>
                </div>
            </div>
            
            <div class="posts-list">
                ${posts.map(post => `
                    <div class="post-card" data-post-id="${post.id}">
                        <div class="post-header">
                            <h3 class="post-title">
                                <a href="#/posts/${post.id}">${post.title}</a>
                            </h3>
                            <div class="post-meta">
                                <span class="post-author">
                                    By ${post.authorName || 'Anonymous'}
                                </span>
                                <span class="post-date">
                                    ${this.formatDate(post.createdAt)}
                                </span>
                            </div>
                        </div>
                        <div class="post-content">
                            <p>${post.content.length > 200 ? post.content.substring(0, 200) + '...' : post.content}</p>
                        </div>
                        <div class="post-footer">
                            <div class="post-stats">
                                <span class="post-likes">
                                    <i class="icon-heart"></i> ${post.likes || 0}
                                </span>
                                <span class="post-comments">
                                    <i class="icon-comment"></i> ${post.commentCount || 0}
                                </span>
                            </div>
                            <div class="post-categories">
                                ${(post.categories || []).map(category => `
                                    <span class="post-category">${category.name}</span>
                                `).join('')}
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
            
            ${this.renderPagination()}
        `;
        
        return container;
    }
}

// Register component
ComponentRegistry.register('PostsComponent', PostsComponent);

// Export the component
export default PostsComponent; 