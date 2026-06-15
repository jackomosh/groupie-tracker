// Client-Server Request-Response Pattern: Dynamic Filtering Logic Loop
const searchInput = document.getElementById('search-input');
const artistsGrid = document.getElementById('artists-grid');

if (searchInput) {
    searchInput.addEventListener('input', (e) => {
        const query = e.target.value.trim();

        // Query the live JSON API query handler engine
        fetch(`/api/search?q=${encodeURIComponent(query)}`)
            .then(res => res.json())
            .then(artists => {
                artistsGrid.innerHTML = '';
                if (artists.length === 0) {
                    artistsGrid.innerHTML = '<p style="grid-column: 1/-1; text-align: center; color: var(--muted);">No matching entries discovered across tracking registry.</p>';
                    return;
                }
                artists.forEach(art => {
                    artistsGrid.innerHTML += `
                        <div class="card" onclick="window.location.href='/artist?id=${art.id}'">
                            <img src="${art.image}" alt="${art.name}" loading="lazy">
                            <div class="card-body">
                                <h3>${art.name}</h3>
                                <p class="tag">Since ${art.creationDate}</p>
                            </div>
                        </div>
                    `;
                });
            })
            .catch(err => console.error("Error fetching filtered runtime metrics:", err));
    });
}