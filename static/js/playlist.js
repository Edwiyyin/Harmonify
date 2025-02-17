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
});