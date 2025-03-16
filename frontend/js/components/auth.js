// frontend/js/components/auth.js
const AuthComponent = {
    // Render the login form
    renderLogin() {
        const container = document.createElement('div');
        container.className = 'auth-container';
        
        container.innerHTML = `
            <h2>Login to Forum</h2>
            <form id="login-form">
                <div class="form-group">
                    <label for="identifier">Nickname or Email</label>
                    <input type="text" id="identifier" name="identifier" required>
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required>
                </div>
                <button type="submit" class="btn-primary">Login</button>
            </form>
            <div class="auth-switch">
                <p>Don't have an account? <a href="#" id="register-link">Register</a></p>
            </div>
        `;
        
        // Add event listeners
        setTimeout(() => {
            const form = document.getElementById('login-form');
            form.addEventListener('submit', this.handleLogin);
            
            const registerLink = document.getElementById('register-link');
            registerLink.addEventListener('click', (e) => {
                e.preventDefault();
                App.renderRegister();
            });
        }, 0);
        
        return container;
    },
    
    // Render the registration form
    renderRegister() {
        const container = document.createElement('div');
        container.className = 'auth-container';
        
        container.innerHTML = `
            <h2>Create an Account</h2>
            <form id="register-form">
                <div class="form-group">
                    <label for="nickname">Nickname</label>
                    <input type="text" id="nickname" name="nickname" required>
                </div>
                <div class="form-group">
                    <label for="age">Age</label>
                    <input type="number" id="age" name="age" min="13" required>
                </div>
                <div class="form-group">
                    <label for="gender">Gender</label>
                    <select id="gender" name="gender" required>
                        <option value="">Select your gender</option>
                        <option value="male">Male</option>
                        <option value="female">Female</option>
                        <option value="other">Other</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="firstName">First Name</label>
                    <input type="text" id="firstName" name="firstName" required>
                </div>
                <div class="form-group">
                    <label for="lastName">Last Name</label>
                    <input type="text" id="lastName" name="lastName" required>
                </div>
                <div class="form-group">
                    <label for="email">Email</label>
                    <input type="email" id="email" name="email" required>
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required>
                </div>
                <button type="submit" class="btn-primary">Register</button>
            </form>
            <div class="auth-switch">
                <p>Already have an account? <a href="#" id="login-link">Login</a></p>
            </div>
        `;
        
        // Add event listeners
        setTimeout(() => {
            const form = document.getElementById('register-form');
            form.addEventListener('submit', this.handleRegister);
            
            const loginLink = document.getElementById('login-link');
            loginLink.addEventListener('click', (e) => {
                e.preventDefault();
                App.renderLogin();
            });
        }, 0);
        
        return container;
    },
    
    // Handle login form submission
    async handleLogin(e) {
        e.preventDefault();
        
        const identifier = document.getElementById('identifier').value;
        const password = document.getElementById('password').value;
        
        try {
            await AuthService.login({ identifier, password });
            App.renderHome();
        } catch (error) {
            alert('Login failed: ' + error.message);
        }
    },
    
    // Handle registration form submission
    async handleRegister(e) {
        e.preventDefault();
        
        const userData = {
            nickname: document.getElementById('nickname').value,
            age: parseInt(document.getElementById('age').value),
            gender: document.getElementById('gender').value,
            firstName: document.getElementById('firstName').value,
            lastName: document.getElementById('lastName').value,
            email: document.getElementById('email').value,
            password: document.getElementById('password').value
        };
        
        try {
            await AuthService.register(userData);
            App.renderHome();
        } catch (error) {
            alert('Registration failed: ' + error.message);
        }
    }
};