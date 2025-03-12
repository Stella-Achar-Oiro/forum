// PostDetailComponent.js - Displays a single post and its comments
import Component from '../../core/Component.js';
import ComponentRegistry from '../../core/ComponentRegistry.js';

class PostDetailComponent extends Component {
    constructor(props) {
        super(props);
        this.state = {
            post: null,
            comments: [],
            loading: true,
            error: null,
            newComment: '',
            editMode: false,
            editedPost: {
                title: '',
                content: ''
            }
        };
        
        // Extract post ID from props or URL
        this.postId = props?.postId || this.getPostIdFromUrl();
        console.log('PostDetailComponent constructed with postId:', this.postId);
    }

    componentDidMount() {
        console.log('PostDetailComponent mounted');
        if (this.postId) {
            this.loadPostDetails();
        } else {
            this.setState({
                loading: false,
                error: 'Post ID is missing. Please try again.'
            });
        }
    }

    // Extract post ID from the URL hash
    getPostIdFromUrl() {
        const hash = window.location.hash;
        const match = hash.match(/#\/posts\/(\d+)/);
        return match ? match[1] : null;
    }

    // Load post details from the API
    async loadPostDetails() {
        try {
            this.setState({ loading: true, error: null });
            
            // Fetch post details
            const response = await fetch(`/api/posts/${this.postId}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`Failed to load post: ${response.status} ${response.statusText}`);
            }
            
            const postData = await response.json();
            console.log('Post loaded:', postData);
            
            // Fetch comments for this post
            const commentsResponse = await fetch(`/api/posts/${this.postId}/comments`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include'
            });
            
            if (!commentsResponse.ok) {
                throw new Error(`Failed to load comments: ${commentsResponse.status} ${commentsResponse.statusText}`);
            }
            
            const commentsData = await commentsResponse.json();
            console.log('Comments loaded:', commentsData);
            
            this.setState({
                post: postData,
                comments: commentsData.comments || [],
                loading: false,
                editedPost: {
                    title: postData.title,
                    content: postData.content
                }
            });
        } catch (error) {
            console.error('Error loading post details:', error);
            this.setState({
                error: error.message || 'Failed to load post details. Please try again.',
                loading: false
            });
        }
    }
    
    // Handle new comment input change
    handleCommentInputChange(e) {
        this.setState({ newComment: e.target.value });
    }
    
    // Submit a new comment
    async submitComment(e) {
        e.preventDefault();
        
        const { newComment } = this.state;
        if (!newComment.trim()) {
            return;
        }
        
        try {
            const response = await fetch(`/api/posts/${this.postId}/comments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': this.getCsrfToken()
                },
                credentials: 'include',
                body: JSON.stringify({
                    content: newComment
                })
            });
            
            if (!response.ok) {
                throw new Error(`Failed to submit comment: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            console.log('Comment submitted:', data);
            
            // Add the new comment to the state and clear the input
            this.setState({
                comments: [...this.state.comments, data.comment],
                newComment: ''
            });
        } catch (error) {
            console.error('Error submitting comment:', error);
            alert('Failed to submit comment: ' + (error.message || 'Unknown error'));
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
    
    // Toggle like on post
    async toggleLike() {
        try {
            const { post } = this.state;
            const isLiked = post.isLiked;
            
            const response = await fetch(`/api/posts/${this.postId}/like`, {
                method: isLiked ? 'DELETE' : 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': this.getCsrfToken()
                },
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`Failed to ${isLiked ? 'unlike' : 'like'} post: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            console.log('Like toggled:', data);
            
            // Update post state with new like count and status
            this.setState({
                post: {
                    ...post,
                    likes: data.likes,
                    isLiked: !isLiked
                }
            });
        } catch (error) {
            console.error('Error toggling like:', error);
            alert('Failed to update like: ' + (error.message || 'Unknown error'));
        }
    }
    
    // Toggle edit mode
    toggleEditMode() {
        const { editMode, post } = this.state;
        
        this.setState({
            editMode: !editMode,
            editedPost: {
                title: post.title,
                content: post.content
            }
        });
    }
    
    // Handle edit form input changes
    handleEditInputChange(e) {
        const { name, value } = e.target;
        
        this.setState({
            editedPost: {
                ...this.state.editedPost,
                [name]: value
            }
        });
    }
    
    // Submit edited post
    async submitEditedPost(e) {
        e.preventDefault();
        
        const { editedPost } = this.state;
        
        try {
            const response = await fetch(`/api/posts/${this.postId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': this.getCsrfToken()
                },
                credentials: 'include',
                body: JSON.stringify({
                    title: editedPost.title,
                    content: editedPost.content
                })
            });
            
            if (!response.ok) {
                throw new Error(`Failed to update post: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            console.log('Post updated:', data);
            
            // Update post state and exit edit mode
            this.setState({
                post: {
                    ...this.state.post,
                    title: editedPost.title,
                    content: editedPost.content,
                    updatedAt: data.updatedAt
                },
                editMode: false
            });
        } catch (error) {
            console.error('Error updating post:', error);
            alert('Failed to update post: ' + (error.message || 'Unknown error'));
        }
    }
    
    // Delete post
    async deletePost() {
        if (!confirm('Are you sure you want to delete this post? This action cannot be undone.')) {
            return;
        }
        
        try {
            const response = await fetch(`/api/posts/${this.postId}`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': this.getCsrfToken()
                },
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`Failed to delete post: ${response.status} ${response.statusText}`);
            }
            
            console.log('Post deleted successfully');
            
            // Redirect to posts list
            window.location.hash = '#/posts';
        } catch (error) {
            console.error('Error deleting post:', error);
            alert('Failed to delete post: ' + (error.message || 'Unknown error'));
        }
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
    
    // Attach event handlers after rendering
    afterRender() {
        const { editMode } = this.state;
        
        // Like button handler
        const likeButton = this._container.querySelector('.like-button');
        if (likeButton) {
            likeButton.addEventListener('click', () => this.toggleLike());
        }
        
        // Edit button handler
        const editButton = this._container.querySelector('.edit-button');
        if (editButton) {
            editButton.addEventListener('click', () => this.toggleEditMode());
        }
        
        // Delete button handler
        const deleteButton = this._container.querySelector('.delete-button');
        if (deleteButton) {
            deleteButton.addEventListener('click', () => this.deletePost());
        }
        
        // Comment form handler
        const commentForm = this._container.querySelector('.comment-form');
        if (commentForm) {
            commentForm.addEventListener('submit', e => this.submitComment(e));
            
            const commentInput = commentForm.querySelector('textarea');
            if (commentInput) {
                commentInput.addEventListener('input', e => this.handleCommentInputChange(e));
            }
        }
        
        // Edit form handlers
        if (editMode) {
            const editForm = this._container.querySelector('.edit-form');
            if (editForm) {
                editForm.addEventListener('submit', e => this.submitEditedPost(e));
                
                const titleInput = editForm.querySelector('input[name="title"]');
                const contentInput = editForm.querySelector('textarea[name="content"]');
                
                if (titleInput) {
                    titleInput.addEventListener('input', e => this.handleEditInputChange(e));
                }
                
                if (contentInput) {
                    contentInput.addEventListener('input', e => this.handleEditInputChange(e));
                }
                
                const cancelButton = editForm.querySelector('.cancel-button');
                if (cancelButton) {
                    cancelButton.addEventListener('click', () => this.toggleEditMode());
                }
            }
        }
        
        // Back button handler
        const backButton = this._container.querySelector('.back-button');
        if (backButton) {
            backButton.addEventListener('click', () => {
                window.history.back();
            });
        }
    }
    
    renderContent() {
        console.log('PostDetailComponent renderContent called');
        
        const { post, comments, loading, error, newComment, editMode, editedPost } = this.state;
        
        // Create container
        const container = document.createElement('div');
        container.className = 'post-detail-container';
        
        // Handle loading state
        if (loading) {
            container.innerHTML = `
                <div class="loading-container">
                    <div class="spinner"></div>
                    <p>Loading post...</p>
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
                    <div class="error-actions">
                        <button class="btn btn-primary retry-btn">Retry</button>
                        <button class="btn btn-secondary back-button">Back to Posts</button>
                    </div>
                </div>
            `;
            
            // Attach retry handler
            setTimeout(() => {
                const retryBtn = container.querySelector('.retry-btn');
                if (retryBtn) {
                    retryBtn.addEventListener('click', () => this.loadPostDetails());
                }
            }, 0);
            
            return container;
        }
        
        // If post not found
        if (!post) {
            container.innerHTML = `
                <div class="post-not-found">
                    <h3>Post Not Found</h3>
                    <p>The post you're looking for doesn't exist or has been removed.</p>
                    <button class="btn btn-primary back-button">Back to Posts</button>
                </div>
            `;
            return container;
        }
        
        // Determine if user can edit/delete the post
        const isAuthor = post.authorId === (window.currentUser?.id || null);
        
        // Render edit form or post content
        if (editMode && isAuthor) {
            container.innerHTML = `
                <div class="post-detail">
                    <div class="post-actions">
                        <button class="btn btn-secondary back-button">Back to Posts</button>
                    </div>
                    
                    <form class="edit-form">
                        <h2>Edit Post</h2>
                        
                        <div class="form-group">
                            <label for="title">Title</label>
                            <input 
                                type="text" 
                                name="title" 
                                id="title" 
                                class="form-control" 
                                value="${editedPost.title}" 
                                required
                            >
                        </div>
                        
                        <div class="form-group">
                            <label for="content">Content</label>
                            <textarea 
                                name="content" 
                                id="content" 
                                class="form-control" 
                                rows="10" 
                                required
                            >${editedPost.content}</textarea>
                        </div>
                        
                        <div class="form-actions">
                            <button type="submit" class="btn btn-primary">Update Post</button>
                            <button type="button" class="btn btn-secondary cancel-button">Cancel</button>
                        </div>
                    </form>
                </div>
            `;
        } else {
            // Render post details
            container.innerHTML = `
                <div class="post-detail">
                    <div class="post-actions">
                        <button class="btn btn-secondary back-button">Back to Posts</button>
                        ${isAuthor ? `
                            <div class="author-actions">
                                <button class="btn btn-primary edit-button">Edit</button>
                                <button class="btn btn-danger delete-button">Delete</button>
                            </div>
                        ` : ''}
                    </div>
                    
                    <div class="post-header">
                        <h1 class="post-title">${post.title}</h1>
                        <div class="post-meta">
                            <span class="post-author">By ${post.authorName || 'Anonymous'}</span>
                            <span class="post-date">
                                Posted on ${this.formatDate(post.createdAt)}
                                ${post.updatedAt && post.updatedAt !== post.createdAt 
                                    ? ` (Edited on ${this.formatDate(post.updatedAt)})` 
                                    : ''}
                            </span>
                        </div>
                        
                        <div class="post-categories">
                            ${(post.categories || []).map(category => `
                                <span class="post-category">${category.name}</span>
                            `).join('')}
                        </div>
                    </div>
                    
                    <div class="post-content">
                        ${post.content.split('\n').map(paragraph => 
                            paragraph ? `<p>${paragraph}</p>` : ''
                        ).join('')}
                    </div>
                    
                    <div class="post-footer">
                        <div class="post-stats">
                            <button class="like-button ${post.isLiked ? 'liked' : ''}">
                                <i class="icon-heart"></i>
                                <span>${post.likes || 0}</span>
                            </button>
                            <span class="post-comments-count">
                                <i class="icon-comment"></i>
                                <span>${comments.length}</span> Comments
                            </span>
                        </div>
                    </div>
                    
                    <div class="comments-section">
                        <h3>Comments</h3>
                        
                        ${window.currentUser ? `
                            <form class="comment-form">
                                <div class="form-group">
                                    <textarea 
                                        class="form-control" 
                                        placeholder="Add a comment..." 
                                        rows="3"
                                        required
                                    >${newComment}</textarea>
                                </div>
                                <button type="submit" class="btn btn-primary">Post Comment</button>
                            </form>
                        ` : `
                            <div class="login-to-comment">
                                <p>Please <a href="#/login">log in</a> to comment.</p>
                            </div>
                        `}
                        
                        <div class="comments-list">
                            ${comments.length === 0 ? `
                                <div class="no-comments">
                                    <p>No comments yet. Be the first to comment!</p>
                                </div>
                            ` : comments.map(comment => `
                                <div class="comment" data-comment-id="${comment.id}">
                                    <div class="comment-header">
                                        <span class="comment-author">${comment.authorName || 'Anonymous'}</span>
                                        <span class="comment-date">${this.formatDate(comment.createdAt)}</span>
                                    </div>
                                    <div class="comment-content">
                                        ${comment.content.split('\n').map(paragraph => 
                                            paragraph ? `<p>${paragraph}</p>` : ''
                                        ).join('')}
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                </div>
            `;
        }
        
        return container;
    }
}

// Register component
ComponentRegistry.register('PostDetailComponent', PostDetailComponent);

// Export the component
export default PostDetailComponent; 