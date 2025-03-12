// LoginComponent.js - Handles user login functionality
import Component from '../../core/Component.js';
import ComponentRegistry from '../../core/ComponentRegistry.js';

class LoginComponent extends Component {
    constructor() {
        super();
        this.state = {
            identifier: '',
            password: '',
            error: null,
            loading: false
        };
        
        // Remove excessive debug logs but keep essential ones
        console.log('LoginComponent constructed');
    }

    componentDidMount() {
        console.log('LoginComponent mounted');
        // Add a direct DOM test to verify component is mounted
        this.addDirectDomTest();
        
        // We'll attach events in afterRender
    }

    // This is called after the component has been rendered
    afterRender() {
        console.log('LoginComponent afterRender');
        
        // Use a small delay to ensure DOM is fully updated
        setTimeout(() => {
            this.attachFormEventHandlers();
        }, 50);
    }

    attachFormEventHandlers() {
        console.log('Trying to attach login form handlers');
        
        // First try by ID
        let form = document.getElementById('loginForm');
        
        // If not found by ID, try finding within our container
        if (!form && this._container) {
            form = this._container.querySelector('form');
            console.log('Found form by querying container:', !!form);
        }
        
        if (form) {
            console.log('Login form found, attaching event handlers');
            
            // Store reference to this component on the form
            form._component = this;
            
            // Remove any existing event listeners to prevent duplicates
            const newForm = form.cloneNode(true);
            form.parentNode.replaceChild(newForm, form);
            form = newForm;
            
            // Add the submit handler
            form.addEventListener('submit', (e) => {
                console.log('Login form submit triggered');
                this.handleLogin(e);
            });
            
            // Add input handlers for username and password
            const usernameInput = form.querySelector('#identifier');
            const passwordInput = form.querySelector('#password');
            
            if (usernameInput) {
                usernameInput._component = this;
                usernameInput.addEventListener('input', (e) => {
                    console.log('Username input change:', e.target.value);
                    this.handleInputChange('identifier', e.target.value);
                });
            } else {
                console.warn('Username field not found in login form');
            }
            
            if (passwordInput) {
                passwordInput._component = this;
                passwordInput.addEventListener('input', (e) => {
                    console.log('Password input change:', e.target.value);
                    this.handleInputChange('password', e.target.value);
                });
            } else {
                console.warn('Password field not found in login form');
            }
            
            // Add social login button handlers
            const googleBtn = form.querySelector('#google-login');
            const githubBtn = form.querySelector('#github-login');
            
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
            
            console.log('All login form handlers attached successfully');
        } else {
            console.warn('Login form not found in DOM (tried both ID and container query)');
        }
    }

    handleInputChange(field, value) {
        this.setState({ [field]: value });
    }

    handleLogin(e) {
        if (e) e.preventDefault();
        console.log('Login handler called');
        
        const { identifier, password } = this.state;
        
        // Basic validation
        if (!identifier || !password) {
            this.setState({
                error: 'Please enter both username and password.'
            });
            return;
        }
        
        this.setState({ loading: true, error: null });
        
        console.log('Attempting to login with identifier:', identifier);
        
        // Add CSRF token if available
        const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
        const headers = csrfToken ? { 'X-CSRF-Token': csrfToken } : {};
        
        // Call login method from AuthService
        AuthService.login(identifier, password, headers)
            .then(response => {
                console.log('Login successful');
                this.setState({ loading: false });
                
                // Store intended destination for redirection
                const redirectPath = sessionStorage.getItem('redirectAfterLogin') || '/';
                
                // Redirect to the intended destination
                window.location.hash = `#${redirectPath}`;
                sessionStorage.removeItem('redirectAfterLogin');
            })
            .catch(error => {
                console.error('Login error:', error);
                this.setState({
                    error: error.message || 'Login failed. Please check your credentials.',
                    loading: false
                });
            });
    }
    
    // Handle Google login
    handleGoogleLogin() {
        console.log('Google login clicked');
        AuthService.initiateGoogleAuth();
    }
    
    // Handle GitHub login
    handleGithubLogin() {
        console.log('GitHub login clicked');
        AuthService.initiateGithubAuth();
    }

    // Add a visual indicator to confirm component is mounted
    addDirectDomTest() {
        if (!this._container) return;
        
        const indicator = document.createElement('div');
        indicator.textContent = 'Login Component Mounted';
        indicator.style.position = 'fixed';
        indicator.style.top = '10px';
        indicator.style.right = '10px';
        indicator.style.backgroundColor = '#4CAF50';
        indicator.style.color = 'white';
        indicator.style.padding = '5px 10px';
        indicator.style.borderRadius = '4px';
        indicator.style.zIndex = '9999';
        document.body.appendChild(indicator);
    }

    renderContent() {
        const { identifier, password, error, loading } = this.state;
        
        // Create a more robust unique ID for the form
        const formId = 'loginForm_' + Date.now().toString().slice(-4);
        
        return `
            <div class="auth-container">
                <h1>Login to Your Account</h1>
                
                ${error ? `<div class="error-message">${error}</div>` : ''}
                
                <form id="${formId}" class="login-form">
                    <div class="form-group">
                        <label for="identifier">Username or Email</label>
                        <input type="text" id="identifier" name="identifier" value="${identifier}" placeholder="Enter your username or email" required autofocus>
                    </div>
                    
                    <div class="form-group">
                        <label for="password">Password</label>
                        <input type="password" id="password" name="password" value="${password}" placeholder="Enter your password" required>
                    </div>
                    
                    <button type="submit" class="btn btn-primary" ${loading ? 'disabled' : ''}>
                        ${loading ? 'Logging in...' : 'Login'}
                    </button>
                    
                    <div class="social-login">
                        <p>Or login with</p>
                        <div class="social-buttons">
                            <button id="google-login" class="btn social-btn google-btn">
                                <img src="/static/images/google-icon.svg" alt="Google" onerror="this.src='/static/images/google-icon.png'">
                                Google
                            </button>
                            <button id="github-login" class="btn social-btn github-btn">
                                <img src="/static/images/github-icon.svg" alt="GitHub" onerror="this.src='/static/images/github-icon.png'">
                                GitHub
                            </button>
                        </div>
                    </div>
                </form>
                
                <div class="auth-links">
                    <a href="#/forgot-password">Forgot password?</a>
                    <span>Don't have an account? <a href="#/register">Register</a></span>
                </div>
            </div>
        `;
    }
}

// Register the component
ComponentRegistry.register('LoginComponent', LoginComponent);

// Export the component
export default LoginComponent; 