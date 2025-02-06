document.addEventListener('DOMContentLoaded', () => {
    // Handle like/dislike actions
    document.querySelectorAll('.like-button, .dislike-button').forEach(button => {
        button.addEventListener('click', async (e) => {
            e.preventDefault();
            
            const type = button.classList.contains('like-button') ? 'like' : 'dislike';
            const itemType = button.dataset.type; // 'post' or 'comment'
            const itemId = button.dataset.id;
            
            try {
                const response = await fetch(`/api/${itemType}s/like?${itemType}_id=${itemId}&like=${type === 'like'}`, {
                    method: 'POST',
                    credentials: 'same-origin'
                });

                if (!response.ok) {
                    if (response.status === 401) {
                        window.location.href = '/login';
                        return;
                    }
                    throw new Error('Network response was not ok');
                }

                // Update like/dislike count
                const countElement = button.querySelector('.count');
                if (countElement) {
                    const currentCount = parseInt(countElement.textContent);
                    countElement.textContent = currentCount + 1;
                }

                // Toggle active state
                button.classList.toggle('active');
                
                // Remove active state from opposite button
                const oppositeButton = type === 'like' 
                    ? button.nextElementSibling 
                    : button.previousElementSibling;
                if (oppositeButton && oppositeButton.classList.contains('active')) {
                    oppositeButton.classList.remove('active');
                    const oppositeCount = oppositeButton.querySelector('.count');
                    if (oppositeCount) {
                        const currentCount = parseInt(oppositeCount.textContent);
                        oppositeCount.textContent = Math.max(0, currentCount - 1);
                    }
                }
            } catch (error) {
                console.error('Error:', error);
                alert('Failed to process your action. Please try again.');
            }
        });
    });

    // Handle form submissions
    document.querySelectorAll('form[data-ajax="true"]').forEach(form => {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const submitButton = form.querySelector('button[type="submit"]');
            if (submitButton) {
                submitButton.disabled = true;
            }

            try {
                const formData = new FormData(form);
                const response = await fetch(form.action, {
                    method: form.method,
                    body: JSON.stringify(Object.fromEntries(formData)),
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'same-origin'
                });

                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }

                // Handle success (redirect or update UI)
                const redirectUrl = form.dataset.redirectUrl;
                if (redirectUrl) {
                    window.location.href = redirectUrl;
                } else {
                    form.reset();
                    // Optionally update UI to show success message
                }
            } catch (error) {
                console.error('Error:', error);
                alert('Failed to submit form. Please try again.');
            } finally {
                if (submitButton) {
                    submitButton.disabled = false;
                }
            }
        });
    });

    // Handle category filtering
    const categoryFilter = document.getElementById('category-filter');
    if (categoryFilter) {
        categoryFilter.addEventListener('change', () => {
            const category = categoryFilter.value;
            const url = new URL(window.location);
            
            if (category) {
                url.searchParams.set('category', category);
            } else {
                url.searchParams.delete('category');
            }
            
            window.location.href = url.toString();
        });
    }

    // Handle mobile menu toggle
    const menuToggle = document.querySelector('.menu-toggle');
    const navLinks = document.querySelector('.nav-links');
    
    if (menuToggle && navLinks) {
        menuToggle.addEventListener('click', () => {
            navLinks.classList.toggle('active');
            menuToggle.classList.toggle('active');
        });
    }

    // Auto-expand textarea
    document.querySelectorAll('textarea[data-auto-expand]').forEach(textarea => {
        textarea.addEventListener('input', () => {
            textarea.style.height = 'auto';
            textarea.style.height = textarea.scrollHeight + 'px';
        });
    });

    // Show/hide password toggle
    document.querySelectorAll('.password-toggle').forEach(toggle => {
        toggle.addEventListener('click', () => {
            const input = toggle.previousElementSibling;
            const type = input.type === 'password' ? 'text' : 'password';
            input.type = type;
            toggle.textContent = type === 'password' ? 'Show' : 'Hide';
        });
    });
}); 