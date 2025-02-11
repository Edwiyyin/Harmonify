document.addEventListener('DOMContentLoaded', () => {
    const removeFavoriteButtons = document.querySelectorAll('.remove-favorite');
    
    removeFavoriteButtons.forEach(button => {
        button.addEventListener('click', async (e) => {
            const songCard = e.target.closest('.song-card');
            const songId = songCard.dataset.songId;

            try {
                const response = await fetch('/remove-favorite', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ id: songId })
                });

                const result = await response.json();
                if (result.success) {
                    songCard.remove();
                    
                    const remainingFavorites = document.querySelectorAll('.song-card');
                    if (remainingFavorites.length === 0) {
                        const grid = document.querySelector('.results-grid');
                        grid.innerHTML = `
                            <div class="no-favorites">
                                No favorite songs yet. Start searching and add some!
                            </div>
                        `;
                    }
                } else {
                    alert('Failed to remove from favorites');
                }
            } catch (error) {
                console.error('Error removing favorite:', error);
                alert('Failed to remove from favorites');
            }
        });
    });
});