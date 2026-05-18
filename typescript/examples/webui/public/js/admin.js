// /public/js/admin.js

document.addEventListener('DOMContentLoaded', async () => {
    const envFormContainer = document.getElementById('env-form-container');
    const envForm = document.getElementById('env-form');
    const envMessage = document.getElementById('env-message');
    const confirmModal = document.getElementById('confirm-modal');
    const confirmSaveButton = document.getElementById('confirm-save');
    const cancelSaveButton = document.getElementById('cancel-save');
    const successNotification = document.getElementById('successNotification');
    const themeSection = document.getElementById('theme-section');
    const themeSelect = document.getElementById('theme-select');
    const applyThemeBtn = document.getElementById('apply-theme');
    const editThemeBtn = document.getElementById('edit-theme');
    const previewFrame = document.getElementById('preview-frame');
    const colorPickerPanel = document.getElementById('color-picker-panel');
    const selectedElementText = document.getElementById('selected-element');
    const elementColor = document.getElementById('element-color');
    const elementBackground = document.getElementById('element-background');
    const applyColorBtn = document.getElementById('apply-color');
    const closePickerBtn = document.getElementById('close-picker');
    const saveThemeBtn = document.getElementById('save-theme');
    const cancelThemeBtn = document.getElementById('cancel-theme');
    const editorPreview = document.getElementById('editor-preview');
    const saveThemeChangesBtn = document.getElementById('save-theme-changes');
    const createThemeBtn = document.getElementById('create-theme');
    const createThemeModal = document.getElementById('create-theme-modal');
    const confirmCreateThemeBtn = document.getElementById('confirm-create-theme');
    const cancelCreateThemeBtn = document.getElementById('cancel-create-theme');
    const newThemeNameInput = document.getElementById('new-theme-name');
  
    let currentTheme = localStorage.getItem('theme') || 'default';
    let selectedElement = null;
    let isEditing = false;
    let editingTheme = null;
  
    // Function to display messages
    function displayMessage(element, text, type) {
      element.textContent = text;
      element.className = `message ${type}`;
      element.style.display = "block";
    }
  
    // Function to clear messages
    function clearMessage(element) {
      element.textContent = "";
      element.className = "message";
      element.style.display = "none";
    }
  
    // Function to show success notification
    function showSuccessNotification(message = 'Changes saved successfully!') {
        successNotification.textContent = message;
        successNotification.classList.add('show');
        setTimeout(() => {
            successNotification.classList.remove('show');
        }, 3000);
    }

    // Function to apply theme to preview only
    function applyThemeToPreview(theme) {
        const isDark = localStorage.getItem('theme')?.includes('dark') || false;
        const actualTheme = theme === 'kpmg-light' ? 
            (isDark ? 'kpmg-dark' : 'kpmg-light') : 
            (isDark ? 'dark' : 'default');
        
        previewFrame.setAttribute('data-theme', actualTheme);
        
        // Clone current theme variables to preview frame
        const computedStyle = getComputedStyle(document.documentElement);
        const themeVars = {};
        for (const prop of computedStyle) {
            if (prop.startsWith('--')) {
                themeVars[prop] = computedStyle.getPropertyValue(prop);
            }
        }
        // Apply theme variables to preview frame
        Object.entries(themeVars).forEach(([prop, value]) => {
            previewFrame.style.setProperty(prop, value);
        });
    }

    // Function to apply theme globally
    function applyThemeGlobally(theme) {
        const isDark = localStorage.getItem('theme')?.includes('dark') || false;
        const actualTheme = theme === 'kpmg-light' ? 
            (isDark ? 'kpmg-dark' : 'kpmg-light') : 
            (isDark ? 'dark' : 'default');
        
        document.documentElement.setAttribute('data-theme', actualTheme);
        localStorage.setItem('theme', actualTheme);
        showSuccessNotification('Theme applied successfully!');
    }

    // Initialize theme selection
    if (themeSelect) {
        themeSelect.value = currentTheme.includes('kpmg') ? 'kpmg-light' : 'default';
        applyThemeToPreview(themeSelect.value);

        // Handle theme changes in preview
        themeSelect.addEventListener('change', (e) => {
            applyThemeToPreview(e.target.value);
        });
    }

    // Handle apply theme button
    if (applyThemeBtn) {
        applyThemeBtn.addEventListener('click', () => {
            const selectedTheme = themeSelect.value;
            applyThemeGlobally(selectedTheme);
        });
    }

    // Handle edit theme button
    if (editThemeBtn) {
        editThemeBtn.addEventListener('click', () => {
            isEditing = !isEditing;
            previewFrame.setAttribute('data-editable', isEditing);
            editThemeBtn.textContent = isEditing ? 'Stop Editing' : 'Edit Theme';
            saveThemeChangesBtn.style.display = isEditing ? 'inline-block' : 'none';
            colorPickerPanel.classList.remove('show');
            selectedElement = null;
        });
    }

    // Handle preview element clicks
    if (previewFrame) {
        previewFrame.addEventListener('click', (e) => {
            if (!isEditing) return;
            
            const element = e.target.closest('.preview-element');
            if (!element) return;

            // Remove selected class from all elements
            previewFrame.querySelectorAll('.preview-element').forEach(el => {
                el.classList.remove('selected');
            });

            // Add selected class to clicked element
            element.classList.add('selected');
            selectedElement = element;

            // Update color picker
            const computedStyle = getComputedStyle(element);
            elementColor.value = rgbToHex(computedStyle.color);
            elementBackground.value = rgbToHex(computedStyle.backgroundColor);
            
            // Update selected element text
            selectedElementText.textContent = `Editing: ${element.dataset.element}`;
            
            // Show color picker panel
            colorPickerPanel.classList.add('show');
        });
    }

    // Handle color picker close
    if (closePickerBtn) {
        closePickerBtn.addEventListener('click', () => {
            colorPickerPanel.classList.remove('show');
            if (selectedElement) {
                selectedElement.classList.remove('selected');
                selectedElement = null;
            }
        });
    }

    // Handle color application
    if (applyColorBtn) {
        applyColorBtn.addEventListener('click', () => {
            if (!selectedElement) return;

            const elementType = selectedElement.dataset.element;
            const color = elementColor.value;
            const background = elementBackground.value;

            // Apply colors to the element
            selectedElement.style.color = color;
            selectedElement.style.backgroundColor = background;

            // Store the colors in CSS variables
            const varPrefix = `--color-${elementType}`;
            document.documentElement.style.setProperty(`${varPrefix}-text`, color);
            document.documentElement.style.setProperty(`${varPrefix}-bg`, background);
        });
    }

    // Handle theme save
    if (saveThemeBtn) {
        saveThemeBtn.addEventListener('click', () => {
            // Save theme logic here
            // For now, just close the modal
            colorPickerPanel.classList.remove('show');
            showSuccessNotification('Theme saved successfully!');
        });
    }

    // Handle theme editor cancel
    if (cancelThemeBtn) {
        cancelThemeBtn.addEventListener('click', () => {
            colorPickerPanel.classList.remove('show');
        });
    }

    // Handle save theme changes
    if (saveThemeChangesBtn) {
        saveThemeChangesBtn.addEventListener('click', async () => {
            const themeStyles = {};
            previewFrame.querySelectorAll('.preview-element').forEach(element => {
                const elementType = element.dataset.element;
                const computedStyle = getComputedStyle(element);
                themeStyles[`--color-${elementType}-text`] = computedStyle.color;
                themeStyles[`--color-${elementType}-bg`] = computedStyle.backgroundColor;
            });

            try {
                // Save theme changes to CSS
                await makeAdminRequest('/admin/update-theme', 'POST', {
                    theme: themeSelect.value,
                    styles: themeStyles
                });
                showSuccessNotification('Theme changes saved successfully!');
            } catch (error) {
                console.error('Error saving theme:', error);
                showSuccessNotification('Error saving theme changes');
            }
        });
    }

    // Handle create theme button
    if (createThemeBtn) {
        createThemeBtn.addEventListener('click', () => {
            createThemeModal.style.display = 'block';
            newThemeNameInput.value = '';
        });
    }

    // Handle create theme confirmation
    if (confirmCreateThemeBtn) {
        confirmCreateThemeBtn.addEventListener('click', async () => {
            const newThemeName = newThemeNameInput.value.trim();
            if (!newThemeName) {
                showSuccessNotification('Please enter a theme name');
                return;
            }

            try {
                // Create new theme based on current theme
                await makeAdminRequest('/admin/create-theme', 'POST', {
                    name: newThemeName,
                    baseTheme: themeSelect.value
                });
                
                // Add new theme to select dropdown
                const option = document.createElement('option');
                option.value = newThemeName;
                option.textContent = newThemeName;
                themeSelect.appendChild(option);
                themeSelect.value = newThemeName;
                
                // Apply the new theme to preview
                applyThemeToPreview(newThemeName);
                
                createThemeModal.style.display = 'none';
                showSuccessNotification('New theme created successfully!');
            } catch (error) {
                console.error('Error creating theme:', error);
                showSuccessNotification('Error creating new theme');
            }
        });
    }

    // Handle create theme cancellation
    if (cancelCreateThemeBtn) {
        cancelCreateThemeBtn.addEventListener('click', () => {
            createThemeModal.style.display = 'none';
        });
    }

    // Close modals when clicking outside
    window.addEventListener('click', (e) => {
        if (e.target === colorPickerPanel) {
            colorPickerPanel.classList.remove('show');
        }
        if (e.target === createThemeModal) {
            createThemeModal.style.display = 'none';
        }
    });

    // Helper function to convert RGB to Hex
    function rgbToHex(rgb) {
        if (!rgb || rgb === 'transparent' || rgb === 'none') return '#000000';
        
        // Check if already hex
        if (rgb.startsWith('#')) return rgb;
        
        // Convert rgb(r, g, b) to hex
        const rgbMatch = rgb.match(/^rgb\((\d+),\s*(\d+),\s*(\d+)\)$/);
        if (!rgbMatch) return '#000000';
        
        const r = parseInt(rgbMatch[1]);
        const g = parseInt(rgbMatch[2]);
        const b = parseInt(rgbMatch[3]);
        
        return '#' + ((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1);
    }

    // Function to load environment variables
    async function loadEnvironmentVariables() {
        try {
            // Use makeAdminRequest from common.js
            const data = await makeAdminRequest('/admin/env-variables', 'GET');
            
            if (data) {
                const envVariables = document.getElementById('env-variables');
                envVariables.innerHTML = '';
                
                Object.entries(data).forEach(([key, value]) => {
                    const field = document.createElement('div');
                    field.className = 'env-field';
                    field.innerHTML = `
                        <label for="${key}">${key}:</label>
                        <input type="text" id="${key}" name="${key}" value="${value || ''}" />
                    `;
                    envVariables.appendChild(field);
                });
            }
        } catch (error) {
            console.error('Error:', error);
            displayMessage(envMessage, error.message || 'An error occurred while loading environment variables', 'error');
            
            // Hide content if authentication fails
            if (error.message === 'Authentication failed' || error.message === 'Authentication cancelled') {
                envFormContainer.style.display = 'none';
                themeSection.style.display = 'none';
            }
        }
    }

    // Handle environment form submission
    if (envForm) {
        envForm.addEventListener('submit', (e) => {
            e.preventDefault();
            confirmModal.style.display = 'block';
        });
    }

    // Handle confirmation modal buttons
    if (confirmSaveButton) {
        confirmSaveButton.addEventListener('click', async () => {
            const formData = new FormData(envForm);
            const data = Object.fromEntries(formData.entries());

            try {
                // First verify the canvas if CANVAS_ID is being changed
                if (data.CANVAS_ID) {
                    try {
                        const verifyResponse = await fetch(`/verifycanvas/${data.CANVAS_ID}`);
                        const verifyData = await verifyResponse.json();
                        
                        if (!verifyData.exists) {
                            displayMessage(envMessage, 'Canvas ID does not exist on server.', 'error');
                            return;
                        }
                        
                        // If canvas name is different, update the form field
                        if (verifyData.name && data.CANVAS_NAME !== verifyData.name) {
                            const canvasNameInput = envForm.querySelector('[name="CANVAS_NAME"]');
                            if (canvasNameInput) {
                                canvasNameInput.value = verifyData.name;
                                data.CANVAS_NAME = verifyData.name;
                            }
                        }
                    } catch (verifyError) {
                        displayMessage(envMessage, 'Failed to verify Canvas ID.', 'error');
                        return;
                    }
                }

                const result = await makeAdminRequest('/admin/update-env', 'POST', data);
                if (result && result.success) {
                    showSuccessNotification('Environment variables updated successfully!');
                    confirmModal.style.display = 'none';
                    
                    // If any variables were auto-updated (like CANVAS_NAME), update the form
                    if (result.updatedVars) {
                        Object.entries(result.updatedVars).forEach(([key, value]) => {
                            const input = envForm.querySelector(`[name="${key}"]`);
                            if (input && input.value !== value) {
                                input.value = value;
                            }
                        });
                    }
                    
                    // Broadcast the environment update event
                    const event = new CustomEvent('env-updated', { 
                        detail: { 
                            canvas_id: data.CANVAS_ID,
                            canvas_name: data.CANVAS_NAME,
                            timestamp: Date.now()
                        } 
                    });
                    window.dispatchEvent(event);
                    
                    // Store the last update timestamp
                    localStorage.setItem('lastEnvUpdate', Date.now().toString());
                    
                    // Force reload all active endpoints with cache busting
                    const timestamp = Date.now();
                    await Promise.all([
                        fetch(`/get-zones?t=${timestamp}`, { 
                            headers: { 
                                'Cache-Control': 'no-cache',
                                'Pragma': 'no-cache'
                            }
                        }),
                        fetch(`/get-canvas-info?t=${timestamp}`, { 
                            headers: { 
                                'Cache-Control': 'no-cache',
                                'Pragma': 'no-cache'
                            }
                        })
                    ]);
                    
                    // Reload the page after a short delay to ensure all updates are applied
                    setTimeout(() => {
                        window.location.reload();
                    }, 1000);
                }
            } catch (error) {
                console.error('Error:', error);
                displayMessage(envMessage, error.message || 'An error occurred while updating environment variables', 'error');
            }
        });
    }

    if (cancelSaveButton) {
        cancelSaveButton.addEventListener('click', () => {
            confirmModal.style.display = 'none';
        });
    }

    // Add event listener for storage changes to sync across tabs
    window.addEventListener('storage', (e) => {
        if (e.key === 'lastEnvUpdate') {
            // Reload the page when environment is updated in another tab
            window.location.reload();
        }
    });

    // Initially hide content until authenticated
    envFormContainer.style.display = 'none';
    themeSection.style.display = 'none';

    // Trigger authentication and content loading immediately
    try {
        await loadEnvironmentVariables();
        // Show content after successful authentication
        envFormContainer.style.display = 'block';
        themeSection.style.display = 'block';
    } catch (error) {
        console.error('Initial authentication failed:', error);
    }

    // Add event listener for environment updates
    window.addEventListener('env-updated', async (event) => {
        try {
            // Reload environment variables
            await loadEnvironmentVariables();
            
            // Update navbar canvas info
            const canvasInfo = document.getElementById('canvasInfo');
            if (canvasInfo && event.detail.canvas_name) {
                canvasInfo.textContent = `Currently Connected to: ${event.detail.canvas_name}`;
            }
            
            // Force reload any cached data
            const endpoints = ['/get-zones', '/get-canvas-info'];
            await Promise.all(endpoints.map(endpoint => 
                fetch(endpoint, { headers: { 'Cache-Control': 'no-cache' }})
            ));
        } catch (error) {
            console.error('Error handling environment update:', error);
        }
    });
});
  