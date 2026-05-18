// /public/js/main-page.js
// Client selection and workspace monitoring

document.addEventListener("DOMContentLoaded", () => {
    const clientSelect = document.getElementById("clientSelect");
    const canvasInfo = document.getElementById("canvasInfo");
    const canvasName = document.getElementById("canvasName");
    const connectionStatus = document.getElementById("connectionStatus");
    const continueButton = document.getElementById("continueButton");
    const confirmationSpan = document.getElementById("confirmation");
    const errorSpan = document.getElementById("error");

    let workspaceSubscription = null;
    let currentCanvasId = null;
    let clientRefreshInterval = null;
    const CLIENT_REFRESH_INTERVAL = 15000; // 15 seconds

    // Initialize
    MessageSystem.initialize();
    loadClients();
    startClientRefresh();

    // Load available clients (preserves selection on refresh)
    async function loadClients(isRefresh = false) {
        try {
            if (!isRefresh) {
                displayConfirmation("Loading connected clients...");
            }

            const currentSelection = clientSelect.value;
            const response = await fetch("/api/clients");
            const data = await response.json();

            if (!data.success || !data.clients) {
                throw new Error(data.error || "Failed to load clients");
            }

            populateClientDropdown(data.clients, currentSelection);

            if (!isRefresh) {
                clearMessages();
            }

            if (data.clients.length === 0) {
                displayError("No clients connected to the server.");
            }
        } catch (error) {
            console.error("Error loading clients:", error);
            if (!isRefresh) {
                displayError("Failed to load clients: " + error.message);
                clearDropdown();
                addDropdownOption("", "Error loading clients");
            }
        }
    }

    // Start periodic client list refresh
    function startClientRefresh() {
        if (clientRefreshInterval) {
            clearInterval(clientRefreshInterval);
        }
        clientRefreshInterval = setInterval(() => {
            loadClients(true);
        }, CLIENT_REFRESH_INTERVAL);
    }

    // Stop client refresh (cleanup)
    function stopClientRefresh() {
        if (clientRefreshInterval) {
            clearInterval(clientRefreshInterval);
            clientRefreshInterval = null;
        }
    }

    // Clear dropdown options
    function clearDropdown() {
        while (clientSelect.firstChild) {
            clientSelect.removeChild(clientSelect.firstChild);
        }
    }

    // Add option to dropdown
    function addDropdownOption(value, text, dataset) {
        const option = document.createElement("option");
        option.value = value;
        option.textContent = text;
        if (dataset) {
            Object.keys(dataset).forEach(key => {
                option.dataset[key] = dataset[key];
            });
        }
        clientSelect.appendChild(option);
    }

    // Populate the client dropdown (optionally restore previous selection)
    function populateClientDropdown(clients, previousSelection = null) {
        clearDropdown();
        addDropdownOption("", "-- Select a display --");

        clients.forEach(client => {
            const displayName = client.name || client.hostname || client.id;
            addDropdownOption(client.id, displayName, {
                clientName: client.name || client.hostname || "Unknown"
            });
        });

        clientSelect.disabled = false;

        // Restore previous selection if it still exists
        if (previousSelection) {
            const optionExists = Array.from(clientSelect.options).some(
                opt => opt.value === previousSelection
            );
            if (optionExists) {
                clientSelect.value = previousSelection;
            }
        }
    }

    // Handle client selection
    clientSelect.addEventListener("change", async () => {
        const clientId = clientSelect.value;

        // Close existing subscription
        if (workspaceSubscription) {
            workspaceSubscription.close();
            workspaceSubscription = null;
        }

        // Reset UI
        canvasInfo.style.display = "none";
        continueButton.style.display = "none";
        continueButton.disabled = true;
        clearMessages();

        if (!clientId) {
            return;
        }

        try {
            displayConfirmation("Connecting to client workspace...");

            // Get current workspace
            const response = await fetch(`/api/clients/${clientId}/workspace`);
            const data = await response.json();

            if (!data.success || !data.workspace) {
                throw new Error(data.error || "Failed to get workspace");
            }

            const workspace = data.workspace;
            currentCanvasId = workspace.canvas_id;

            if (!currentCanvasId) {
                displayError("No canvas open on this client.");
                return;
            }

            // Get canvas details
            await switchToCanvas(currentCanvasId);

            // Start monitoring for changes
            startWorkspaceSubscription(clientId);

        } catch (error) {
            console.error("Error connecting to client:", error);
            displayError("Failed to connect: " + error.message);
        }
    });

    // Switch to a canvas
    async function switchToCanvas(canvasId) {
        try {
            // Notify server of canvas switch
            const response = await fetch("/api/canvas/switch", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ canvas_id: canvasId })
            });

            const data = await response.json();

            if (!data.success) {
                throw new Error(data.error || "Failed to switch canvas");
            }

            // Update UI
            canvasName.textContent = data.canvas_name || canvasId;
            canvasInfo.style.display = "block";
            continueButton.style.display = "block";
            continueButton.disabled = false;

            displayConfirmation("Connected to canvas: " + (data.canvas_name || canvasId));

        } catch (error) {
            console.error("Error switching canvas:", error);
            displayError("Failed to switch canvas: " + error.message);
        }
    }

    // Start SSE subscription for workspace changes
    function startWorkspaceSubscription(clientId) {
        console.log("Starting workspace subscription for client:", clientId);

        workspaceSubscription = new EventSource("/api/clients/" + clientId + "/workspace/subscribe");

        workspaceSubscription.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                console.log("Workspace event:", data);

                if (data.type === "connected") {
                    connectionStatus.textContent = "Monitoring";
                    connectionStatus.className = "status connected";
                }

                if (data.type === "canvas_change") {
                    if (data.canvas_id !== currentCanvasId) {
                        console.log("Canvas changed:", data.canvas_id);
                        currentCanvasId = data.canvas_id;

                        // Update UI
                        canvasName.textContent = data.canvas_name || data.canvas_id;
                        displayConfirmation("Canvas changed to: " + (data.canvas_name || data.canvas_id));

                        // Notify server
                        fetch("/api/canvas/switch", {
                            method: "POST",
                            headers: { "Content-Type": "application/json" },
                            body: JSON.stringify({
                                canvas_id: data.canvas_id,
                                canvas_name: data.canvas_name
                            })
                        });
                    }
                }

                if (data.type === "error") {
                    connectionStatus.textContent = "Error";
                    connectionStatus.className = "status error";
                }
            } catch (e) {
                console.error("Error parsing SSE data:", e);
            }
        };

        workspaceSubscription.onerror = (error) => {
            console.error("SSE connection error:", error);
            connectionStatus.textContent = "Disconnected";
            connectionStatus.className = "status disconnected";

            // Try to reconnect after a delay
            setTimeout(() => {
                if (clientSelect.value === clientId) {
                    console.log("Attempting to reconnect...");
                    startWorkspaceSubscription(clientId);
                }
            }, 5000);
        };
    }

    // Continue button handler
    continueButton.addEventListener("click", () => {
        if (currentCanvasId) {
            window.location.href = "/pages.html";
        }
    });

    // Message display helpers
    function displayConfirmation(message) {
        confirmationSpan.textContent = message;
        confirmationSpan.className = "message success";
        confirmationSpan.style.display = "inline";
        errorSpan.style.display = "none";
    }

    function displayError(message) {
        errorSpan.textContent = message;
        errorSpan.className = "message error";
        errorSpan.style.display = "inline";
        confirmationSpan.style.display = "none";
    }

    function clearMessages() {
        confirmationSpan.textContent = "";
        confirmationSpan.style.display = "none";
        errorSpan.textContent = "";
        errorSpan.style.display = "none";
    }

    // Cleanup on page unload
    window.addEventListener("beforeunload", () => {
        stopClientRefresh();
        if (workspaceSubscription) {
            workspaceSubscription.close();
        }
    });
});
