document.addEventListener('DOMContentLoaded', function () {
    const copyButton = document.querySelector('.btn-copy');
    const addRemoveButtons = document.querySelectorAll('.btn-add-playlist, .btn-remove-playlist');
    const backButton = document.querySelector('.btn-back');

    copyButton.addEventListener('click', async () => {
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
            const url = e.target.href;

            fetch(url)
                .then(response => {
                    if (response.ok) {
                        const isAddButton = e.target.classList.contains('btn-add-playlist');
                        if (isAddButton) {
                        
                            e.target.textContent = 'Remove from Playlist';
                            e.target.classList.remove('btn-add-playlist');
                            e.target.classList.add('btn-remove-playlist');
                            showToast('Added to playlist!', 'success');
                        } else {
                            
                            e.target.textContent = 'Add to Playlist';
                            e.target.classList.remove('btn-remove-playlist');
                            e.target.classList.add('btn-add-playlist');
                            showToast('Removed from playlist!', 'success');
                        }
                    } else {
                        showToast('Failed to update playlist', 'error');
                    }
                })
                .catch(error => {
                    console.error('Failed to update playlist:', error);
                    showToast('Failed to update playlist', 'error');
                });
        });
    });

    if (backButton) {
        backButton.addEventListener('click', function (e) {
        });
    }

    function showToast(message, status) {
        const toast = document.getElementById('toast');
        toast.textContent = message;
        toast.className = 'toast ' + status;
        toast.style.display = 'block';

        setTimeout(function () {
            toast.style.display = 'none';
        }, 3000);
    }
});