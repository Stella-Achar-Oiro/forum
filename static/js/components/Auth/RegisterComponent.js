// RegisterComponent.js - Handles user registration
import Component from '../../core/Component.js';
import ComponentRegistry from '../../core/ComponentRegistry.js';
import AuthService from '../../services/AuthService.js';

class RegisterComponent extends Component {
    constructor() {
        super();
        this.state = {
            nickname: '',
            email: '',
            password: '',
            confirmPassword: '',
            firstName: '',
            lastName: '',
            age: '',
            gender: 'prefer_not_to_say',
            error: null,
            formErrors: [],
            success: false,
            loading: false,
            passwordStrength: 0,
            passwordStrengthInfo: { text: 'Very Weak', color: '#ff0000' },
            passwordMismatch: false
        };
        
        console.log('RegisterComponent constructed');
    }

    componentDidMount() {
        console.log('RegisterComponent mounted');
        // Add direct DOM test to verify component is mounted
        this.addDirectDomTest();
        
        // We'll try attaching events later in afterRender
    }

    // This method is called after the component has been rendered
    afterRender() {
        console.log('RegisterComponent afterRender');
        
        // Use a small delay to ensure DOM is fully updated
        setTimeout(() => {
            this.attachFormEventHandlers();
        }, 50);
    }

    attachFormEventHandlers() {
        console.log('Trying to attach register form handlers');
        
        // First try by ID
        let form = document.getElementById('registerForm');
        
        // If not found by ID, try finding within our container
        if (!form && this._container) {
            form = this._container.querySelector('form');
            console.log('Found form by querying container:', !!form);
        }
        
        if (form) {
            console.log('Register form found, attaching event handlers');
            
            // Store reference to this component on the form
            form._component = this;
            
            // Remove any existing event listeners first to prevent duplicates
            const newForm = form.cloneNode(true);
            form.parentNode.replaceChild(newForm, form);
            form = newForm;
            
            // Add the submit handler
            form.addEventListener('submit', (e) => {
                console.log('Register form submit triggered');
                this.handleRegister(e);
            });
            
            // Add input handlers for all fields
            const inputFields = [
                'nickname', 'email', 'password', 'confirmPassword', 
                'firstName', 'lastName', 'age', 'gender'
            ];
            
            inputFields.forEach(field => {
                const input = form.querySelector(`#${field}`);
                if (input) {
                    input._component = this; // Store reference to component
                    const eventType = input.tagName === 'SELECT' ? 'change' : 'input';
                    
                    input.addEventListener(eventType, (e) => {
                        console.log(`Input change on ${field}:`, e.target.value);
                        this.handleInputChange(field, e.target.value);
                    });
                } else {
                    console.warn(`Field ${field} not found in register form`);
                }
            });
            
            // Add social login button handlers
            const googleBtn = form.querySelector('#google-register');
            const githubBtn = form.querySelector('#github-register');
            
            if (googleBtn) {
                googleBtn.addEventListener('click', (e) => {
                    e.preventDefault();
                    this.handleGoogleLogin();
                });
            }
            
            if (githubBtn) {
                githubBtn.addEventListener('click', (e) => {
                    e.preventDefault();
                    this.handleGithubLogin();
                });
            }
            
            console.log('All register form handlers attached successfully');
        } else {
            console.warn('Register form not found in DOM (tried both ID and container query)');
        }
    }

    handleInputChange(field, value) {
        this.setState({ [field]: value }, () => {
            // Clear error when user types
            if (this.state.error) {
                this.setState({ error: null });
            }
            
            // Do real-time validation
            if (field === 'password' || field === 'confirmPassword') {
                const validation = this.validateForm();
                
                // Show password mismatch immediately
                if (field === 'confirmPassword' && this.state.password !== value) {
                    this.setState({ passwordMismatch: true });
                } else if (field === 'confirmPassword') {
                    this.setState({ passwordMismatch: false });
                }
                
                // Update strength indicator for password
                if (field === 'password') {
                    const strength = this.calculatePasswordStrength(value);
                    this.setState({ 
                        passwordStrength: strength,
                        passwordStrengthInfo: this.getPasswordStrengthInfo(strength)
                    });
                }
            }
        });
    }

    handleRegister(e) {
        if (e) e.preventDefault();
        console.log('Register handler called');
        
        // Validate form data
        const validation = this.validateForm();
        if (!validation.valid) {
            this.setState({ 
                error: validation.errors.join('. '),
                formErrors: validation.errors 
            });
            return;
        }
        
        this.setState({ loading: true, error: null, formErrors: [] });
        
        const { nickname, email, password, firstName, lastName, age, gender } = this.state;
        
        // Create user data object
        const userData = {
            nickname,
            email,
            password,
            firstName: firstName || '',
            lastName: lastName || '',
            age: age ? parseInt(age, 10) : 0,
            gender: gender || 'prefer_not_to_say'
        };
        
        console.log('Attempting to register with:', { nickname, email });
        
        // Add CSRF token if available
        const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
        const headers = csrfToken ? { 'X-CSRF-Token': csrfToken } : {};
        
        // Call the register method from AuthService
        AuthService.register(userData, headers)
            .then(() => {
                console.log('Registration successful');
                this.setState({ 
                    success: true, 
                    loading: false,
                    error: null,
                    formErrors: []
                });
                // Redirect to login after successful registration (with a delay)
                setTimeout(() => {
                    window.location.href = window.location.origin + '/#/login';
                }, 2000);
            })
            .catch(error => {
                console.error('Registration error:', error);
                this.setState({ 
                    error: error.message || 'Registration failed. Please try again.',
                    loading: false,
                    formErrors: error.errors || []
                });
            });
    }

    validateForm() {
        const { nickname, email, password, confirmPassword, firstName, lastName, age } = this.state;
        const errors = [];
        
        // Check required fields
        if (!nickname || !email || !password || !confirmPassword) {
            errors.push('All required fields must be filled');
        }
        
        // Nickname validation
        if (nickname) {
            if (nickname.length < 3) {
                errors.push('Username must be at least 3 characters long');
            }
            if (!/^[a-zA-Z0-9_]+$/.test(nickname)) {
                errors.push('Username can only contain letters, numbers, and underscores');
            }
        }
        
        // Email validation
        if (email) {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            if (!emailRegex.test(email)) {
                errors.push('Please enter a valid email address');
            }
        }
        
        // Password validation
        if (password) {
            if (password.length < 8) {
                errors.push('Password must be at least 8 characters long');
            }
            if (!/[A-Z]/.test(password)) {
                errors.push('Password must contain at least one uppercase letter');
            }
            if (!/[a-z]/.test(password)) {
                errors.push('Password must contain at least one lowercase letter');
            }
            if (!/[0-9]/.test(password)) {
                errors.push('Password must contain at least one number');
            }
            if (!/[^A-Za-z0-9]/.test(password)) {
                errors.push('Password must contain at least one special character');
            }
        }
        
        // Confirm password
        if (password && confirmPassword && password !== confirmPassword) {
            errors.push('Passwords do not match');
        }
        
        // Age validation
        if (age) {
            const ageNum = parseInt(age, 10);
            if (isNaN(ageNum) || ageNum < 13 || ageNum > 120) {
                errors.push('Age must be between 13 and 120');
            }
        }
        
        return { 
            valid: errors.length === 0,
            errors: errors
        };
    }

    // Calculate password strength
    calculatePasswordStrength(password) {
        if (!password) return 0;
        
        let strength = 0;
        
        // Length check
        if (password.length >= 8) strength += 1;
        if (password.length >= 12) strength += 1;
        
        // Character variety checks
        if (/[A-Z]/.test(password)) strength += 1;
        if (/[a-z]/.test(password)) strength += 1;
        if (/[0-9]/.test(password)) strength += 1;
        if (/[^A-Za-z0-9]/.test(password)) strength += 1;
        
        // Prevent easy sequences
        if (/123|abc|qwerty|password/i.test(password)) strength -= 1;
        
        // Clamp to range 0-5
        return Math.max(0, Math.min(5, strength));
    }

    // Get password strength text and color
    getPasswordStrengthInfo(strength) {
        const strengthInfo = [
            { text: 'Very Weak', color: '#ff0000' },
            { text: 'Weak', color: '#ff4500' },
            { text: 'Fair', color: '#ffa500' },
            { text: 'Good', color: '#9acd32' },
            { text: 'Strong', color: '#008000' },
            { text: 'Very Strong', color: '#006400' }
        ];
        
        return strengthInfo[strength] || strengthInfo[0];
    }

    // Handle Google login
    handleGoogleLogin() {
        console.log('Google registration clicked');
        AuthService.initiateGoogleAuth();
    }
    
    // Handle GitHub login
    handleGithubLogin() {
        console.log('GitHub registration clicked');
        AuthService.initiateGithubAuth();
    }

    addDirectDomTest() {
        // Add a small indicator to show the component is mounted
        const testIndicator = document.createElement('div');
        testIndicator.style.position = 'fixed';
        testIndicator.style.top = '80px';
        testIndicator.style.right = '10px';
        testIndicator.style.backgroundColor = 'blue';
        testIndicator.style.color = 'white';
        testIndicator.style.padding = '5px 10px';
        testIndicator.style.borderRadius = '3px';
        testIndicator.style.zIndex = '10000';
        testIndicator.style.fontSize = '12px';
        testIndicator.innerHTML = 'Register Component Mounted';
        document.body.appendChild(testIndicator);
    }

    renderContent() {
        console.log('RegisterComponent renderContent called');
        
        // Create a container element for our form
        const container = document.createElement('div');
        container.className = 'auth-container';
        
        // Get password strength information
        const { passwordStrengthInfo } = this.state;
        
        // Set the HTML content directly
        container.innerHTML = `
            <div class="auth-card">
                <h2 class="auth-title">Create an Account</h2>
                
                ${this.state.error ? `<div class="error-message">${this.state.error}</div>` : ''}
                ${this.state.formErrors.length > 0 ? `
                    <div class="error-list">
                        <ul>
                            ${this.state.formErrors.map(err => `<li>${err}</li>`).join('')}
                        </ul>
                    </div>
                ` : ''}
                ${this.state.success ? `<div class="success-message">Registration successful! Redirecting to login...</div>` : ''}
                
                <form id="registerForm" class="auth-form">
                    <div class="form-group">
                        <label for="nickname">Username <span class="required">*</span></label>
                        <input 
                            type="text" 
                            id="nickname" 
                            value="${this.state.nickname}"
                            placeholder="Choose a username"
                            required
                        >
                        <small class="form-hint">Username must be at least 3 characters and can only contain letters, numbers, and underscores.</small>
                    </div>
                    
                    <div class="form-group">
                        <label for="email">Email <span class="required">*</span></label>
                        <input 
                            type="email" 
                            id="email" 
                            value="${this.state.email}"
                            placeholder="Enter your email"
                            required
                        >
                    </div>
                    
                    <div class="form-group">
                        <label for="password">Password <span class="required">*</span></label>
                        <input 
                            type="password" 
                            id="password" 
                            value="${this.state.password}"
                            placeholder="Create a password"
                            required
                        >
                        <div class="password-strength-meter">
                            <div class="strength-bar">
                                <div class="strength-indicator" style="width: ${(this.state.passwordStrength / 5) * 100}%; background-color: ${passwordStrengthInfo.color}"></div>
                            </div>
                            <div class="strength-text" style="color: ${passwordStrengthInfo.color}">${passwordStrengthInfo.text}</div>
                        </div>
                        <small class="form-text">Password must be at least 8 characters with uppercase letter, lowercase letter, number, and special character</small>
                    </div>
                    
                    <div class="form-group">
                        <label for="confirmPassword">Confirm Password <span class="required">*</span></label>
                        <input 
                            type="password" 
                            id="confirmPassword" 
                            value="${this.state.confirmPassword}"
                            placeholder="Confirm your password"
                            required
                            class="${this.state.passwordMismatch ? 'input-error' : ''}"
                        >
                        ${this.state.passwordMismatch ? '<small class="input-error-text">Passwords do not match</small>' : ''}
                    </div>
                    
                    <div class="form-group">
                        <label for="firstName">First Name</label>
                        <input 
                            type="text" 
                            id="firstName" 
                            value="${this.state.firstName}"
                            placeholder="Enter your first name"
                        >
                    </div>
                    
                    <div class="form-group">
                        <label for="lastName">Last Name</label>
                        <input 
                            type="text" 
                            id="lastName" 
                            value="${this.state.lastName}"
                            placeholder="Enter your last name"
                        >
                    </div>
                    
                    <div class="form-row">
                        <div class="form-group half">
                            <label for="age">Age</label>
                            <input 
                                type="number" 
                                id="age" 
                                value="${this.state.age}"
                                placeholder="Your age"
                                min="13"
                                max="120"
                            >
                            <small class="form-hint">Must be 13 or older</small>
                        </div>
                        
                        <div class="form-group half">
                            <label for="gender">Gender</label>
                            <select id="gender">
                                <option value="male" ${this.state.gender === 'male' ? 'selected' : ''}>Male</option>
                                <option value="female" ${this.state.gender === 'female' ? 'selected' : ''}>Female</option>
                                <option value="other" ${this.state.gender === 'other' ? 'selected' : ''}>Other</option>
                                <option value="prefer_not_to_say" ${this.state.gender === 'prefer_not_to_say' ? 'selected' : ''}>Prefer not to say</option>
                            </select>
                        </div>
                    </div>
                    
                    <div class="form-group">
                        <div class="checkbox-group">
                            <input type="checkbox" id="terms" required>
                            <label for="terms">I agree to the <a href="#/terms" target="_blank">Terms of Service</a> and <a href="#/privacy" target="_blank">Privacy Policy</a></label>
                        </div>
                    </div>
                    
                    <div class="form-actions">
                        <button 
                            type="submit" 
                            class="btn btn-primary"
                            ${this.state.loading || this.state.success ? 'disabled' : ''}>
                            ${this.state.loading ? 'Registering...' : 'Register'}
                        </button>
                    </div>
                    
                    <div class="social-login">
                        <p>Or register with</p>
                        <div class="social-buttons">
                            <button id="google-register" class="btn social-btn google-btn">
                                <img src="/static/images/google-icon.svg" alt="Google" onerror="this.src='/static/images/google-icon.png'">
                                Google
                            </button>
                            <button id="github-register" class="btn social-btn github-btn">
                                <img src="/static/images/github-icon.svg" alt="GitHub" onerror="this.src='/static/images/github-icon.png'">
                                GitHub
                            </button>
                        </div>
                    </div>
                    
                    <div class="auth-links">
                        <a href="#/login">Already have an account? Login</a>
                    </div>
                </form>
            </div>
        `;
        
        return container;
    }
}

// Register component
ComponentRegistry.register('RegisterComponent', RegisterComponent);

// Export the component
export default RegisterComponent; 