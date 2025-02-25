document.addEventListener('DOMContentLoaded', function () {
    const copyButton = document.querySelector('.btn-copy');
    const addRemoveButtons = document.querySelectorAll('.btn-add-playlist, .btn-remove-playlist');
    const backButton = document.querySelector('.btn-back');

    copyButton?.addEventListener('click', async () => {
        const lyrics = document.querySelector('.lyrics-pre').textContent;
        try {
            await navigator.clipboard.writeText(lyrics);
            showToast('Lyrics copied to clipboard!', 'success');
        } catch (err) {
            console.error('Failed to copy:', err);
            showToast('Failed to copy lyrics to clipboard', 'error');
        }
    });

    addRemoveButtons.forEach(button => {
        button.addEventListener('click', function (e) {
            e.preventDefault();
            const url = new URL(e.target.href);
            const currentQuery = new URLSearchParams(window.location.search);

            if (currentQuery.has('query')) {
                url.searchParams.append('query', currentQuery.get('query'));
            }
            if (currentQuery.has('page')) {
                url.searchParams.append('page', currentQuery.get('page'));
            }

            fetch(url, {
                redirect: 'follow'
            })
            .then(response => {
                if (response.redirected) {
                    // Extract action from redirect URL
                    const redirectUrl = new URL(response.url);
                    const action = redirectUrl.searchParams.get('action');
                    
                    if (action === 'added') {
                        showToast('Added to playlist!', 'success');
                        e.target.textContent = 'Remove from Playlist';
                        e.target.classList.remove('btn-add-playlist');
                        e.target.classList.add('btn-remove-playlist');
                        const songId = url.searchParams.get('id');
                        e.target.href = `/remove-from-playlist?id=${songId}`;
                    } else if (action === 'removed') {
                        showToast('Removed from playlist!', 'success');
                        e.target.textContent = 'Add to Playlist';
                        e.target.classList.remove('btn-remove-playlist');
                        e.target.classList.add('btn-add-playlist');
                        const songId = url.searchParams.get('id');
                        const title = url.searchParams.get('title');
                        const artist = url.searchParams.get('artist');
                        e.target.href = `/add-to-playlist?id=${songId}&title=${title}&artist=${artist}`;
                    } else if (action === 'already_exists') {
                        showToast('Song is already in playlist!', 'info');
                    } else if (action === 'failed') {
                        showToast('Failed to update playlist', 'error');
                    }
                    if (!window.location.pathname.includes('/lyrics')) {
                        window.location.href = response.url;
                    }
                }
            })
            .catch(error => {
                console.error('Failed to update playlist:', error);
                showToast('Failed to update playlist', 'error');
            });
        });
    });

    function showToast(message, status) {
        const toast = document.getElementById('toast');
        if (toast) {
            toast.textContent = message;
            toast.className = 'toast ' + status;
            toast.style.display = 'block';

            setTimeout(function () {
                toast.style.display = 'none';
            }, 3000);
        }
    }
});