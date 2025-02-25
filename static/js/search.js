document.addEventListener('DOMContentLoaded', function () {
    const urlParams = new URLSearchParams(window.location.search);
    const action = urlParams.get('action');

    if (action === 'added') {
        showToast('Added to playlist!', 'success');
    } else if (action === 'already_exists') {
        showToast('Song already in playlist!', 'info');
    } else if (action === 'failed') {
        showToast('Failed to add to playlist. Please try again.', 'error');
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

    function flipCard(card) {
        card.classList.toggle('flipped');
    }

    const songCards = document.querySelectorAll('.song-card');
    songCards.forEach(card => {
        let clickCount = 0;

        card.addEventListener('click', function () {
            clickCount++;
            if (clickCount === 2) {
                flipCard(card);
                clickCount = 0;
            }
        });

        card.addEventListener('mouseleave', function () {
            clickCount = 0;
        });
    });

    const addRemoveButtons = document.querySelectorAll('.btn-add-playlist, .btn-remove-playlist');
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
                    const redirectUrl = new URL(response.url);
                    const action = redirectUrl.searchParams.get('action');
                    
                    if (action === 'added') {
                        showToast('Added to playlist!', 'success');
                        e.target.textContent = 'Remove';
                        e.target.classList.remove('btn-add-playlist');
                        e.target.classList.add('btn-remove-playlist');
                    
                        const songId = url.searchParams.get('id');
                        e.target.href = `/remove-from-playlist?id=${songId}`;
                    } else if (action === 'removed') {
                        showToast('Removed from playlist!', 'success');
                    
                        e.target.textContent = 'Add';
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
                }
            })
            .catch(error => {
                console.error('Failed to update playlist:', error);
                showToast('Failed to update playlist', 'error');
            });
        });
    });

    document.getElementById('minDurationMinutes').addEventListener('input', () => updateDuration('min'));
    document.getElementById('minDurationSeconds').addEventListener('input', () => updateDuration('min'));
    document.getElementById('maxDurationMinutes').addEventListener('input', () => updateDuration('max'));
    document.getElementById('maxDurationSeconds').addEventListener('input', () => updateDuration('max'));

    document.getElementById('filterForm').addEventListener('submit', function (e) {
        updateDuration('min');
        updateDuration('max');
    });

    document.getElementById('clearFilters').addEventListener('click', function () {
        document.getElementById('startDate').value = '';
        document.getElementById('endDate').value = '';
        document.getElementById('minDurationMinutes').value = '';
        document.getElementById('minDurationSeconds').value = '';
        document.getElementById('maxDurationMinutes').value = '';
        document.getElementById('maxDurationSeconds').value = '';
        document.getElementById('minDuration').value = '0';
        document.getElementById('maxDuration').value = '0';
        document.getElementById('lyricsFilter').value = '';
        document.getElementById('playlistFilter').value = 'all';
        document.getElementById('sortOrder').value = 'asc';
        document.getElementById('sortBy').value = 'date';

        document.getElementById('filterForm').submit();
    });
});

function updateDuration(type) {
    const minutes = parseInt(document.getElementById(`${type}DurationMinutes`).value) || 0;
    const seconds = parseInt(document.getElementById(`${type}DurationSeconds`).value) || 0;
    const totalMilliseconds = (minutes * 60 + seconds) * 1000;
    document.getElementById(`${type}Duration`).value = totalMilliseconds;
}