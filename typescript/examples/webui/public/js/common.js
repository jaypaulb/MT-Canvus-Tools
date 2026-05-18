// /public/js/main.js

// UI Component: Message System
const MessageSystem = {
    container: null,
    
    initialize() {
        if (!this.container) {
            this.container = document.createElement('div');
            this.container.id = 'messageContainer';
            this.container.style.position = 'fixed';
            this.container.style.top = '20px';
            this.container.style.right = '20px';
            this.container.style.zIndex = '1000';
            document.body.appendChild(this.container);
        }
    },
    
    display(text, type = 'info') {
        this.initialize();
        
        // Remove existing messages of the same type
        const existingMessages = this.container.querySelectorAll(`.message.${type}`);
        existingMessages.forEach(msg => msg.remove());
        
        const msgDiv = document.createElement('div');
        msgDiv.className = `message ${type}`;
        msgDiv.textContent = text;
        msgDiv.style.padding = '10px';
        msgDiv.style.marginBottom = '10px';
        msgDiv.style.borderRadius = '4px';
        msgDiv.style.backgroundColor = type === 'success' ? '#4CAF50' : 
                                     type === 'error' ? '#f44336' : 
                                     type === 'info' ? '#2196F3' : '#ff9800';
        msgDiv.style.color = 'white';
        
        this.container.appendChild(msgDiv);
        
        setTimeout(() => msgDiv.remove(), 5000);
    }
};

// UI Component: Admin Modal
const AdminModal = {
    modal: null,
    passwordInput: null,
    confirmBtn: null,
    cancelBtn: null,
    cleanup: null,
    
    initialize() {
        if (this.modal) return; // Already initialized
        
        this.modal = document.getElementById('adminModal');
        this.passwordInput = document.getElementById('adminPassword');
        this.confirmBtn = document.getElementById('adminConfirm');
        this.cancelBtn = document.getElementById('adminCancel');
        
        if (!this.modal || !this.passwordInput || !this.confirmBtn || !this.cancelBtn) {
            console.error('Required modal elements not found:', {
                modal: !!this.modal,
                passwordInput: !!this.passwordInput,
                confirmBtn: !!this.confirmBtn,
                cancelBtn: !!this.cancelBtn
            });
            throw new Error('Required modal elements not found');
        }

        // Add click outside handler
        this.modal.addEventListener('click', (e) => {
            if (e.target === this.modal) {
                this.hide();
            }
        });
    },
    
    show() {
        return new Promise((resolve, reject) => {
            try {
                this.initialize();
                
                // Setup handlers
                const handleConfirm = async () => {
                    const password = this.passwordInput.value;
                    if (!password) {
                        MessageSystem.display("Please enter the admin password", "error");
                        return;
                    }
                    
                    try {
                        const response = await fetch('/validateAdmin', {
                            method: 'POST',
                            headers: {
                                'Authorization': `Bearer ${password}`,
                                'Content-Type': 'application/json'
                            }
                        });
                        
                        const data = await response.json();
                        if (!response.ok || !data.success) {
                            throw new Error(data.error || 'Authentication failed');
                        }
                        
                        this.hide();
                        resolve(password);
                    } catch (error) {
                        console.error('Admin authentication failed:', error);
                        MessageSystem.display(error.message || "Authentication failed", "error");
                    }
                };
                
                const handleCancel = () => {
                    this.hide();
                    resolve(null);
                };
                
                const handleKeydown = (event) => {
                    if (event.key === 'Enter' && document.activeElement === this.passwordInput) {
                        event.preventDefault();
                        handleConfirm();
                    } else if (event.key === 'Escape') {
                        event.preventDefault();
                        handleCancel();
                    }
                };
                
                // Clean up old listeners if they exist
                if (this.cleanup) {
                    this.cleanup();
                }
                
                // Set up new listeners
                this.confirmBtn.addEventListener('click', handleConfirm);
                this.cancelBtn.addEventListener('click', handleCancel);
                document.addEventListener('keydown', handleKeydown);
                
                // Store cleanup function
                this.cleanup = () => {
                    this.confirmBtn.removeEventListener('click', handleConfirm);
                    this.cancelBtn.removeEventListener('click', handleCancel);
                    document.removeEventListener('keydown', handleKeydown);
                };
                
                // Reset and show modal
                this.passwordInput.value = '';
                this.modal.classList.add('show');
                setTimeout(() => this.passwordInput.focus(), 100); // Focus after animation starts
                
            } catch (error) {
                console.error('Modal initialization failed:', error);
                reject(error);
            }
        });
    },
    
    hide() {
        if (this.modal) {
            this.modal.classList.remove('show');
            if (this.cleanup) {
                this.cleanup();
                this.cleanup = null;
            }
            this.passwordInput.value = '';
        }
    }
};

// Admin Authentication Service
let adminToken = null;

async function makeAdminRequest(url, method = 'POST', body = null) {
    console.log('Making admin request to', url);
    
    try {
        if (!adminToken) {
            // Show modal and get authentication
            const token = await AdminModal.show();
            if (!token) {
                throw new Error('Authentication cancelled');
            }
            adminToken = token;
        }

        // Make the actual request with the token
        const response = await fetch(url, {
            method: method,
            headers: {
                'Authorization': `Bearer ${adminToken}`,
                'Content-Type': 'application/json'
            },
            ...(body ? { body: JSON.stringify(body) } : {})
        });

        const data = await response.json();
        
        if (!response.ok) {
            // If auth failed, clear token and retry once
            if (response.status === 401) {
                adminToken = null;
                return makeAdminRequest(url, method, body);
            }
            throw new Error(data.error || 'Request failed');
        }

        return data;
    } catch (error) {
        console.error('Admin request failed:', error);
        MessageSystem.display(error.message || "Request failed", "error");
        throw error;
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    MessageSystem.initialize();
});
  