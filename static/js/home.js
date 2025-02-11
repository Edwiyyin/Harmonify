document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.querySelector('input[name="query"]');
    
    searchInput.addEventListener('input', async (e) => {
        const query = e.target.value;
        if (query.length > 2) {
            try {
                const response = await fetch(`/search-suggestions?q=${encodeURIComponent(query)}`);
                const suggestions = await response.json();
                
                let suggestionDropdown = document.getElementById('search-suggestions');
                if (!suggestionDropdown) {
                    suggestionDropdown = document.createElement('div');
                    suggestionDropdown.id = 'search-suggestions';
                    searchInput.parentNode.appendChild(suggestionDropdown);
                }
                
                suggestionDropdown.innerHTML = suggestions.map(
                    suggestion => `<div class="suggestion-item">${suggestion}</div>`
                ).join('');
            } catch (error) {
                console.error('Suggestions error:', error);
            }
        }
    });
});