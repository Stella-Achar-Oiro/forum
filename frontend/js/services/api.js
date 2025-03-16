// frontend/js/services/api.js
const API = {
    // Base API methods
    async request(url, options = {}) {
        try {
            console.log(`API Request: ${options.method || 'GET'} ${url}`);
            if (options.body) {
                console.log('Request body:', options.body);
            }
            
            const response = await fetch(url, {
                ...options,
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                }
            });
            
            // For better debugging
            console.log(`API Response status: ${response.status}`);
            
            let responseData;
            const responseText = await response.text();
            try {
                responseData = responseText ? JSON.parse(responseText) : {};
                console.log('API Response data:', responseData);
            } catch (e) {
                console.log('API Response text:', responseText);
                responseData = {};
            }
            
            if (!response.ok) {
                throw new Error(responseText || 'API request failed');
            }
            
            return responseData;
        } catch (error) {
            console.error('API Error:', error);
            throw error;
        }
    },
    
    // Auth endpoints
    auth: {
        register(userData) {
            return API.request('/api/register', {
                method: 'POST',
                body: JSON.stringify(userData)
            });
        },
        
        login(credentials) {
            return API.request('/api/login', {
                method: 'POST',
                body: JSON.stringify(credentials)
            });
        },
        
        logout() {
            return API.request('/api/logout', {
                method: 'POST'
            });
        },
        
        getCurrentUser() {
            return API.request('/api/me');
        }
    },
    
    // Posts endpoints
    posts: {
        getAllPosts() {
            return API.request('/api/posts');
        },
        
        getPost(postId) {
            return API.request(`/api/post?id=${postId}`);
        },
        
        createPost(postData) {
            return API.request('/api/posts', {
                method: 'POST',
                body: JSON.stringify(postData)
            });
        },
        
        createComment(postId, content) {
            return API.request(`/api/comments?postId=${postId}`, {
                method: 'POST',
                body: JSON.stringify({ content })
            });
        }
    },
    
    // Messages endpoints
    messages: {
        getChats() {
            return API.request('/api/chats');
        },
        
        getMessages(userId, limit = 10, offset = 0) {
            return API.request(`/api/messages?userId=${userId}&limit=${limit}&offset=${offset}`);
        },
        
        // frontend/js/services/api.js - Update messages.sendMessage

        sendMessage(receiverId, content, imageUrl = '') {
            return API.request('/api/send-message', {
                method: 'POST',
                body: JSON.stringify({
                    receiverId,
                    content,
                    imageUrl
                })
            });
        }
    },
    
    // Profile endpoints
    profile: {
        getProfile(userId) {
            return API.request(`/api/profile?userId=${userId}`);
        },
        
        updateProfile(profileData) {
            return API.request('/api/update-profile', {
                method: 'POST',
                body: JSON.stringify(profileData)
            });
        }
    }
};