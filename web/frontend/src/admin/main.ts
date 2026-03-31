// Cadmus Admin - Main JavaScript

// Theme Toggle
function initThemeToggle() {
	const toggleBtn = document.getElementById('theme-toggle');
	if (!toggleBtn) return;

	toggleBtn.addEventListener('click', () => {
		const currentTheme = document.documentElement.getAttribute('data-theme');
		const newTheme = currentTheme === 'dark' ? 'light' : 'dark';

		document.documentElement.setAttribute('data-theme', newTheme);
		localStorage.setItem('theme', newTheme);
	});
}

// Mobile Sidebar Toggle
function initSidebarToggle() {
	const sidebar = document.querySelector('.admin-sidebar');
	if (!sidebar) return;

	// Create toggle button for mobile
	const toggleBtn = document.createElement('button');
	toggleBtn.className = 'sidebar-toggle-btn';
	toggleBtn.innerHTML = `
		<svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6h16M4 12h16M4 18h16"></path>
		</svg>
	`;
	toggleBtn.style.cssText = `
		position: fixed;
		bottom: 24px;
		right: 24px;
		width: 56px;
		height: 56px;
		border-radius: 50%;
		background: var(--accent-primary);
		color: white;
		border: none;
		cursor: pointer;
		display: none;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-lg);
		z-index: 1000;
		transition: transform 0.2s ease;
	`;

	toggleBtn.querySelector('svg').style.cssText = 'width: 24px; height: 24px;';

	// Show on mobile
	const mediaQuery = window.matchMedia('(max-width: 1024px)');
	const updateVisibility = () => {
		toggleBtn.style.display = mediaQuery.matches ? 'flex' : 'none';
	};

	mediaQuery.addEventListener('change', updateVisibility);
	updateVisibility();

	document.body.appendChild(toggleBtn);

	toggleBtn.addEventListener('click', () => {
		sidebar.classList.toggle('open');
	});
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
	initThemeToggle();
	initSidebarToggle();
});

// Export for module usage
export { initThemeToggle, initSidebarToggle };