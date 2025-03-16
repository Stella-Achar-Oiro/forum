// frontend/js/components/posts.js
const PostsComponent = {
    posts: [],
    currentPost: null,
    
    // Render posts feed
    async renderPosts() {
        const container = document.createElement('div');
        container.className = 'content';
        
        try {
            // Load posts
            this.posts = await API.posts.getAllPosts() || [];
            
            // Create HTML
            let postsHTML = '';
            
            if (!this.posts || this.posts.length === 0) {
                postsHTML = '<p>No posts yet. Be the first to create one!</p>';
            } else {
                postsHTML = this.posts.map(post => `
                    <div class="post" data-post-id="${post.id}">
                        <div class="post-header">
                            <h3>${post.title}</h3>
                            <span class="post-category">${post.category}</span>
                        </div>
                        <div class="post-meta">
                            Posted by ${post.user.nickname} on ${new Date(post.createdAt).toLocaleString()}
                        </div>
                        <div class="post-content">
                            ${post.content}
                        </div>
                        <div class="post-actions">
                            <button class="view-comments-btn" data-post-id="${post.id}">
                                View Comments
                            </button>
                        </div>
                    </div>
                `).join('');
            }
            
            container.innerHTML = `
                <div class="posts-header">
                    <h2>Recent Posts</h2>
                    <button id="new-post-btn">Create New Post</button>
                </div>
                <div class="posts-container">
                    ${postsHTML}
                </div>
                <div id="post-form-container" style="display: none;">
                    <h3>Create a New Post</h3>
                    <form id="post-form">
                        <div class="form-group">
                            <label for="post-title">Title</label>
                            <input type="text" id="post-title" name="title" required>
                        </div>
                        <div class="form-group">
                            <label for="post-category">Category</label>
                            <select id="post-category" name="category" required>
                                <option value="">Select a category</option>
                                <option value="General">General</option>
                                <option value="Technology">Technology</option>
                                <option value="Sports">Sports</option>
                                <option value="Gaming">Gaming</option>
                                <option value="Movies">Movies</option>
                                <option value="Music">Music</option>
                                <option value="Other">Other</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label for="post-content">Content</label>
                            <textarea id="post-content" name="content" rows="5" required></textarea>
                        </div>
                        <div class="btn-group">
                            <button type="button" id="cancel-post-btn">Cancel</button>
                            <button type="submit">Create Post</button>
                        </div>
                    </form>
                </div>
            `;
            
            // Add event listeners
            setTimeout(() => {
                try {
                    const newPostBtn = document.getElementById('new-post-btn');
                    if (newPostBtn) {
                        newPostBtn.addEventListener('click', this.togglePostForm.bind(this));
                    }
                    
                    const cancelPostBtn = document.getElementById('cancel-post-btn');
                    if (cancelPostBtn) {
                        cancelPostBtn.addEventListener('click', this.togglePostForm.bind(this));
                    }
                    
                    const postForm = document.getElementById('post-form');
                    if (postForm) {
                        postForm.addEventListener('submit', this.handleCreatePost.bind(this));
                    }
                    
                    const viewCommentsBtns = document.querySelectorAll('.view-comments-btn');
                    if (viewCommentsBtns && viewCommentsBtns.length > 0) {
                        viewCommentsBtns.forEach(btn => {
                            btn.addEventListener('click', (e) => {
                                const postId = parseInt(e.target.dataset.postId);
                                this.handleViewComments(postId);
                            });
                        });
                    }
                    
                    // Register for real-time updates
                    if (WebSocketService && typeof WebSocketService.onNewPost === 'function') {
                        WebSocketService.onNewPost(this.handleNewPost.bind(this));
                    }
                    if (WebSocketService && typeof WebSocketService.onNewComment === 'function') {
                        WebSocketService.onNewComment(this.handleNewComment.bind(this));
                    }
                } catch (error) {
                    console.error("Error setting up post event listeners:", error);
                }
            }, 0);
            
        } catch (error) {
            container.innerHTML = `
                <div class="error-message">
                    <p>Error loading posts: ${error.message}</p>
                    <button id="retry-btn">Retry</button>
                </div>
            `;
            
            setTimeout(() => {
                const retryBtn = document.getElementById('retry-btn');
                retryBtn.addEventListener('click', () => {
                    App.renderHome();
                });
            }, 0);
        }
        
        return container;
    },
    
    // Toggle post form visibility
    togglePostForm() {
        const formContainer = document.getElementById('post-form-container');
        formContainer.style.display = formContainer.style.display === 'none' ? 'block' : 'none';
    },
    
    // Handle creating a new post
    async handleCreatePost(e) {
        e.preventDefault();
        
        const title = document.getElementById('post-title').value;
        const category = document.getElementById('post-category').value;
        const content = document.getElementById('post-content').value;
        
        try {
            const newPost = await API.posts.createPost({ title, category, content });
            
            // Add to the posts array
            this.posts.unshift(newPost);
            
            // Notify other users
            WebSocketService.sendNewPostNotification(newPost.id);
            
            // Update the UI
            const postsContainer = document.querySelector('.posts-container');
            const postHTML = `
                <div class="post" data-post-id="${newPost.id}">
                    <div class="post-header">
                        <h3>${newPost.title}</h3>
                        <span class="post-category">${newPost.category}</span>
                    </div>
                    <div class="post-meta">
                        Posted by ${newPost.user.nickname} on ${new Date(newPost.createdAt).toLocaleString()}
                    </div>
                    <div class="post-content">
                        ${newPost.content}
                    </div>
                    <div class="post-actions">
                        <button class="view-comments-btn" data-post-id="${newPost.id}">
                            View Comments
                        </button>
                    </div>
                </div>
            `;
            
            postsContainer.insertAdjacentHTML('afterbegin', postHTML);
            
            // Add event listener to the new button
            const viewCommentsBtn = document.querySelector(`.view-comments-btn[data-post-id="${newPost.id}"]`);
            viewCommentsBtn.addEventListener('click', () => {
                this.handleViewComments(newPost.id);
            });
            
            // Reset form and hide it
            document.getElementById('post-form').reset();
            this.togglePostForm();
            
        } catch (error) {
            alert('Error creating post: ' + error.message);
        }
    },
    
    // Handle viewing a post with comments
    async handleViewComments(postId) {
        try {
            // Load post with comments
            const post = await API.posts.getPost(postId);
            this.currentPost = post;
            
            // Replace posts container with post details
            const content = document.querySelector('.content');
            
            content.innerHTML = `
                <div class="post-detail">
                    <button id="back-to-posts-btn">← Back to Posts</button>
                    
                    <div class="post">
                        <div class="post-header">
                            <h2>${post.title}</h2>
                            <span class="post-category">${post.category}</span>
                        </div>
                        <div class="post-meta">
                            Posted by ${post.user.nickname} on ${new Date(post.createdAt).toLocaleString()}
                        </div>
                        <div class="post-content">
                            ${post.content}
                        </div>
                    </div>
                    
                    <div class="comments-section">
                        <h3>Comments</h3>
                        <div class="comments-container">
                            ${this.renderComments(post.comments)}
                        </div>
                        
                        <div class="comment-form-container">
                            <h4>Add a Comment</h4>
                            <form id="comment-form">
                                <div class="form-group">
                                    <textarea id="comment-content" name="content" rows="3" required></textarea>
                                </div>
                                <button type="submit">Post Comment</button>
                            </form>
                        </div>
                    </div>
                </div>
            `;
            
            // Add event listeners
            setTimeout(() => {
                const backBtn = document.getElementById('back-to-posts-btn');
                backBtn.addEventListener('click', () => {
                    App.renderHome();
                });
                
                const commentForm = document.getElementById('comment-form');
                commentForm.addEventListener('submit', (e) => {
                    e.preventDefault();
                    this.handleCreateComment(post.id);
                });
            }, 0);
            
        } catch (error) {
            alert('Error loading post: ' + error.message);
        }
    },
    
    // Render comments HTML
    renderComments(comments) {
        if (comments.length === 0) {
            return '<p>No comments yet. Be the first to comment!</p>';
        }
        
        return comments.map(comment => `
            <div class="comment" data-comment-id="${comment.id}">
                <div class="comment-meta">
                    ${comment.user.nickname} commented on ${new Date(comment.createdAt).toLocaleString()}
                </div>
                <div class="comment-content">
                    ${comment.content}
                </div>
            </div>
        `).join('');
    },
    
    // Handle creating a new comment
    async handleCreateComment(postId) {
        const content = document.getElementById('comment-content').value;
        
        try {
            const newComment = await API.posts.createComment(postId, content);
            
            // Notify other users
            WebSocketService.sendNewCommentNotification(postId, newComment.id);
            
            // Add comment to UI
            const commentsContainer = document.querySelector('.comments-container');
            const noCommentsMsg = commentsContainer.querySelector('p');
            if (noCommentsMsg) {
                commentsContainer.innerHTML = '';
            }
            
            const commentHTML = `
                <div class="comment" data-comment-id="${newComment.id}">
                    <div class="comment-meta">
                        ${newComment.user.nickname} commented on ${new Date(newComment.createdAt).toLocaleString()}
                    </div>
                    <div class="comment-content">
                        ${newComment.content}
                    </div>
                </div>
            `;
            
            commentsContainer.insertAdjacentHTML('beforeend', commentHTML);
            
            // Reset form
            document.getElementById('comment-form').reset();
            
        } catch (error) {
            alert('Error creating comment: ' + error.message);
        }
    },
    
    // Handle new post notification
    handleNewPost(data) {
        // Refresh posts if we're on the home page
        if (!this.currentPost) {
            App.renderHome();
        }
    },
    
    // Handle new comment notification
    handleNewComment(data) {
        // Refresh comments if we're viewing the affected post
        if (this.currentPost && this.currentPost.id === data.postId) {
            this.handleViewComments(data.postId);
        }
    }
};