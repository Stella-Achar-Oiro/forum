// ProfileComponent.js - User profile management
import Component from '../../core/Component.js';
import ComponentRegistry from '../../core/ComponentRegistry.js';

class ProfileComponent extends Component {
    constructor(props) {
        super(props);
        this.state = {
            user: props?.user || window.currentUser || null,
            loading: true,
            error: null,
            editMode: false,
            formData: {}
        };
        
        console.log('ProfileComponent constructed with props:', props);
    }

    componentDidMount() {
        console.log('ProfileComponent mounted');
        
        // Check if user is logged in
        if (!this.state.user) {
            // Redirect to login page if not logged in
            window.location.hash = '#/login?redirect=' + encodeURIComponent('#/profile');
            return;
        }
        
        // Load user profile data
        this.loadUserProfile();
    }
    
    // Load user profile from the API
    async loadUserProfile() {
        try {
            this.setState({ loading: true, error: null });
            
            // Fetch user profile
            const response = await fetch('/api/user/profile', {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`Failed to load profile: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            console.log('Profile data loaded:', data);
            
            // Update state with profile data
            this.setState({
                loading: false,
                formData: {
                    nickname: data.nickname || '',
                    firstName: data.firstName || '',
                    lastName: data.lastName || '',
                    email: data.email || '',
                    bio: data.bio || '',
                    age: data.age || '',
                    gender: data.gender || 'prefer_not_to_say'
                }
            });
        } catch (error) {
            console.error('Error loading profile:', error);
            this.setState({
                error: error.message || 'Failed to load profile. Please try again.',
                loading: false
            });
        }
    }
    
    // Toggle edit mode
    toggleEditMode() {
        this.setState({ 
            editMode: !this.state.editMode 
        });
    }
    
    // Handle form input changes
    handleInputChange(e) {
        const { name, value } = e.target;
        this.setState({
            formData: {
                ...this.state.formData,
                [name]: value
            }
        });
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
    
    // Submit profile changes
    async submitProfileChanges(e) {
        e.preventDefault();
        
        this.setState({ loading: true, error: null });
        
        try {
            const response = await fetch('/api/user/profile', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': this.getCsrfToken()
                },
                credentials: 'include',
                body: JSON.stringify(this.state.formData)
            });
            
            if (!response.ok) {
                throw new Error(`Failed to update profile: ${response.status} ${response.statusText}`);
            }
            
            const data = await response.json();
            console.log('Profile updated:', data);
            
            this.setState({
                loading: false,
                editMode: false,
                success: 'Profile updated successfully'
            });
            
            // Clear success message after a few seconds
            setTimeout(() => {
                this.setState({ success: null });
            }, 3000);
        } catch (error) {
            console.error('Error updating profile:', error);
            this.setState({
                error: error.message || 'Failed to update profile. Please try again.',
                loading: false
            });
        }
    }
    
    // Attach event handlers after rendering
    afterRender() {
        // Toggle edit mode button
        const editButton = this._container.querySelector('.edit-profile-btn');
        if (editButton) {
            editButton.addEventListener('click', () => this.toggleEditMode());
        }
        
        // Form handlers in edit mode
        if (this.state.editMode) {
            const form = this._container.querySelector('.profile-form');
            if (form) {
                form.addEventListener('submit', e => this.submitProfileChanges(e));
            }
            
            // Input field handlers
            const inputs = this._container.querySelectorAll('input, textarea, select');
            inputs.forEach(input => {
                input.addEventListener('change', e => this.handleInputChange(e));
            });
            
            // Cancel button
            const cancelButton = this._container.querySelector('.cancel-btn');
            if (cancelButton) {
                cancelButton.addEventListener('click', () => this.toggleEditMode());
            }
        }
    }
    
    renderContent() {
        console.log('ProfileComponent renderContent called');
        
        const { user, loading, error, success, editMode, formData } = this.state;
        
        // Create container
        const container = document.createElement('div');
        container.className = 'profile-container';
        
        // If not logged in
        if (!user) {
            container.innerHTML = `
                <div class="not-logged-in">
                    <h2>Please log in to view your profile</h2>
                    <p>You need to be logged in to view and edit your profile.</p>
                    <div class="action-buttons">
                        <a href="#/login" class="btn btn-primary">Log In</a>
                    </div>
                </div>
            `;
            return container;
        }
        
        // Handle loading state
        if (loading) {
            container.innerHTML = `
                <div class="loading-container">
                    <div class="spinner"></div>
                    <p>Loading profile...</p>
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
            return container;
        }
        
        // In edit mode, show form
        if (editMode) {
            container.innerHTML = `
                <div class="profile-card">
                    <div class="profile-header">
                        <h2>Edit Profile</h2>
                    </div>
                    
                    ${success ? `<div class="success-message">${success}</div>` : ''}
                    
                    <form class="profile-form">
                        <div class="form-group">
                            <label for="nickname">Username</label>
                            <input 
                                type="text" 
                                id="nickname" 
                                name="nickname" 
                                value="${formData.nickname || ''}" 
                                placeholder="Username"
                                required
                            >
                        </div>
                        
                        <div class="form-row">
                            <div class="form-group half">
                                <label for="firstName">First Name</label>
                                <input 
                                    type="text" 
                                    id="firstName" 
                                    name="firstName" 
                                    value="${formData.firstName || ''}" 
                                    placeholder="First Name"
                                >
                            </div>
                            
                            <div class="form-group half">
                                <label for="lastName">Last Name</label>
                                <input 
                                    type="text" 
                                    id="lastName" 
                                    name="lastName" 
                                    value="${formData.lastName || ''}" 
                                    placeholder="Last Name"
                                >
                            </div>
                        </div>
                        
                        <div class="form-group">
                            <label for="email">Email</label>
                            <input 
                                type="email" 
                                id="email" 
                                name="email" 
                                value="${formData.email || ''}" 
                                placeholder="Email"
                                required
                            >
                        </div>
                        
                        <div class="form-group">
                            <label for="bio">Bio</label>
                            <textarea 
                                id="bio" 
                                name="bio" 
                                rows="4" 
                                placeholder="Tell us about yourself"
                            >${formData.bio || ''}</textarea>
                        </div>
                        
                        <div class="form-row">
                            <div class="form-group half">
                                <label for="age">Age</label>
                                <input 
                                    type="number" 
                                    id="age" 
                                    name="age" 
                                    value="${formData.age || ''}" 
                                    placeholder="Age"
                                    min="13"
                                    max="120"
                                >
                            </div>
                            
                            <div class="form-group half">
                                <label for="gender">Gender</label>
                                <select id="gender" name="gender">
                                    <option value="male" ${formData.gender === 'male' ? 'selected' : ''}>Male</option>
                                    <option value="female" ${formData.gender === 'female' ? 'selected' : ''}>Female</option>
                                    <option value="other" ${formData.gender === 'other' ? 'selected' : ''}>Other</option>
                                    <option value="prefer_not_to_say" ${formData.gender === 'prefer_not_to_say' ? 'selected' : ''}>Prefer not to say</option>
                                </select>
                            </div>
                        </div>
                        
                        <div class="form-actions">
                            <button type="submit" class="btn btn-primary save-btn">Save Changes</button>
                            <button type="button" class="btn btn-secondary cancel-btn">Cancel</button>
                        </div>
                    </form>
                </div>
            `;
        } else {
            // View mode, show profile
            container.innerHTML = `
                <div class="profile-card">
                    <div class="profile-header">
                        <h2>Your Profile</h2>
                        <button class="btn btn-primary edit-profile-btn">Edit Profile</button>
                    </div>
                    
                    ${success ? `<div class="success-message">${success}</div>` : ''}
                    
                    <div class="profile-info">
                        <div class="profile-avatar">
                            <img src="${user.avatar || '/static/images/default-avatar.png'}" alt="Profile Avatar">
                        </div>
                        
                        <div class="profile-details">
                            <h3 class="profile-name">${formData.firstName || ''} ${formData.lastName || ''}</h3>
                            <p class="profile-username">@${formData.nickname || user.nickname || 'username'}</p>
                            <p class="profile-email">${formData.email || user.email || 'email@example.com'}</p>
                            
                            ${formData.bio ? `
                                <div class="profile-bio">
                                    <h4>Bio</h4>
                                    <p>${formData.bio}</p>
                                </div>
                            ` : ''}
                            
                            <div class="profile-meta">
                                ${formData.age ? `<span class="meta-item">Age: ${formData.age}</span>` : ''}
                                ${formData.gender && formData.gender !== 'prefer_not_to_say' ? 
                                    `<span class="meta-item">Gender: ${formData.gender.charAt(0).toUpperCase() + formData.gender.slice(1)}</span>` : ''}
                            </div>
                        </div>
                    </div>
                    
                    <!-- Placeholder for future sections like activity, settings, etc. -->
                </div>
            `;
        }
        
        return container;
    }
}

// Register component
ComponentRegistry.register('ProfileComponent', ProfileComponent);

// Export the component
export default ProfileComponent; 