// /public/js/pages.js

document.addEventListener('DOMContentLoaded', () => {
    const messageDiv = document.getElementById("message");
  
    // Function to display a message
    function displayMessage(text, type) {
      messageDiv.textContent = text;
      messageDiv.className = `message ${type}`;
      messageDiv.style.display = "block";
    }
  
    // Function to clear messages
    function clearMessage() {
      messageDiv.textContent = "";
      messageDiv.className = "message";
      messageDiv.style.display = "none";
    }
  
    // Function to enable or disable all buttons
    function toggleButtons(disabled) {
      const buttons = document.querySelectorAll("button");
      buttons.forEach(btn => (btn.disabled = disabled));
    }
  
    // Function to fetch and populate SubZones dropdown
    async function fetchAndPopulateSubZones() {
      console.log("Fetching zones...");
      try {
        const response = await fetch('/get-zones', {
          method: 'GET',
          headers: {
            'Cache-Control': 'no-cache' // Prevent caching
          }
        });
        const data = await response.json();
        console.log("Data received from /get-zones:", data);
  
        if (data.success) {
          const subZoneSelect = document.getElementById('subZone');
          subZoneSelect.innerHTML = '<option value="">-- Select Zone --</option>'; // Clear existing options
  
          // Check if data.zones is an array
          if (Array.isArray(data.zones)) {
            // Sort zones by anchor_name
            const sortedZones = [...data.zones].sort((a, b) => {
              const nameA = (a.anchor_name || '').toLowerCase();
              const nameB = (b.anchor_name || '').toLowerCase();
              return nameA.localeCompare(nameB, undefined, { numeric: true });
            });
  
            sortedZones.forEach(zone => {
              const option = document.createElement('option');
              option.value = zone.id;
              option.textContent = zone.anchor_name;
              subZoneSelect.appendChild(option);
            });
            console.log("Dropdown updated with sorted zones.");
          } else {
            console.error("data.zones is not an array:", data.zones);
            displayMessage("Unexpected data format received.", "error");
          }
        } else {
          displayMessage('Failed to fetch zones.', 'error');
        }
      } catch (error) {
        console.error('Error fetching zones:', error);
        displayMessage('Error fetching zones.', 'error');
      }
    }
  
    // Call this function on page load to populate the SubZones dropdown
    window.addEventListener('load', () => {
      fetchAndPopulateSubZones();
    });
  
    // Function to create zones or subzones
    async function createZones() {
      clearMessage();
      toggleButtons(true); // Disable buttons
  
      const gridSize = parseInt(document.getElementById('gridSize').value);
      const gridPattern = document.getElementById('gridPattern').value;
      const subZoneId = document.getElementById('subZone').value;
      const subZoneArray = document.getElementById('subZoneArray').value;
  
      displayMessage("Creating zones, please wait...", "loading");
  
      try {
        const response = await fetch("/create-zones", {
          method: "POST",
          headers: { 
            "Content-Type": "application/json"
          },
          body: JSON.stringify({
            gridSize,
            gridPattern,
            subZoneId,
            subZoneArray
          })
        });
  
        const data = await response.json();
        console.log("Response from /create-zones:", data);
  
        if (data.success) {
          displayMessage(data.message, "success");
          console.log("Calling fetchAndPopulateSubZones after successful zone creation.");
          // Refresh the dropdown of subzones
          await fetchAndPopulateSubZones();
          console.log("fetchAndPopulateSubZones completed.");
        } else {
          displayMessage(data.error || "Failed to create zones.", "error");
        }
      } catch (error) {
        console.error("Error:", error);
        displayMessage("An error occurred while creating zones.", "error");
      } finally {
        toggleButtons(false); // Re-enable buttons
      }
    }
  
    // Function to delete zones
    async function deleteZones() {
      clearMessage();
      toggleButtons(true); // Disable buttons
  
      displayMessage("Deleting zones, please wait...", "loading");
  
      try {
        const response = await fetch("/delete-zones", {
          method: "DELETE",
          headers: { 
            "Content-Type": "application/json"
          }
        });
  
        const data = await response.json();
        console.log("Response from /delete-zones:", data);
  
        if (data.success) {
          displayMessage(data.message, "success");
          console.log("Calling fetchAndPopulateSubZones after successful zone deletion.");
          // Refresh the dropdown of subzones
          await fetchAndPopulateSubZones();
          console.log("fetchAndPopulateSubZones completed.");
        } else {
          displayMessage(data.error || "Failed to delete zones.", "error");
        }
      } catch (error) {
        console.error("Error:", error);
        displayMessage("An error occurred while deleting zones.", "error");
      } finally {
        toggleButtons(false); // Re-enable buttons
      }
    }
  
    // Attach event listeners
    document.getElementById("createZones").addEventListener("click", createZones);
    document.getElementById("deleteZones").addEventListener("click", deleteZones);
  });
  