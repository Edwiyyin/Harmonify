function updateDuration(type) {
    const minutes = parseInt(document.getElementById(type + 'DurationMinutes').value) || 0;
    const seconds = parseInt(document.getElementById(type + 'DurationSeconds').value) || 0;
    const totalSeconds = (minutes * 60) + seconds;
    document.getElementById(type + 'Duration').value = totalSeconds;
}
    document.addEventListener('DOMContentLoaded', () => {
        const favoriteButtons = document.querySelectorAll('.favorite-btn');
        
        favoriteButtons.forEach(button => {
            button.addEventListener('click', async (e) => {
                const songCard = e.target.closest('.song-card');
                const songId = songCard.dataset.songId;
                const title = songCard.querySelector('h2').textContent;
                const artist = songCard.querySelector('p').textContent;
                const coverUrl = songCard.querySelector('.cover-image')?.src || '';

                try {
                    const response = await fetch('/add-favorite', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({ 
                            id: songId, 
                            title: title, 
                            artist: artist, 
                            coverUrl: coverUrl 
                        })
                    });

                    const result = await response.json();
                    alert(result.message);
                } catch (error) {
                    console.error('Error adding favorite:', error);
                    alert('Failed to add to favorites');
                }
            });
        });
    });