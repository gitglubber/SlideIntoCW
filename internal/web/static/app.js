// Global state
let state = {
    slideClients: [],
    cwClients: [],
    mappings: [],
    alerts: [],
    ticketMappings: [],
    boards: [],
    statuses: [],
    priorities: [],
    types: [],
    members: [],
    config: {}
};

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    initModals();
    checkHealth();
    loadDashboard();
    initMappings();
    initTicketing();
    initAlerts();
    initTickets();

    // Auto-refresh dashboard every 30 seconds
    setInterval(loadDashboard, 30000);
});

// Tab navigation
function initTabs() {
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');

    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const tabName = btn.dataset.tab;

            tabBtns.forEach(b => b.classList.remove('active'));
            tabContents.forEach(c => c.classList.remove('active'));

            btn.classList.add('active');
            document.getElementById(tabName).classList.add('active');

            // Load data when switching tabs
            switch(tabName) {
                case 'dashboard':
                    loadDashboard();
                    break;
                case 'mappings':
                    loadMappings();
                    break;
                case 'ticketing':
                    loadTicketingConfig();
                    break;
                case 'alerts':
                    loadAlerts();
                    break;
                case 'tickets':
                    loadTicketMappings();
                    break;
            }
        });
    });
}

// Health check
async function checkHealth() {
    try {
        const response = await fetch('/api/health');
        const data = await response.json();

        if (data.status === 'ok') {
            document.getElementById('healthStatus').innerHTML = `
                <span class="status-indicator"></span>
                <span class="status-text">Service Online</span>
            `;
        }
    } catch (error) {
        document.getElementById('healthStatus').innerHTML = `
            <span class="status-indicator" style="background: var(--danger-color);"></span>
            <span class="status-text">Service Offline</span>
        `;
    }
}

// Dashboard
async function loadDashboard() {
    try {
        const response = await fetch('/api/dashboard');
        const data = await response.json();

        document.getElementById('unresolvedAlerts').textContent = data.unresolvedAlerts;
        document.getElementById('totalAlerts').textContent = `of ${data.totalAlerts} total`;
        document.getElementById('mappedClients').textContent = data.mappedClients;
        document.getElementById('totalClients').textContent = `of ${data.totalClients} total`;
        document.getElementById('openTickets').textContent = data.openTickets;
    } catch (error) {
        console.error('Error loading dashboard:', error);
    }
}

// Mappings
function initMappings() {
    document.getElementById('autoMapBtn').addEventListener('click', autoMapClients);
    document.getElementById('refreshMappingsBtn').addEventListener('click', loadMappings);
    document.getElementById('mappingSearch').addEventListener('input', filterMappings);
}

async function loadMappings() {
    const container = document.getElementById('mappingsList');
    container.innerHTML = '<div class="loading">Loading mappings...</div>';

    try {
        const [mappingsRes, cwClientsRes] = await Promise.all([
            fetch('/api/mappings'),
            fetch('/api/connectwise/clients')
        ]);

        state.mappings = await mappingsRes.json();
        state.cwClients = await cwClientsRes.json();

        renderMappings();
    } catch (error) {
        container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">⚠️</div><p>Error loading mappings</p></div>';
        console.error('Error loading mappings:', error);
    }
}

function renderMappings(filter = '') {
    const container = document.getElementById('mappingsList');

    let mappings = state.mappings;
    if (filter) {
        mappings = mappings.filter(m =>
            m.slideClientName.toLowerCase().includes(filter.toLowerCase()) ||
            (m.connectWiseName && m.connectWiseName.toLowerCase().includes(filter.toLowerCase()))
        );
    }

    if (mappings.length === 0) {
        container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">📋</div><p>No mappings found</p></div>';
        return;
    }

    container.innerHTML = mappings.map(mapping => `
        <div class="mapping-item">
            <div class="mapping-info">
                <div class="mapping-title">
                    ${mapping.slideClientName}
                    ${mapping.mapped ? '<span class="badge badge-success">✓ Mapped</span>' : '<span class="badge badge-warning">⚠ Unmapped</span>'}
                </div>
                ${mapping.mapped ? `<div class="mapping-subtitle">→ ${mapping.connectWiseName} (ID: ${mapping.connectWiseId})</div>` : ''}
            </div>
            <div class="mapping-actions">
                ${mapping.mapped ?
                    `<button class="btn btn-danger" onclick="deleteMapping('${mapping.slideClientId}')">🗑️ Delete</button>` :
                    `<button class="btn btn-primary" onclick="createMapping('${mapping.slideClientId}', '${escapeHtml(mapping.slideClientName)}')">➕ Map</button>`
                }
            </div>
        </div>
    `).join('');
}

function filterMappings(e) {
    renderMappings(e.target.value);
}

async function autoMapClients() {
    const btn = document.getElementById('autoMapBtn');
    btn.disabled = true;
    btn.textContent = '⏳ Mapping...';

    try {
        const response = await fetch('/api/mappings/auto', { method: 'POST' });
        if (response.ok) {
            showNotification('Auto-mapping completed!', 'success');
            loadMappings();
        } else {
            showNotification('Auto-mapping failed', 'error');
        }
    } catch (error) {
        showNotification('Error: ' + error.message, 'error');
    } finally {
        btn.disabled = false;
        btn.textContent = '🤖 Auto-Map Clients';
    }
}

function createMapping(slideClientId, slideClientName) {
    const modal = document.getElementById('mappingModal');
    document.getElementById('modalSlideClient').value = slideClientName;

    const cwSelect = document.getElementById('modalCWClient');
    cwSelect.innerHTML = '<option value="">Select a company...</option>' +
        state.cwClients.map(c => `<option value="${c.id}" data-name="${escapeHtml(c.name)}">${c.name}</option>`).join('');

    modal.classList.add('active');

    document.getElementById('saveMappingBtn').onclick = async () => {
        const cwId = parseInt(cwSelect.value);
        if (!cwId) {
            alert('Please select a ConnectWise company');
            return;
        }

        const cwName = cwSelect.options[cwSelect.selectedIndex].dataset.name;

        try {
            const response = await fetch('/api/mappings/create', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    slideClientId,
                    slideClientName,
                    connectWiseId: cwId,
                    connectWiseName: cwName
                })
            });

            if (response.ok) {
                showNotification('Mapping created successfully!', 'success');
                modal.classList.remove('active');
                loadMappings();
            } else {
                showNotification('Failed to create mapping', 'error');
            }
        } catch (error) {
            showNotification('Error: ' + error.message, 'error');
        }
    };
}

async function deleteMapping(slideClientId) {
    if (!confirm('Are you sure you want to delete this mapping?')) return;

    try {
        const response = await fetch('/api/mappings/delete', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ slideClientId })
        });

        if (response.ok) {
            showNotification('Mapping deleted successfully!', 'success');
            loadMappings();
        } else {
            showNotification('Failed to delete mapping', 'error');
        }
    } catch (error) {
        showNotification('Error: ' + error.message, 'error');
    }
}

// Ticketing Configuration
function initTicketing() {
    const form = document.getElementById('ticketingForm');
    const boardSelect = document.getElementById('boardSelect');
    const autoAssign = document.getElementById('autoAssignTech');

    boardSelect.addEventListener('change', async (e) => {
        const boardId = parseInt(e.target.value);
        if (boardId) {
            await loadStatusesAndTypes(boardId);
        }
    });

    autoAssign.addEventListener('change', (e) => {
        document.getElementById('technicianGroup').style.display = e.target.checked ? 'block' : 'none';
    });

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        await saveTicketingConfig();
    });

    document.getElementById('previewTemplateBtn').addEventListener('click', previewTemplate);
}

async function loadTicketingConfig() {
    try {
        const [configRes, boardsRes, prioritiesRes, membersRes] = await Promise.all([
            fetch('/api/ticketing/config'),
            fetch('/api/connectwise/boards'),
            fetch('/api/connectwise/priorities'),
            fetch('/api/connectwise/members')
        ]);

        state.config = await configRes.json();
        state.boards = await boardsRes.json();
        state.priorities = await prioritiesRes.json();
        state.members = await membersRes.json();

        // Populate boards
        const boardSelect = document.getElementById('boardSelect');
        boardSelect.innerHTML = '<option value="">Select a board...</option>' +
            state.boards.map(b => `<option value="${b.id}" data-name="${escapeHtml(b.name)}">${b.name}</option>`).join('');

        // Populate priorities
        const prioritySelect = document.getElementById('prioritySelect');
        prioritySelect.innerHTML = '<option value="">Select a priority...</option>' +
            state.priorities.map(p => `<option value="${p.id}" data-name="${escapeHtml(p.name)}">${p.name}</option>`).join('');

        // Populate members
        const memberSelect = document.getElementById('technicianSelect');
        memberSelect.innerHTML = '<option value="">Select a technician...</option>' +
            state.members.map(m => `<option value="${m.id}" data-name="${escapeHtml(m.firstName + ' ' + m.lastName)}">${m.firstName} ${m.lastName}</option>`).join('');

        // Load existing config
        if (state.config.board_id) {
            boardSelect.value = state.config.board_id;
            await loadStatusesAndTypes(state.config.board_id);
            document.getElementById('statusSelect').value = state.config.status_id;
            document.getElementById('typeSelect').value = state.config.type_id;
        }

        if (state.config.priority_id) {
            document.getElementById('prioritySelect').value = state.config.priority_id;
        }

        if (state.config.ticket_summary) {
            document.getElementById('ticketSummary').value = state.config.ticket_summary;
        }

        if (state.config.ticket_template) {
            document.getElementById('ticketTemplate').value = state.config.ticket_template;
        }

        if (state.config.auto_assign_tech) {
            document.getElementById('autoAssignTech').checked = true;
            document.getElementById('technicianGroup').style.display = 'block';
            if (state.config.technician_id) {
                document.getElementById('technicianSelect').value = state.config.technician_id;
            }
        }
    } catch (error) {
        console.error('Error loading ticketing config:', error);
    }
}

async function loadStatusesAndTypes(boardId) {
    try {
        const [statusesRes, typesRes] = await Promise.all([
            fetch(`/api/connectwise/statuses?boardId=${boardId}`),
            fetch(`/api/connectwise/types?boardId=${boardId}`)
        ]);

        state.statuses = await statusesRes.json();
        state.types = await typesRes.json();

        const statusSelect = document.getElementById('statusSelect');
        statusSelect.innerHTML = '<option value="">Select a status...</option>' +
            state.statuses.map(s => `<option value="${s.id}" data-name="${escapeHtml(s.name)}">${s.name}</option>`).join('');

        const typeSelect = document.getElementById('typeSelect');
        typeSelect.innerHTML = '<option value="">Select a type...</option>' +
            state.types.map(t => `<option value="${t.id}" data-name="${escapeHtml(t.name)}">${t.name}</option>`).join('');
    } catch (error) {
        console.error('Error loading statuses and types:', error);
    }
}

async function saveTicketingConfig() {
    const boardSelect = document.getElementById('boardSelect');
    const statusSelect = document.getElementById('statusSelect');
    const prioritySelect = document.getElementById('prioritySelect');
    const typeSelect = document.getElementById('typeSelect');
    const techSelect = document.getElementById('technicianSelect');

    const config = {
        board_id: parseInt(boardSelect.value),
        board_name: boardSelect.options[boardSelect.selectedIndex].dataset.name,
        status_id: parseInt(statusSelect.value),
        status_name: statusSelect.options[statusSelect.selectedIndex].dataset.name,
        priority_id: parseInt(prioritySelect.value),
        priority_name: prioritySelect.options[prioritySelect.selectedIndex].dataset.name,
        type_id: parseInt(typeSelect.value),
        type_name: typeSelect.options[typeSelect.selectedIndex].dataset.name,
        ticket_summary: document.getElementById('ticketSummary').value,
        ticket_template: document.getElementById('ticketTemplate').value,
        auto_assign_tech: document.getElementById('autoAssignTech').checked,
        technician_id: null,
        technician_name: ''
    };

    if (config.auto_assign_tech && techSelect.value) {
        config.technician_id = parseInt(techSelect.value);
        config.technician_name = techSelect.options[techSelect.selectedIndex].dataset.name;
    }

    try {
        const response = await fetch('/api/ticketing/config/save', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });

        if (response.ok) {
            showConfigStatus('Configuration saved successfully!', 'success');
        } else {
            showConfigStatus('Failed to save configuration', 'error');
        }
    } catch (error) {
        showConfigStatus('Error: ' + error.message, 'error');
    }
}

function previewTemplate() {
    const summary = document.getElementById('ticketSummary').value;
    const template = document.getElementById('ticketTemplate').value;

    const sampleData = {
        '{{alert_type}}': 'agent_backup_failed',
        '{{client_name}}': 'Acme Corporation',
        '{{device_name}}': 'SERVER-01',
        '{{alert_message}}': 'Backup failed: Disk space insufficient',
        '{{alert_timestamp}}': new Date().toLocaleString(),
        '{{agent_name}}': 'Veeam Agent',
        '{{agent_hostname}}': 'BACKUP-SERVER'
    };

    let previewSummary = summary;
    let previewTemplate = template;

    Object.entries(sampleData).forEach(([key, value]) => {
        previewSummary = previewSummary.replace(new RegExp(key.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g'), value);
        previewTemplate = previewTemplate.replace(new RegExp(key.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g'), value);
    });

    alert(`Preview:\n\nSummary:\n${previewSummary}\n\nDescription:\n${previewTemplate}`);
}

// Alerts
function initAlerts() {
    document.getElementById('refreshAlertsBtn').addEventListener('click', loadAlerts);
    document.getElementById('alertSearch').addEventListener('input', filterAlerts);
    document.getElementById('unresolvedOnly').addEventListener('change', filterAlerts);
}

async function loadAlerts() {
    const container = document.getElementById('alertsList');
    container.innerHTML = '<div class="loading">Loading alerts...</div>';

    try {
        const response = await fetch('/api/alerts');
        state.alerts = await response.json();
        renderAlerts();
    } catch (error) {
        container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">⚠️</div><p>Error loading alerts</p></div>';
        console.error('Error loading alerts:', error);
    }
}

function renderAlerts() {
    const container = document.getElementById('alertsList');
    const searchTerm = document.getElementById('alertSearch').value.toLowerCase();
    const unresolvedOnly = document.getElementById('unresolvedOnly').checked;

    let filtered = state.alerts.filter(alert => {
        if (unresolvedOnly && alert.resolved) return false;
        if (searchTerm && !JSON.stringify(alert).toLowerCase().includes(searchTerm)) return false;
        return true;
    });

    if (filtered.length === 0) {
        container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">✅</div><p>No alerts found</p></div>';
        return;
    }

    container.innerHTML = filtered.map(alert => `
        <div class="alert-item">
            <div class="alert-info">
                <div class="alert-title">
                    ${alert.type}
                    ${alert.resolved ? '<span class="badge badge-success">✓ Resolved</span>' : '<span class="badge badge-danger">⚠ Active</span>'}
                    ${alert.ticketId ? `<span class="badge badge-info">🎫 Ticket #${alert.ticketId}</span>` : ''}
                </div>
                <div class="alert-subtitle">
                    ${alert.clientName || alert.clientId} • ${alert.deviceId || 'Unknown Device'}
                </div>
                <div class="alert-subtitle">${alert.message || 'No message'}</div>
                <div class="timestamp">${new Date(alert.timestamp).toLocaleString()}</div>
            </div>
            <div class="alert-actions">
                ${!alert.resolved ? `<button class="btn btn-primary" onclick="closeAlert('${alert.id}')">✓ Close</button>` : ''}
            </div>
        </div>
    `).join('');
}

function filterAlerts() {
    renderAlerts();
}

async function closeAlert(alertId) {
    if (!confirm('Are you sure you want to close this alert?')) return;

    try {
        const response = await fetch('/api/alerts/close', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ alertId })
        });

        if (response.ok) {
            showNotification('Alert closed successfully!', 'success');
            loadAlerts();
        } else {
            showNotification('Failed to close alert', 'error');
        }
    } catch (error) {
        showNotification('Error: ' + error.message, 'error');
    }
}

// Tickets
function initTickets() {
    document.getElementById('refreshTicketsBtn').addEventListener('click', loadTicketMappings);
    document.getElementById('openTicketsOnly').addEventListener('change', renderTicketMappings);
}

async function loadTicketMappings() {
    const container = document.getElementById('ticketsList');
    container.innerHTML = '<div class="loading">Loading tickets...</div>';

    try {
        const response = await fetch('/api/tickets/mappings');
        state.ticketMappings = await response.json();
        renderTicketMappings();
    } catch (error) {
        container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">⚠️</div><p>Error loading tickets</p></div>';
        console.error('Error loading tickets:', error);
    }
}

function renderTicketMappings() {
    const container = document.getElementById('ticketsList');
    const openOnly = document.getElementById('openTicketsOnly').checked;

    let filtered = state.ticketMappings.filter(ticket => {
        if (openOnly && ticket.closedAt) return false;
        return true;
    });

    if (filtered.length === 0) {
        container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">📋</div><p>No tickets found</p></div>';
        return;
    }

    container.innerHTML = filtered.map(ticket => {
        // Determine status badge based on real-time CW status
        let statusBadge = '';
        if (ticket.ticketStatusError) {
            statusBadge = '<span class="badge badge-warning">⚠ Status Unknown</span>';
        } else if (ticket.ticketClosed) {
            statusBadge = `<span class="badge badge-success">✓ Closed (${ticket.ticketStatus})</span>`;
        } else {
            statusBadge = `<span class="badge badge-info">📂 ${ticket.ticketStatus}</span>`;
        }

        // Show sync warning if ticket is closed in CW but not in our DB
        const syncWarning = ticket.needsSync ? '<span class="badge badge-warning">⚠ Needs Sync</span>' : '';

        return `
            <div class="ticket-item">
                <div class="ticket-info">
                    <div class="alert-title">
                        Alert: ${ticket.alertId} → Ticket #${ticket.ticketId}
                        ${statusBadge}
                        ${syncWarning}
                    </div>
                    <div class="alert-subtitle">
                        ConnectWise Status: ${ticket.ticketStatus || 'Unknown'}
                        ${ticket.ticketClosedFlag ? ' (Closed Status Flag: Yes)' : ''}
                    </div>
                    <div class="timestamp">
                        Created: ${new Date(ticket.createdAt).toLocaleString()}
                        ${ticket.closedAt ? ` • Closed in DB: ${new Date(ticket.closedAt).toLocaleString()}` : ''}
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

// Modal handlers
function initModals() {
    const modal = document.getElementById('mappingModal');
    const closeBtn = modal.querySelector('.modal-close');
    const cancelBtn = document.getElementById('cancelMappingBtn');

    closeBtn.addEventListener('click', () => modal.classList.remove('active'));
    cancelBtn.addEventListener('click', () => modal.classList.remove('active'));

    window.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.classList.remove('active');
        }
    });
}

// Utility functions
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showNotification(message, type) {
    // Simple alert for now - could be enhanced with toast notifications
    alert(message);
}

function showConfigStatus(message, type) {
    const statusEl = document.getElementById('configStatus');
    statusEl.textContent = message;
    statusEl.className = `status-message ${type}`;

    setTimeout(() => {
        statusEl.className = 'status-message';
    }, 5000);
}