// frontend/js/components/navigation.js
const NavigationComponent = {
    // Render navigation bar
    renderNavbar() {
        const nav = document.createElement('nav');
        nav.className = 'navbar';
        
        nav.innerHTML = `
            <div class="container">
                <h1>Real-Time Forum</h1>
                <ul>
                    <li><a href="#" id="home-link">Home</a></li>
                    <li><a href="#" id="profile-link">My Profile</a></li>
                    <li><a href="#" id="logout-link">Logout</a></li>
                </ul>
            </div>
        `;
        
        // Add event listeners
        setTimeout(() => {
            const homeLink = document.getElementById('home-link');
            homeLink.addEventListener('click', (e) => {
                e.preventDefault();
                App.renderHome();
            });
            
            const profileLink = document.getElementById('profile-link');
            profileLink.addEventListener('click', (e) => {
                e.preventDefault();
                App.renderProfile(AuthService.user.id);
            });
            
            const logoutLink = document.getElementById('logout-link');
            logoutLink.addEventListener('click', (e) => {
                e.preventDefault();
                AuthService.logout();
            });
        }, 0);
        
        return nav;
    }
};