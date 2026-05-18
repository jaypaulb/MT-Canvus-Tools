// Color utility function
function isColorDark(hexColor) {
    const r = parseInt(hexColor.slice(1, 3), 16);
    const g = parseInt(hexColor.slice(3, 5), 16);
    const b = parseInt(hexColor.slice(5, 7), 16);
    return (r * 0.299 + g * 0.587 + b * 0.114) / 255 < 0.5;
}

// Admin functions
// Function to display user list
async function displayUserList(users) {
    console.log('displayUserList called with:', users);
    const tbody = document.querySelector('#userListTable tbody');
    const message = document.getElementById('userListMessage');
    
    if (!tbody || !message) {
        console.error('Table elements not found:', { tbody: !!tbody, message: !!message });
        return;
    }
    
    if (!users || users.length === 0) {
        console.log('No users to display');
        tbody.innerHTML = '';
        message.style.display = 'block';
        return;
    }

    console.log('Rendering users table');
    message.style.display = 'none';
    const html = users.map(user => `
        <tr>
            <td>Team ${user.team}</td>
            <td>${user.name}</td>
            <td>${user.color.toUpperCase()}</td>
        </tr>
    `).join('');
    console.log('Generated HTML:', html);
    tbody.innerHTML = html;
}

async function createTargets() {
    console.log('createTargets called');
    try {
        const result = await makeAdminRequest('/admin/createTargets', 'POST');
        if (result && result.success) {
            MessageSystem.display(result.message || "Targets created successfully", "success");
        }
    } catch (error) {
        console.error('Create targets failed:', error);
        MessageSystem.display(error.message || "Failed to create targets", "error");
    }
}

async function deleteTargets() {
    console.log('deleteTargets called');
    try {
        const result = await makeAdminRequest('/admin/deleteTargets', 'POST');
        if (result && result.success) {
            MessageSystem.display(result.message || "Targets deleted successfully", "success");
        }
    } catch (error) {
        console.error('Delete targets failed:', error);
        MessageSystem.display(error.message || "Failed to delete targets", "error");
    }
}

async function listUsers() {
    console.log('listUsers called');
    try {
        const result = await makeAdminRequest('/admin/listUsers', 'GET');
        console.log('Server response:', result);
        if (result && result.success) {
            console.log('Users data received:', result.users);
            await displayUserList(result.users);
            MessageSystem.display("Users listed successfully", "success");
        }
    } catch (error) {
        console.error('List users failed:', error);
        MessageSystem.display(error.message || "Failed to list users", "error");
    }
}

async function deleteUsers() {
    console.log('deleteUsers called');
    try {
        const result = await makeAdminRequest('/admin/deleteUsers', 'POST', { all: true });
        if (result && result.success) {
            MessageSystem.display(result.message || "Users deleted successfully", "success");
            // Clear the user list
            displayUserList([]);
        }
    } catch (error) {
        console.error('Delete users failed:', error);
        MessageSystem.display(error.message || "Failed to delete users", "error");
    }
}

document.addEventListener('DOMContentLoaded', () => {
    // Initialize common components
    MessageSystem.initialize();
    AdminModal.initialize();

    // Initialize empty user list with proper error handling
    try {
        displayUserList([]);
    } catch (error) {
        console.warn('Could not initialize user list:', error);
    }

    // User Identification Elements
    const teamButtons = document.querySelectorAll('.team-button');
    const userNameInput = document.getElementById('username');
    const userIdentification = document.getElementById('user-identification');
    const identificationMessage = document.getElementById('message');
    const userDashboard = document.getElementById('user-dashboard');
    const welcomeMessage = document.getElementById('welcomeMessage');
    const noteSquare = document.getElementById('noteSquare');
    const uploadItemButton = document.getElementById('uploadItemButton');
    const postNoteButton = document.getElementById('postNoteButton');
    const backButton = document.getElementById('backButton');

    // Back button functionality
    backButton.addEventListener('click', () => {
        userDashboard.style.display = "none";
        userIdentification.style.display = "block";
        // Clear the note square
        noteSquare.textContent = "";
        // Reset team selection
        teamButtons.forEach(btn => btn.classList.remove('active'));
        // Clear user input
        userNameInput.value = "";
        // Reset variables
        selectedTeam = null;
        userName = null;
        userColor = null;
    });

    // Admin Elements
    const createTargetsBtn = document.getElementById('createTargetsBtn');
    const deleteTargetsBtn = document.getElementById('deleteTargetsBtn');
    const listUsersBtn = document.getElementById('listUsersBtn');
    const deleteUsersBtn = document.getElementById('deleteUsersBtn');

    let selectedTeam = null;
    let userName = null;
    let userColor = null;

    // Team colors (ROYGBIV)
    const teamColors = {
        1: "#FF0000FF", // Red
        2: "#FF7F00FF", // Orange
        3: "#FFFF00FF", // Yellow
        4: "#00FF00FF", // Green
        5: "#0000FFFF", // Blue
        6: "#4B0082FF", // Indigo
        7: "#8B00FFFF"  // Violet
    };

    // Function to update identification message
    function updateIdentificationMessage(text, type) {
        if (identificationMessage) {
            identificationMessage.textContent = text;
            identificationMessage.className = `message ${type}`;
            identificationMessage.style.display = 'block';
        }
    }

    // Team selection and immediate identification
    teamButtons.forEach(button => {
        const teamNumber = button.dataset.team;
        button.style.backgroundColor = teamColors[teamNumber];
        button.style.color = isColorDark(teamColors[teamNumber]) ? '#FFFFFF' : '#000000';
        
        button.addEventListener('click', async () => {
            const name = userNameInput.value.trim();
            if (!name) {
                updateIdentificationMessage("Please enter your name first.", "error");
                    return;
                }

            teamButtons.forEach(btn => btn.classList.remove('active'));
            button.classList.add('active');
            selectedTeam = parseInt(teamNumber);
            userName = name;

            try {
                // Get user color from server
                const response = await fetch('/identify-user', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ team: selectedTeam, name: userName })
                });

                const data = await response.json();
                if (response.ok && data.success) {
                    userColor = data.color;

                    // Get canvas name
                    const envResponse = await fetch('/get-canvas-info');
                    const envData = await envResponse.json();
                    const canvas_name = envData.canvas_name || "Unknown Canvas";

                    // Update visibility of sections
                    if (userIdentification) userIdentification.style.display = "none";
                    if (userDashboard) {
                        userDashboard.style.display = "block";
                        // Update welcome message and note square
                        welcomeMessage.textContent = `Welcome, ${userName}, you are currently posting to Team ${selectedTeam} on ${canvas_name}.`;
                        noteSquare.style.backgroundColor = userColor;
                        noteSquare.style.color = isColorDark(userColor) ? '#FFFFFF' : '#000000';
                    }
                    
                    updateIdentificationMessage("Identification successful!", "success");
                    MessageSystem.display("Identification successful!", "success");
                } else {
                    updateIdentificationMessage(data.error || "Identification failed.", "error");
                    MessageSystem.display(data.error || "Identification failed.", "error");
                }
            } catch (error) {
                console.error("Error identifying user:", error);
                updateIdentificationMessage("An error occurred during identification.", "error");
                MessageSystem.display("An error occurred during identification.", "error");
            }
        });
    });

    // Handle post note button
    postNoteButton.addEventListener('click', async () => {
        const noteText = noteSquare.textContent.trim();
        if (!noteText) {
            MessageSystem.display("Please enter text in the note.", "error");
            return;
        }

        try {
            const response = await fetch('/create-note', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    team: selectedTeam,
                    name: userName,
                    text: noteText,
                    color: userColor
                })
            });

            const data = await response.json();
            if (response.ok && data.success) {
                MessageSystem.display("Note posted successfully!", "success");
                noteSquare.textContent = ""; // Clear the note
            } else {
                MessageSystem.display(data.error || "Failed to post note.", "error");
            }
        } catch (error) {
            console.error("Error posting note:", error);
            MessageSystem.display("An error occurred while posting the note.", "error");
        }
    });

    // Handle file upload button
    uploadItemButton.addEventListener('click', () => {
        const fileInput = document.createElement('input');
        fileInput.type = 'file';
        fileInput.accept = '.jpg,.jpeg,.png,.gif,.bmp,.tiff,.mp4,.avi,.mov,.wmv,.pdf,.mkv';

        fileInput.onchange = async () => {
            const file = fileInput.files[0];
            if (file) {
                const formData = new FormData();
                formData.append('team', selectedTeam);
                formData.append('name', userName);
                formData.append('file', file);

                try {
                    MessageSystem.display("Uploading file...", "loading");
                    const response = await fetch('/upload-item', {
                        method: 'POST',
                        body: formData
                    });

                    const data = await response.json();
                    if (response.ok && data.success) {
                        MessageSystem.display(data.message, "success");
                } else {
                        MessageSystem.display(data.error || "Upload failed.", "error");
                    }
                } catch (error) {
                    console.error("Error uploading file:", error);
                    MessageSystem.display("An error occurred during upload.", "error");
                }
            }
        };

        fileInput.click();
    });

    // Event listeners for admin buttons
    if (createTargetsBtn) {
        console.log('Found createTargets button, adding listener');
        createTargetsBtn.addEventListener('click', createTargets);
    } else {
        console.warn('createTargetsBtn button not found');
    }
    
    if (deleteTargetsBtn) {
        console.log('Found deleteTargets button, adding listener');
        deleteTargetsBtn.addEventListener('click', deleteTargets);
            } else {
        console.warn('deleteTargetsBtn button not found');
    }
    
    if (listUsersBtn) {
        console.log('Found listUsers button, adding listener');
        listUsersBtn.addEventListener('click', listUsers);
                } else {
        console.warn('listUsersBtn button not found');
    }
    
    if (deleteUsersBtn) {
        console.log('Found deleteUsers button, adding listener');
        deleteUsersBtn.addEventListener('click', deleteUsers);
            } else {
        console.warn('deleteUsersBtn button not found');
    }

    // Update welcome message to include team number
    function updateWelcomeMessage(name, teamNumber) {
        if (welcomeMessage) {
            welcomeMessage.textContent = `Welcome ${name}! You are on Team ${teamNumber}`;
        }
    }
});