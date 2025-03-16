// frontend/js/components/profile.js
const ProfileComponent = {
    currentProfile: null,
    
    // Render user profile
    async renderProfile(userId) {
        const container = document.createElement('div');
        container.className = 'content';
        
        try {
            // Load profile
            const data = await API.profile.getProfile(userId);
            this.currentProfile = data;
            
            const user = data.user;
            const profile = data.profile;
            const isOwnProfile = AuthService.user.id === userId;
            
            container.innerHTML = `
                <div class="profile-container">
                    <button id="back-btn" class="back-btn">← Back</button>
                    
                    <div class="profile-header">
                        <div class="profile-avatar">
                            <img src="/uploads/avatars/${profile.avatar}" alt="${user.nickname}" onerror="this.src='/img/default-avatar.png'">
                        </div>
                        <div class="profile-info">
                            <h2>${user.nickname}</h2>
                            <p>${user.firstName} ${user.lastName}</p>
                            <p>Member since: ${new Date(user.createdAt).toLocaleDateString()}</p>
                            <div class="profile-stats">
                                <span>${profile.postCount} posts</span>
                                <span>${profile.commentCount} comments</span>
                            </div>
                        </div>
                        ${isOwnProfile ? '<button id="edit-profile-btn" class="btn-primary">Edit Profile</button>' : ''}
                    </div>
                    
                    <div class="profile-bio">
                        <h3>Bio</h3>
                        <p>${profile.bio || 'No bio yet.'}</p>
                    </div>
                    
                    ${isOwnProfile ? this.renderProfileForm(profile) : ''}
                </div>
            `;
            
            // Add event listeners
            setTimeout(() => {
                try {
                    const backBtn = document.getElementById('back-btn');
                    if (backBtn) {
                        backBtn.addEventListener('click', () => {
                            App.renderHome();
                        });
                    }
                    
                    if (isOwnProfile) {
                        const editBtn = document.getElementById('edit-profile-btn');
                        const profileForm = document.getElementById('profile-form-container');
                        
                        if (editBtn && profileForm) {
                            editBtn.addEventListener('click', () => {
                                profileForm.style.display = profileForm.style.display === 'none' ? 'block' : 'none';
                            });
                        }
                        
                        const form = document.getElementById('profile-form');
                        if (form) {
                            form.addEventListener('submit', this.handleUpdateProfile.bind(this));
                        }
                        
                        const avatarInput = document.getElementById('avatar-input');
                        if (avatarInput) {
                            avatarInput.addEventListener('change', this.handleAvatarPreview);
                        }
                    }
                } catch (error) {
                    console.error("Error setting up profile event listeners:", error);
                }
            }, 0);
            
        } catch (error) {
            container.innerHTML = `
                <div class="error-message">
                    <button id="back-btn" class="back-btn">← Back</button>
                    <p>Error loading profile: ${error.message}</p>
                </div>
            `;
            
            setTimeout(() => {
                const backBtn = document.getElementById('back-btn');
                backBtn.addEventListener('click', () => {
                    App.renderHome();
                });
            }, 0);
        }
        
        return container;
    },
    
    // Render profile edit form
    renderProfileForm(profile) {
        return `
            <div id="profile-form-container" class="profile-form-container" style="display: none;">
                <h3>Edit Profile</h3>
                <form id="profile-form">
                    <div class="form-group">
                        <label for="bio">Bio</label>
                        <textarea id="bio" name="bio" rows="4">${profile.bio || ''}</textarea>
                    </div>
                    <div class="form-group">
                        <label for="avatar-input">Avatar</label>
                        <div class="avatar-preview">
                            <img id="avatar-preview" src="/uploads/avatars/${profile.avatar}" alt="Avatar Preview" onerror="this.src='/img/default-avatar.png'">
                        </div>
                        <input type="file" id="avatar-input" name="avatar" accept="image/*">
                    </div>
                    <button type="submit" class="btn-primary">Save Changes</button>
                </form>
            </div>
        `;
    },
    
    // Handle avatar preview
    handleAvatarPreview(e) {
        const file = e.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = function(e) {
                document.getElementById('avatar-preview').src = e.target.result;
            }
            reader.readAsDataURL(file);
        }
    },
    
    // Handle profile update
    async handleUpdateProfile(e) {
        e.preventDefault();
        
        const bioInput = document.getElementById('bio');
        const avatarInput = document.getElementById('avatar-input');
        
        try {
            // First handle avatar upload if there is one
            let avatarFilename = this.currentProfile.profile.avatar;
            
            if (avatarInput.files.length > 0) {
                const formData = new FormData();
                formData.append('avatar', avatarInput.files[0]);
                
                const response = await fetch('/api/upload-avatar', {
                    method: 'POST',
                    body: formData
                });
                
                if (!response.ok) {
                    throw new Error('Avatar upload failed');
                }
                
                const data = await response.json();
                avatarFilename = data.filename;
            }
            
            // Then update profile
            const profileData = {
                bio: bioInput.value,
                avatar: avatarFilename
            };
            
            await API.profile.updateProfile(profileData);
            
            // Reload profile
            App.renderProfile(AuthService.user.id);
            
        } catch (error) {
            alert('Error updating profile: ' + error.message);
        }
    }
};