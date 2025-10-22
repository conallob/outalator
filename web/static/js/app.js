// API Base URL
const API_BASE = '/api/v1';

// State management
const state = {
    currentView: 'list',
    currentOutageId: null,
    outages: [],
    selectedOutageForMerge: null,
    filters: {
        status: '',
        severity: '',
        search: ''
    }
};

// Utility functions
function showMessage(message, type = 'info') {
    const container = document.getElementById('message-container');
    const messageDiv = document.createElement('div');
    messageDiv.className = `message ${type}`;
    messageDiv.textContent = message;
    container.appendChild(messageDiv);

    setTimeout(() => {
        messageDiv.remove();
    }, 5000);
}

function formatDate(dateString) {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleString();
}

function formatRelativeTime(dateString) {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return formatDate(dateString);
}

// API functions
async function fetchOutages() {
    try {
        const params = new URLSearchParams();
        params.append('limit', '100');
        params.append('offset', '0');

        const response = await fetch(`${API_BASE}/outages?${params}`);
        if (!response.ok) throw new Error('Failed to fetch outages');

        const data = await response.json();
        state.outages = data.outages || [];
        return state.outages;
    } catch (error) {
        showMessage('Error fetching outages: ' + error.message, 'error');
        return [];
    }
}

async function fetchOutageById(id) {
    try {
        const response = await fetch(`${API_BASE}/outages/${id}`);
        if (!response.ok) throw new Error('Failed to fetch outage');
        return await response.json();
    } catch (error) {
        showMessage('Error fetching outage: ' + error.message, 'error');
        return null;
    }
}

async function createOutage(data) {
    try {
        const response = await fetch(`${API_BASE}/outages`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to create outage');
        }

        return await response.json();
    } catch (error) {
        showMessage('Error creating outage: ' + error.message, 'error');
        throw error;
    }
}

async function updateOutage(id, data) {
    try {
        const response = await fetch(`${API_BASE}/outages/${id}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to update outage');
        }

        return await response.json();
    } catch (error) {
        showMessage('Error updating outage: ' + error.message, 'error');
        throw error;
    }
}

async function addNote(outageId, content, format = 'plaintext', author = 'Anonymous') {
    try {
        const response = await fetch(`${API_BASE}/outages/${outageId}/notes`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ content, format, author })
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to add note');
        }

        return await response.json();
    } catch (error) {
        showMessage('Error adding note: ' + error.message, 'error');
        throw error;
    }
}

async function addTag(outageId, key, value) {
    try {
        const response = await fetch(`${API_BASE}/outages/${outageId}/tags`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ key, value })
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to add tag');
        }

        return await response.json();
    } catch (error) {
        showMessage('Error adding tag: ' + error.message, 'error');
        throw error;
    }
}

async function searchByTag(key, value) {
    try {
        const params = new URLSearchParams();
        if (key) params.append('key', key);
        if (value) params.append('value', value);

        const response = await fetch(`${API_BASE}/tags/search?${params}`);
        if (!response.ok) throw new Error('Failed to search by tag');

        const data = await response.json();
        return data.outages || [];
    } catch (error) {
        showMessage('Error searching by tag: ' + error.message, 'error');
        return [];
    }
}

// View rendering functions
function renderOutagesList(outages) {
    const listContainer = document.getElementById('outages-list');

    if (!outages || outages.length === 0) {
        listContainer.innerHTML = `
            <div class="empty-state">
                <h3>No outages found</h3>
                <p>Create a new outage to get started</p>
            </div>
        `;
        return;
    }

    listContainer.innerHTML = outages.map(outage => `
        <div class="outage-card severity-${outage.severity}" data-id="${outage.id}" onclick="showOutageDetail('${outage.id}')">
            <h3>${escapeHtml(outage.title)}</h3>
            <div class="meta">
                <span class="badge status-${outage.status}">${outage.status}</span>
                <span class="badge severity-${outage.severity}">${outage.severity}</span>
                <span class="timestamp">Created ${formatRelativeTime(outage.created_at)}</span>
            </div>
            ${outage.description ? `<div class="description">${escapeHtml(truncate(outage.description, 150))}</div>` : ''}
            <div class="stats">
                <span>📝 ${outage.notes?.length || 0} notes</span>
                <span>🏷️ ${outage.tags?.length || 0} tags</span>
                <span>🚨 ${outage.alerts?.length || 0} alerts</span>
            </div>
        </div>
    `).join('');
}

function renderOutageDetail(outage) {
    const detailContainer = document.getElementById('outage-detail');

    detailContainer.innerHTML = `
        <div class="detail-container">
            <div class="detail-header">
                <h2>${escapeHtml(outage.title)}</h2>
                <div class="detail-meta">
                    <span class="badge status-${outage.status}">${outage.status}</span>
                    <span class="badge severity-${outage.severity}">${outage.severity}</span>
                </div>
                <div class="timestamp">
                    <div>Created: ${formatDate(outage.created_at)}</div>
                    <div>Updated: ${formatDate(outage.updated_at)}</div>
                </div>
                ${outage.description ? `<p style="margin-top: 15px;">${escapeHtml(outage.description)}</p>` : ''}

                <div class="update-status-form">
                    <label>Update Status:</label>
                    <select id="update-status-select">
                        <option value="open" ${outage.status === 'open' ? 'selected' : ''}>Open</option>
                        <option value="investigating" ${outage.status === 'investigating' ? 'selected' : ''}>Investigating</option>
                        <option value="resolved" ${outage.status === 'resolved' ? 'selected' : ''}>Resolved</option>
                        <option value="closed" ${outage.status === 'closed' ? 'selected' : ''}>Closed</option>
                    </select>
                    <button class="btn-primary" onclick="updateOutageStatus()">Update</button>
                </div>

                <div class="detail-actions">
                    <button class="btn-primary" onclick="showMergeModal()">Merge Outages</button>
                </div>
            </div>

            <div class="section">
                <h3>
                    Tags
                    <button class="btn-secondary" onclick="toggleAddTagForm()">+ Add Tag</button>
                </h3>
                <div class="tags-container" id="tags-container">
                    ${renderTags(outage.tags)}
                </div>
                <div id="add-tag-form" style="display: none;">
                    <div class="add-tag-form">
                        <input type="text" id="tag-key-input" placeholder="Key (e.g., jira)">
                        <input type="text" id="tag-value-input" placeholder="Value (e.g., OPS-123)">
                        <button class="btn-primary" onclick="submitAddTag()">Add</button>
                        <button class="btn-secondary" onclick="toggleAddTagForm()">Cancel</button>
                    </div>
                </div>
            </div>

            <div class="section">
                <h3>Notes (${outage.notes?.length || 0})</h3>
                <div class="notes-list">
                    ${renderNotes(outage.notes)}
                </div>
                <div class="add-note-form">
                    <textarea id="note-content" placeholder="Add troubleshooting notes..."></textarea>
                    <div class="form-row">
                        <input type="text" id="note-author" placeholder="Your name (optional)" value="Anonymous">
                        <select id="note-format">
                            <option value="plaintext">Plain Text</option>
                            <option value="markdown">Markdown</option>
                        </select>
                        <button class="btn-primary" onclick="submitAddNote()">Add Note</button>
                    </div>
                </div>
            </div>

            ${outage.alerts?.length > 0 ? `
                <div class="section">
                    <h3>Alerts (${outage.alerts.length})</h3>
                    <div class="notes-list">
                        ${renderAlerts(outage.alerts)}
                    </div>
                </div>
            ` : ''}
        </div>
    `;
}

function renderTags(tags) {
    if (!tags || tags.length === 0) {
        return '<div class="empty-state" style="padding: 20px;">No tags yet</div>';
    }

    return tags.map(tag => `
        <div class="tag">
            <span class="tag-key">${escapeHtml(tag.key)}</span>
            <span>:</span>
            <span>${escapeHtml(tag.value)}</span>
            <button class="remove-tag" onclick="removeTag('${tag.key}')" title="Remove tag">&times;</button>
        </div>
    `).join('');
}

function renderNotes(notes) {
    if (!notes || notes.length === 0) {
        return '<div class="empty-state" style="padding: 20px;">No notes yet</div>';
    }

    return notes.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
        .map(note => `
        <div class="note">
            <div class="note-header">
                <span><strong>${escapeHtml(note.author || 'Anonymous')}</strong></span>
                <span>${formatRelativeTime(note.created_at)}</span>
            </div>
            <div class="note-content ${note.format === 'markdown' ? 'markdown' : ''}">
                ${note.format === 'markdown' ? renderMarkdown(note.content) : escapeHtml(note.content)}
            </div>
        </div>
    `).join('');
}

function renderAlerts(alerts) {
    if (!alerts || alerts.length === 0) return '';

    return alerts.map(alert => `
        <div class="note">
            <div class="note-header">
                <span><strong>${escapeHtml(alert.source || 'Unknown')}</strong> - ${escapeHtml(alert.team_name || 'N/A')}</span>
                <span>${formatRelativeTime(alert.created_at)}</span>
            </div>
            <div class="note-content">
                <strong>${escapeHtml(alert.title)}</strong><br>
                ${alert.description ? escapeHtml(alert.description) : ''}
                ${alert.severity ? `<br><span class="badge severity-${alert.severity}">${alert.severity}</span>` : ''}
            </div>
        </div>
    `).join('');
}

// View management
function showView(viewName) {
    document.querySelectorAll('.view').forEach(view => view.classList.remove('active'));
    document.querySelectorAll('.nav-btn').forEach(btn => btn.classList.remove('active'));

    document.getElementById(`${viewName}-view`).classList.add('active');
    const navBtn = document.getElementById(`nav-${viewName}`);
    if (navBtn) navBtn.classList.add('active');

    state.currentView = viewName;
}

async function showOutageDetail(id) {
    state.currentOutageId = id;
    showView('detail');

    document.getElementById('outage-detail').innerHTML = '<div class="loading">Loading outage details...</div>';

    const outage = await fetchOutageById(id);
    if (outage) {
        renderOutageDetail(outage);
    }
}

async function refreshCurrentView() {
    if (state.currentView === 'list') {
        await loadOutages();
    } else if (state.currentView === 'detail' && state.currentOutageId) {
        await showOutageDetail(state.currentOutageId);
    }
}

async function loadOutages() {
    const listContainer = document.getElementById('outages-list');
    listContainer.innerHTML = '<div class="loading">Loading outages...</div>';

    const outages = await fetchOutages();
    const filtered = filterOutages(outages);
    renderOutagesList(filtered);
}

function filterOutages(outages) {
    return outages.filter(outage => {
        if (state.filters.status && outage.status !== state.filters.status) return false;
        if (state.filters.severity && outage.severity !== state.filters.severity) return false;
        if (state.filters.search) {
            const search = state.filters.search.toLowerCase();
            return outage.title.toLowerCase().includes(search) ||
                   (outage.description && outage.description.toLowerCase().includes(search));
        }
        return true;
    });
}

// Action handlers
async function updateOutageStatus() {
    const newStatus = document.getElementById('update-status-select').value;
    try {
        await updateOutage(state.currentOutageId, { status: newStatus });
        showMessage('Status updated successfully', 'success');
        await refreshCurrentView();
    } catch (error) {
        // Error already shown by updateOutage
    }
}

async function submitAddNote() {
    const content = document.getElementById('note-content').value.trim();
    const author = document.getElementById('note-author').value.trim() || 'Anonymous';
    const format = document.getElementById('note-format').value;

    if (!content) {
        showMessage('Please enter note content', 'error');
        return;
    }

    try {
        await addNote(state.currentOutageId, content, format, author);
        showMessage('Note added successfully', 'success');
        document.getElementById('note-content').value = '';
        await refreshCurrentView();
    } catch (error) {
        // Error already shown by addNote
    }
}

function toggleAddTagForm() {
    const form = document.getElementById('add-tag-form');
    form.style.display = form.style.display === 'none' ? 'block' : 'none';
}

async function submitAddTag() {
    const key = document.getElementById('tag-key-input').value.trim();
    const value = document.getElementById('tag-value-input').value.trim();

    if (!key || !value) {
        showMessage('Please enter both key and value', 'error');
        return;
    }

    try {
        await addTag(state.currentOutageId, key, value);
        showMessage('Tag added successfully', 'success');
        document.getElementById('tag-key-input').value = '';
        document.getElementById('tag-value-input').value = '';
        toggleAddTagForm();
        await refreshCurrentView();
    } catch (error) {
        // Error already shown by addTag
    }
}

async function removeTag(key) {
    // Note: The API doesn't have a delete tag endpoint, so we'll need to update the outage
    // For now, show a message that this feature requires backend support
    showMessage('Remove tag functionality requires backend API support', 'error');
}

async function showMergeModal() {
    const modal = document.getElementById('merge-modal');
    const listContainer = document.getElementById('merge-outage-list');

    // Load all outages except current one
    const outages = state.outages.filter(o => o.id !== state.currentOutageId);

    if (outages.length === 0) {
        showMessage('No other outages available to merge', 'error');
        return;
    }

    listContainer.innerHTML = outages.map(outage => `
        <label class="merge-checkbox">
            <input type="checkbox" value="${outage.id}">
            <div>
                <strong>${escapeHtml(outage.title)}</strong>
                <span class="badge status-${outage.status}">${outage.status}</span>
                <span class="badge severity-${outage.severity}">${outage.severity}</span>
            </div>
        </label>
    `).join('');

    modal.classList.add('active');
}

function closeMergeModal() {
    document.getElementById('merge-modal').classList.remove('active');
}

async function confirmMerge() {
    const checkboxes = document.querySelectorAll('#merge-outage-list input[type="checkbox"]:checked');
    const outageIds = Array.from(checkboxes).map(cb => cb.value);

    if (outageIds.length === 0) {
        showMessage('Please select at least one outage to merge', 'error');
        return;
    }

    try {
        // Fetch the selected outages to merge their data
        const outagePromises = outageIds.map(id => fetchOutageById(id));
        const outages = await Promise.all(outagePromises);

        // Merge notes and tags from selected outages into current outage
        for (const outage of outages) {
            // Add all notes
            if (outage.notes) {
                for (const note of outage.notes) {
                    await addNote(
                        state.currentOutageId,
                        `[Merged from: ${outage.title}]\n\n${note.content}`,
                        note.format,
                        note.author
                    );
                }
            }

            // Add all tags
            if (outage.tags) {
                for (const tag of outage.tags) {
                    try {
                        await addTag(state.currentOutageId, tag.key, tag.value);
                    } catch (error) {
                        // Tag might already exist, continue
                    }
                }
            }
        }

        showMessage(`Successfully merged ${outageIds.length} outage(s)`, 'success');
        closeMergeModal();
        await refreshCurrentView();
    } catch (error) {
        showMessage('Error merging outages: ' + error.message, 'error');
    }
}

// Create form handler
async function handleCreateForm(e) {
    e.preventDefault();

    const title = document.getElementById('new-title').value.trim();
    const description = document.getElementById('new-description').value.trim();
    const status = document.getElementById('new-status').value;
    const severity = document.getElementById('new-severity').value;

    if (!title) {
        showMessage('Please enter a title', 'error');
        return;
    }

    try {
        const outage = await createOutage({ title, description, status, severity });
        showMessage('Outage created successfully', 'success');

        // Reset form
        document.getElementById('create-form').reset();

        // Reload outages and show detail view
        await fetchOutages();
        await showOutageDetail(outage.id);
    } catch (error) {
        // Error already shown by createOutage
    }
}

// Search and filter handlers
async function handleSearch() {
    state.filters.search = document.getElementById('search-input').value.trim();
    await loadOutages();
}

async function handleTagSearch() {
    const key = document.getElementById('tag-key').value.trim();
    const value = document.getElementById('tag-value').value.trim();

    if (!key && !value) {
        showMessage('Please enter a tag key or value', 'error');
        return;
    }

    const listContainer = document.getElementById('outages-list');
    listContainer.innerHTML = '<div class="loading">Searching...</div>';

    const outages = await searchByTag(key, value);
    state.outages = outages;
    renderOutagesList(outages);
}

function handleFilterChange() {
    state.filters.status = document.getElementById('filter-status').value;
    state.filters.severity = document.getElementById('filter-severity').value;
    loadOutages();
}

function clearFilters() {
    state.filters = { status: '', severity: '', search: '' };
    document.getElementById('filter-status').value = '';
    document.getElementById('filter-severity').value = '';
    document.getElementById('search-input').value = '';
    document.getElementById('tag-key').value = '';
    document.getElementById('tag-value').value = '';
    loadOutages();
}

// Utility functions
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function truncate(text, length) {
    if (!text || text.length <= length) return text;
    return text.substring(0, length) + '...';
}

function renderMarkdown(text) {
    // Basic markdown rendering (could be enhanced with a library)
    return escapeHtml(text)
        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.+?)\*/g, '<em>$1</em>')
        .replace(/`(.+?)`/g, '<code>$1</code>')
        .replace(/\n/g, '<br>');
}

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    // Navigation
    document.getElementById('nav-list').addEventListener('click', () => {
        showView('list');
        loadOutages();
    });

    document.getElementById('nav-create').addEventListener('click', () => {
        showView('create');
    });

    document.getElementById('back-to-list').addEventListener('click', () => {
        showView('list');
        loadOutages();
    });

    // Search and filters
    document.getElementById('search-btn').addEventListener('click', handleSearch);
    document.getElementById('search-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') handleSearch();
    });

    document.getElementById('tag-search-btn').addEventListener('click', handleTagSearch);
    document.getElementById('filter-status').addEventListener('change', handleFilterChange);
    document.getElementById('filter-severity').addEventListener('change', handleFilterChange);
    document.getElementById('clear-filters').addEventListener('click', clearFilters);

    // Create form
    document.getElementById('create-form').addEventListener('submit', handleCreateForm);
    document.getElementById('cancel-create').addEventListener('click', () => {
        showView('list');
        loadOutages();
    });

    // Merge modal
    document.getElementById('cancel-merge').addEventListener('click', closeMergeModal);
    document.getElementById('confirm-merge').addEventListener('click', confirmMerge);

    // Close modal on background click
    document.getElementById('merge-modal').addEventListener('click', (e) => {
        if (e.target.id === 'merge-modal') {
            closeMergeModal();
        }
    });

    // Load initial data
    loadOutages();
});
