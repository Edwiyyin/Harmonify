document.addEventListener('DOMContentLoaded', function () {
    const urlParams = new URLSearchParams(window.location.search);
    const action = urlParams.get('action');

    if (action === 'removed') {
        showToast('Removed from playlist!', 'success');
    } else if (action === 'already_exists') {
        showToast('Song already in playlist!', 'info');
    } else if (action === 'not_found') {
        showToast('Song not found in playlist!', 'error');
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
});