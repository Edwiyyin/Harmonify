function updateDuration(type) {
    const minutes = parseInt(document.getElementById(`${type}DurationMinutes`).value) || 0;
    const seconds = parseInt(document.getElementById(`${type}DurationSeconds`).value) || 0;
    
    if (minutes < 0) minutes = 0;
    if (seconds < 0) seconds = 0;
    if (seconds >= 60) seconds = 59;
    
    const totalSeconds = (minutes * 60) + seconds;
    document.getElementById(`${type}Duration`).value = totalSeconds;
    
    document.getElementById(`${type}DurationMinutes`).value = minutes;
    document.getElementById(`${type}DurationSeconds`).value = seconds;
}
document.addEventListener('DOMContentLoaded', function () {
    const urlParams = new URLSearchParams(window.location.search);
    const action = urlParams.get('action');

    if (action === 'added') {
        showToast('Added to playlist!', 'success');
    } else if (action === 'already_exists') {
        showToast('Song already in playlist!', 'info');
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
