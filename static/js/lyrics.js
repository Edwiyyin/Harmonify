document.addEventListener('DOMContentLoaded', () => {
    const favoriteButton = document.querySelector('.btn-favorites');
    const copyButton = document.querySelector('.btn-copy');
    const title = document.getElementById('title').textContent.trim();
    const artist = document.getElementById('artist').textContent.trim();

    copyButton.addEventListener('click', async () => {
        const lyrics = document.querySelector('.lyrics-pre').textContent;
        try {
            await navigator.clipboard.writeText(lyrics);
            copyButton.textContent = 'Copied!';
            setTimeout(() => {
                copyButton.textContent = 'Copy Lyrics';
            }, 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
            alert('Failed to copy lyrics to clipboard');
        }
    });

    favoriteButton.addEventListener('click', async (e) => {
        e.preventDefault();
        const songId = new URLSearchParams(window.location.search).get('id');

        try {
            const response = await fetch('/add-favorite', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    id: songId, 
                    title: title, 
                    artist: artist
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