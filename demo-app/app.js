// Configuration
// Replace with your Railway backend URL
const API_URL = 'https://ticketingsystem-production-4a1d.up.railway.app/';

// State Management
const state = {
    token: localStorage.getItem('token'),
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    currentPage: 'events',
    events: [],
    tickets: [],
    organizerEvents: []
};

// Utility Functions
function showToast(message, type = 'success') {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.className = `toast ${type} show`;
    setTimeout(() => {
        toast.classList.remove('show');
    }, 3000);
}

function showModal(modalId) {
    document.getElementById(modalId).classList.add('show');
}

function hideModal(modalId) {
    document.getElementById(modalId).classList.remove('show');
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function formatCurrency(amount) {
    return `$${parseFloat(amount).toFixed(2)}`;
}

// API Functions
async function apiRequest(endpoint, options = {}) {
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json'
        }
    };

    if (state.token) {
        defaultOptions.headers['Authorization'] = `Bearer ${state.token}`;
    }

    const response = await fetch(`${API_URL}${endpoint}`, {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...options.headers
        }
    });

    if (!response.ok) {
        const error = await response.json().catch(() => ({ message: 'Request failed' }));
        throw new Error(error.message || `HTTP ${response.status}`);
    }

    return response.json();
}

// Authentication Functions
async function login(email, password) {
    try {
        const data = await apiRequest('/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });

        state.token = data.token;
        state.user = { email, id: data.user_id };
        localStorage.setItem('token', data.token);
        localStorage.setItem('user', JSON.stringify(state.user));

        showToast('Login successful!');
        updateUI();
        hideModal('loginModal');
        
        // Check if user is an organizer
        checkOrganizerStatus();
    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function signup(name, email, password, phone) {
    try {
        // Split name into first and last name
        const nameParts = name.trim().split(' ');
        const firstName = nameParts[0] || '';
        const lastName = nameParts.slice(1).join(' ') || nameParts[0]; // Use first name as last name if only one word
        
        // Generate username from email (part before @)
        const username = email.split('@')[0].toLowerCase();
        
        await apiRequest('/register', {
            method: 'POST',
            body: JSON.stringify({
                first_name: firstName,
                last_name: lastName,
                username: username,
                email,
                password,
                phone: phone && phone.trim() ? phone.trim() : null
            })
        });

        showToast('Account created! Please login.');
        hideModal('signupModal');
        showModal('loginModal');
    } catch (error) {
        showToast(error.message, 'error');
        console.error(error)
    }
}

function logout() {
    state.token = null;
    state.user = null;
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    showToast('Logged out successfully');
    updateUI();
    navigateToPage('events');
}

async function checkOrganizerStatus() {
    if (!state.token) return;

    try {
        const data = await apiRequest('/organizers/profile');
        if (data && data.is_verified) {
            document.getElementById('organizerLink').style.display = 'block';
        }
    } catch (error) {
        // User is not an organizer
        document.getElementById('organizerLink').style.display = 'none';
    }
}

// Events Functions
async function loadEvents(search = '') {
    const eventsList = document.getElementById('eventsList');
    eventsList.innerHTML = '<div class="loading">Loading events...</div>';

    try {
        let endpoint = '/events';
        if (search) {
            endpoint += `?search=${encodeURIComponent(search)}`;
        }

        const data = await apiRequest(endpoint);
        state.events = data.events || [];

        if (state.events.length === 0) {
            eventsList.innerHTML = `
                <div class="empty-state">
                    <h3>No events found</h3>
                    <p>Check back soon for upcoming events!</p>
                </div>
            `;
            return;
        }

        eventsList.innerHTML = state.events.map(event => {
            const eventDate = new Date(event.start_date || event.event_date);
            const now = new Date();
            const isPast = eventDate.getTime() < now.getTime();
            const statusBadge = isPast ? '<span class="badge badge-past">Past Event</span>' : '<span class="badge badge-upcoming">Upcoming</span>';
            
            // Get event image - use category-based colors as fallback
            const categoryColors = {
                'music': '#8b5cf6, #6366f1',
                'art': '#ec4899, #f43f5e',
                'sports': '#10b981, #059669',
                'tech': '#06b6d4, #0891b2',
                'food': '#f59e0b, #d97706',
                'default': '#6366f1, #8b5cf6'
            };
            const colors = isPast ? '#94a3b8, #64748b' : (categoryColors[event.category?.toLowerCase()] || categoryColors.default);
            
            let imageStyle = `background: linear-gradient(135deg, ${colors});`;
            if (event.images && event.images.length > 0) {
                const imagePath = event.images[0].image_path;
                // S3 URLs start with http, local paths need demo app server
                const imageUrl = imagePath.startsWith('https') ? imagePath : `http://172.22.29.124:3000/${imagePath}`;
                imageStyle = `background-image: url('${imageUrl}'); background-size: cover; background-position: center;`;
            }
            
            return `
            <div class="event-card ${isPast ? 'event-past' : ''}" onclick="showEventDetails(${event.id})">
                <div class="event-image" style="${imageStyle}">
                    ${statusBadge}
                </div>
                <div class="event-content">
                    <h3>${event.title || event.name || 'Event'}</h3>
                    <div class="event-date">📅 ${formatDate(event.start_date || event.event_date)}</div>
                    <div class="event-location">📍 ${event.location || 'TBA'}</div>
                    <div class="event-footer">
                        <span class="event-price">${event.currency ? event.currency.toUpperCase() : 'KSH'}</span>
                        <span class="event-capacity">${event.max_capacity || 0} capacity</span>
                    </div>
                </div>
            </div>
        `}).join('');
    } catch (error) {
        eventsList.innerHTML = `<div class="empty-state"><h3>Error loading events</h3><p>${error.message}</p></div>`;
    }
}

async function showEventDetails(eventId) {
    try {
        const event = await apiRequest(`/events/${eventId}`);
        const eventDate = new Date(event.start_date || event.event_date);
        const now = new Date();
        const isPast = eventDate.getTime() < now.getTime();
        
        // Store current event for purchase
        window.currentEvent = event;
        
        const eventDetails = `
            <div class="event-detail-header">
                <h2>${event.title || event.name || 'Event'}</h2>
                <div style="margin: 0.5rem 0;">
                    ${isPast ? '<span class="badge badge-past">Past Event</span>' : '<span class="badge badge-upcoming">Upcoming Event</span>'}
                    ${event.status ? `<span class="badge" style="background: #10b981; margin-left: 0.5rem;">${event.status}</span>` : ''}
                </div>
                <p>${event.description || 'No description available'}</p>
            </div>
            <div class="event-detail-info">
                <div class="info-row">
                    <strong>📅 Start Date:</strong> ${formatDate(event.start_date || event.event_date)}
                </div>
                ${event.end_date ? `<div class="info-row"><strong>📅 End Date:</strong> ${formatDate(event.end_date)}</div>` : ''}
                <div class="info-row">
                    <strong>📍 Location:</strong> ${event.location || 'TBA'}
                </div>
                ${event.location_address ? `<div class="info-row"><strong>📍 Address:</strong> ${event.location_address}</div>` : ''}
                ${event.location_country ? `<div class="info-row"><strong>🌍 Country:</strong> ${event.location_country}</div>` : ''}
                <div class="info-row">
                    <strong>🎫 Capacity:</strong> ${event.max_capacity || 0}
                </div>
                <div class="info-row">
                    <strong>💰 Currency:</strong> ${event.currency ? event.currency.toUpperCase() : 'KSH'}
                </div>
                ${event.min_age ? `<div class="info-row"><strong>🔞 Min Age:</strong> ${event.min_age}</div>` : ''}
                ${event.category ? `<div class="info-row"><strong>🎨 Category:</strong> ${event.category}</div>` : ''}
            </div>
            ${event.organizer ? `
                <div class="organizer-info" style="background: var(--bg-color); padding: 1rem; border-radius: 6px; margin-top: 1rem;">
                    <h4>Organizer</h4>
                    <p><strong>${event.organizer.name || 'Organizer'}</strong></p>
                    ${event.organizer.about ? `<p style="font-size: 0.9rem; color: var(--text-secondary);">${event.organizer.about}</p>` : ''}
                </div>
            ` : ''}
            ${event.pre_order_message_display ? `
                <div class="event-message" style="background: #eff6ff; padding: 1rem; border-radius: 6px; margin-top: 1rem; border-left: 4px solid var(--primary-color);">
                    <p style="margin: 0;">${event.pre_order_message_display}</p>
                </div>
            ` : ''}
            ${state.token && !isPast ? `
                <div class="purchase-section" style="background: var(--bg-color); padding: 1.5rem; border-radius: 8px; margin-top: 1rem;">
                    <h3 style="margin-bottom: 1rem;">🎫 Purchase Demo Tickets</h3>
                    <p style="color: var(--text-secondary); font-size: 0.9rem; margin-bottom: 1rem;">This is a demo. In production, this would integrate with ticket classes and payment gateways.</p>
                    <div class="quantity-selector" style="margin: 1rem 0;">
                        <label style="display: block; margin-bottom: 0.5rem;"><strong>Quantity:</strong></label>
                        <div style="display: flex; align-items: center; gap: 1rem;">
                            <button class="btn btn-secondary" onclick="changeTicketQuantity(-1)">-</button>
                            <input type="number" id="demoTicketQuantity" value="1" min="1" max="10" style="width: 80px; text-align: center; padding: 0.5rem; border: 1px solid var(--border-color); border-radius: 6px;">
                            <button class="btn btn-secondary" onclick="changeTicketQuantity(1)">+</button>
                        </div>
                    </div>
                    <button class="btn btn-primary btn-block" onclick="simulateTicketPurchase(${event.id})" style="margin-top: 1rem;">
                        🛒 Simulate Purchase (Demo)
                    </button>
                    ${event.enable_offline_payment ? '<p style="color: var(--success-color); font-size: 0.9rem; margin-top: 0.5rem;">✓ Offline payment available</p>' : ''}
                </div>
            ` : !state.token ? `
                <div class="purchase-section">
                    <p>Please <a href="#" onclick="hideModal('eventModal'); showModal('loginModal'); return false;" style="color: var(--primary-color); text-decoration: none; font-weight: 600;">login</a> to explore ticket purchasing.</p>
                </div>
            ` : ''}
            ${event.tags ? `<p style="margin-top: 1rem; color: var(--text-secondary);"><strong>Tags:</strong> ${event.tags}</p>` : ''}
        `;

        document.getElementById('eventDetails').innerHTML = eventDetails;
        showModal('eventModal');
    } catch (error) {
        showToast(error.message, 'error');
    }
}

function changeTicketQuantity(delta) {
    const input = document.getElementById('demoTicketQuantity');
    if (!input) return;
    
    const newValue = parseInt(input.value) + delta;
    const max = parseInt(input.max);
    const min = parseInt(input.min);
    
    if (newValue >= min && newValue <= max) {
        input.value = newValue;
    }
}

function simulateTicketPurchase(eventId) {
    const quantity = document.getElementById('demoTicketQuantity')?.value || 1;
    const event = window.currentEvent;
    
    showToast(`Demo: Successfully "purchased" ${quantity} ticket(s) for ${event.title}! In production, this would process payment and generate actual tickets.`, 'success');
    hideModal('eventModal');
}

function changeQuantity(delta) {
    const input = document.getElementById('ticketQuantity');
    const newValue = parseInt(input.value) + delta;
    const max = parseInt(input.max);
    
    if (newValue >= 1 && newValue <= max) {
        input.value = newValue;
        updateTotalPrice();
    }
}

function updateTotalPrice() {
    const quantity = document.getElementById('ticketQuantity').value;
    const eventCard = state.events.find(e => e.id === window.currentEventId);
    if (eventCard) {
        const total = quantity * parseFloat(eventCard.ticket_price);
        document.getElementById('totalPrice').textContent = formatCurrency(total);
    }
}

async function purchaseTickets(eventId, ticketPrice) {
    const quantity = parseInt(document.getElementById('ticketQuantity').value);
    
    try {
        // Create order
        const orderData = await apiRequest('/orders', {
            method: 'POST',
            body: JSON.stringify({
                event_id: eventId,
                quantity: quantity
            })
        });

        // Process payment (simplified - in production, integrate with actual payment gateway)
        await apiRequest(`/orders/${orderData.order_id}/payment`, {
            method: 'POST',
            body: JSON.stringify({
                payment_method: 'mpesa', // or 'card'
                phone_number: '+254712345678' // Would be from user input
            })
        });

        showToast('Tickets purchased successfully!', 'success');
        hideModal('eventModal');
        
        // Refresh tickets if on tickets page
        if (state.currentPage === 'tickets') {
            loadTickets();
        }
    } catch (error) {
        showToast(error.message, 'error');
    }
}

// Tickets Functions
async function loadTickets() {
    const ticketsList = document.getElementById('ticketsList');
    ticketsList.innerHTML = '<div class="loading">Loading tickets...</div>';

    try {
        const data = await apiRequest('/tickets');
        state.tickets = data.tickets || [];

        if (state.tickets.length === 0) {
            ticketsList.innerHTML = `
                <div class="empty-state">
                    <h3>No tickets yet</h3>
                    <p>Purchase tickets to see them here!</p>
                </div>
            `;
            return;
        }

        ticketsList.innerHTML = state.tickets.map(ticket => `
            <div class="ticket-card">
                <div class="ticket-info">
                    <h3>${ticket.event_name || 'Event'}</h3>
                    <p>${formatDate(ticket.event_date)}</p>
                    <span class="ticket-number">${ticket.ticket_number}</span>
                </div>
                <div>
                    <span class="ticket-status ${ticket.is_used ? 'used' : 'valid'}">
                        ${ticket.is_used ? 'Used' : 'Valid'}
                    </span>
                    <button class="btn btn-primary" style="margin-top: 0.5rem;" onclick="downloadTicket('${ticket.id}')">
                        Download PDF
                    </button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        ticketsList.innerHTML = `<div class="empty-state"><h3>Error loading tickets</h3><p>${error.message}</p></div>`;
    }
}

async function downloadTicket(ticketId) {
    try {
        const response = await fetch(`${API_URL}/tickets/${ticketId}/pdf`, {
            headers: {
                'Authorization': `Bearer ${state.token}`
            }
        });

        if (!response.ok) throw new Error('Failed to download ticket');

        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `ticket-${ticketId}.pdf`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
        
        showToast('Ticket downloaded!');
    } catch (error) {
        showToast(error.message, 'error');
    }
}

// Organizer Functions
async function loadOrganizerDashboard() {
    try {
        // Load stats
        const stats = await apiRequest('/organizers/dashboard/stats');
        document.getElementById('organizerStats').innerHTML = `
            <div class="stat-card">
                <h4>Total Events</h4>
                <div class="value">${stats.total_events || 0}</div>
            </div>
            <div class="stat-card">
                <h4>Total Revenue</h4>
                <div class="value">${formatCurrency(stats.total_revenue || 0)}</div>
            </div>
            <div class="stat-card">
                <h4>Tickets Sold</h4>
                <div class="value">${stats.total_tickets_sold || 0}</div>
            </div>
            <div class="stat-card">
                <h4>Active Events</h4>
                <div class="value">${stats.active_events || 0}</div>
            </div>
        `;

        // Load organizer events
        const eventsData = await apiRequest('/organizers/events');
        state.organizerEvents = eventsData.events || [];

        const organizerEvents = document.getElementById('organizerEvents');
        if (state.organizerEvents.length === 0) {
            organizerEvents.innerHTML = `
                <div class="empty-state">
                    <h3>No events yet</h3>
                    <p>Create your first event to get started!</p>
                </div>
            `;
            return;
        }

        organizerEvents.innerHTML = state.organizerEvents.map(event => {
            // Get event image
            const categoryColors = {
                'music': '#8b5cf6, #6366f1',
                'art': '#ec4899, #f43f5e',
                'sports': '#10b981, #059669',
                'tech': '#06b6d4, #0891b2',
                'food': '#f59e0b, #d97706',
                'default': '#6366f1, #8b5cf6'
            };
            const colors = categoryColors[event.category?.toLowerCase()] || categoryColors.default;
            
            let imageStyle = `background: linear-gradient(135deg, ${colors});`;
            if (event.images && event.images.length > 0) {
                const imagePath = event.images[0].image_path;
                const imageUrl = imagePath.startsWith('http') ? imagePath : `http://172.22.29.124:3000/${imagePath}`;
                imageStyle = `background-image: url('${imageUrl}'); background-size: cover; background-position: center;`;
            }
            
            return `
            <div class="event-card">
                <div class="event-image" style="${imageStyle}"></div>
                <div class="event-content">
                    <h3>${event.name}</h3>
                    <div class="event-date">📅 ${formatDate(event.event_date)}</div>
                    <div class="event-location">📍 ${event.location || 'TBA'}</div>
                    <div class="event-footer">
                        <span class="event-price">${formatCurrency(event.ticket_price)}</span>
                        <span class="event-capacity">${event.sold_tickets || 0} / ${event.total_capacity || 0}</span>
                    </div>
                    <button class="btn btn-primary btn-block" style="margin-top: 1rem;" onclick="editEvent(${event.id})">
                        Edit Event
                    </button>
                </div>
            </div>
        `;
        }).join('');
    } catch (error) {
        document.getElementById('organizerStats').innerHTML = `<div class="empty-state"><h3>Error loading dashboard</h3><p>${error.message}</p></div>`;
    }
}

async function createEvent(formData) {
    try {
        await apiRequest('/organizers/events', {
            method: 'POST',
            body: JSON.stringify(formData)
        });

        showToast('Event created successfully!');
        hideModal('createEventModal');
        loadOrganizerDashboard();
    } catch (error) {
        showToast(error.message, 'error');
    }
}

function editEvent(eventId) {
    showToast('Edit functionality coming soon!', 'warning');
}

// Navigation Functions
function navigateToPage(pageName) {
    // Update active page
    document.querySelectorAll('.page').forEach(page => page.classList.remove('active'));
    document.getElementById(`${pageName}Page`).classList.add('active');

    // Update active nav link
    document.querySelectorAll('.nav-link').forEach(link => link.classList.remove('active'));
    document.querySelector(`[data-page="${pageName}"]`)?.classList.add('active');

    state.currentPage = pageName;

    // Load page data
    switch (pageName) {
        case 'events':
            loadEvents();
            break;
        case 'tickets':
            loadTickets();
            break;
        case 'organizer':
            loadOrganizerDashboard();
            break;
    }
}

function updateUI() {
    if (state.token && state.user) {
        // Show authenticated UI
        document.getElementById('loginBtn').style.display = 'none';
        document.getElementById('signupBtn').style.display = 'none';
        document.getElementById('userMenu').style.display = 'flex';
        document.getElementById('myTicketsLink').style.display = 'block';
        document.getElementById('userEmail').textContent = state.user.email;
        
        // Check organizer status
        checkOrganizerStatus();
    } else {
        // Show guest UI
        document.getElementById('loginBtn').style.display = 'block';
        document.getElementById('signupBtn').style.display = 'block';
        document.getElementById('userMenu').style.display = 'none';
        document.getElementById('myTicketsLink').style.display = 'none';
        document.getElementById('organizerLink').style.display = 'none';
    }
}

// Event Listeners
document.addEventListener('DOMContentLoaded', () => {
    // Initialize UI
    updateUI();
    loadEvents();

    // Navigation
    document.querySelectorAll('.nav-link').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const page = e.target.dataset.page;
            if (page) navigateToPage(page);
        });
    });

    // Auth buttons
    document.getElementById('loginBtn').addEventListener('click', () => showModal('loginModal'));
    document.getElementById('signupBtn').addEventListener('click', () => showModal('signupModal'));
    document.getElementById('logoutBtn').addEventListener('click', logout);

    // Modal switches
    document.getElementById('switchToSignup').addEventListener('click', (e) => {
        e.preventDefault();
        hideModal('loginModal');
        showModal('signupModal');
    });

    document.getElementById('switchToLogin').addEventListener('click', (e) => {
        e.preventDefault();
        hideModal('signupModal');
        showModal('loginModal');
    });

    // Close modals
    document.querySelectorAll('.close').forEach(closeBtn => {
        closeBtn.addEventListener('click', (e) => {
            e.target.closest('.modal').classList.remove('show');
        });
    });

    // Close modal on outside click
    window.addEventListener('click', (e) => {
        if (e.target.classList.contains('modal')) {
            e.target.classList.remove('show');
        }
    });

    // Login form
    document.getElementById('loginForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('loginEmail').value;
        const password = document.getElementById('loginPassword').value;
        await login(email, password);
    });

    // Signup form
    document.getElementById('signupForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const name = document.getElementById('signupName').value;
        const email = document.getElementById('signupEmail').value;
        const password = document.getElementById('signupPassword').value;
        const phone = document.getElementById('signupPhone').value;
        await signup(name, email, password, phone);
    });

    // Search
    document.getElementById('searchBtn').addEventListener('click', () => {
        const search = document.getElementById('searchInput').value;
        loadEvents(search);
    });

    document.getElementById('searchInput').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            const search = e.target.value;
            loadEvents(search);
        }
    });

    // Create event
    document.getElementById('createEventBtn').addEventListener('click', () => showModal('createEventModal'));

    document.getElementById('createEventForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = {
            name: document.getElementById('eventName').value,
            description: document.getElementById('eventDescription').value,
            event_date: new Date(document.getElementById('eventDate').value).toISOString(),
            total_capacity: parseInt(document.getElementById('eventCapacity').value),
            ticket_price: parseFloat(document.getElementById('eventPrice').value),
            location: document.getElementById('eventLocation').value
        };
        await createEvent(formData);
        e.target.reset();
    });
});
