// public/js/macros.js

document.addEventListener("DOMContentLoaded", () => {
  // 1) Setup tab navigation
  setupTabs();

  // 2) Fetch zones to populate dropdowns
  fetchZones();

  // 3) Bind button clicks for Manage (Move/Copy/Delete)
  document.getElementById("moveButton").addEventListener("click", manageMove);
  document.getElementById("copyButton").addEventListener("click", manageCopy);
  document.getElementById("deleteButton").addEventListener("click", manageDelete);

  // 4) Bind Undelete logic
  document.getElementById("refreshDeletedRecordsBtn").addEventListener("click", refreshDeletedRecords);

  // 5) Bind Grouping logic
  document.getElementById("autoGridButton").addEventListener("click", autoGrid);
  document.getElementById("groupColorButton").addEventListener("click", groupByColor);
  document.getElementById("groupTitleButton").addEventListener("click", groupByTitle);

  // 6) Bind Pinning logic
  document.getElementById("pinAllButton").addEventListener("click", pinAll);
  document.getElementById("unpinAllButton").addEventListener("click", unpinAll);

  // 7) Color tolerance slider
  setupColorToleranceSlider();

  // Modal confirmation
  document.getElementById("undeleteConfirmBtn").addEventListener("click", confirmUndelete);
  document.getElementById("undeleteCancelBtn").addEventListener("click", hideUndeleteModal);

  // Initialize sorting functionality
  sortZoneDropdowns();
  sortDeletedRecords();
});

/* ------------------------------------------------------------------------- */
/* Globals                                                                   */
/* ------------------------------------------------------------------------- */

let selectedRecordId = null;
let selectedRecordWidgetCount = 0;
let selectedRecordTypes = {};

/* ------------------------------ TAB SWITCHING ------------------------------ */
function setupTabs() {
  const tabButtons = document.querySelectorAll('.tab-button');
  const tabContents = document.querySelectorAll('.tab-content');

  tabButtons.forEach(button => {
    button.addEventListener('click', () => {
      // Remove active class from all buttons and contents
      tabButtons.forEach(btn => btn.classList.remove('active'));
      tabContents.forEach(content => content.classList.remove('active'));

      // Add active class to clicked button and corresponding content
      button.classList.add('active');
      const tabId = button.getAttribute('data-tab');
      document.getElementById(`${tabId}-content`).classList.add('active');
    });
  });
}

/* ------------------------------ FETCH ZONES ------------------------------ */
async function fetchZones() {
  try {
    console.log("[fetchZones] Fetching zones and canvas details...");
    const res = await fetch("/get-zones", {
      headers: {
        'Cache-Control': 'no-cache'
      }
    });
    const data = await res.json();

    if (!data.success || !data.zones) {
      throw new Error("Failed to retrieve zones from the server.");
    }

    console.log(`[fetchZones] Retrieved ${data.zones.length} zones.`);

    // Populate the dropdowns with the updated zones list
    populateZoneDropdowns(data.zones);
  } catch (err) {
    console.error("[fetchZones] Error:", err.message);
    displayMessage(err.message);
  }
}

function populateZoneDropdowns(zones) {
  try {
    const dropdowns = {
      "manageSourceZone": document.getElementById("manageSourceZone"),
      "manageTargetZone": document.getElementById("manageTargetZone"),
      "undeleteTargetZone": document.getElementById("undeleteTargetZone"),
      "groupingSourceZone": document.getElementById("groupingSourceZone"),
      "pinningSourceZone": document.getElementById("pinningSourceZone")
    };

    // Verify all dropdowns exist
    Object.entries(dropdowns).forEach(([id, element]) => {
      if (!element) {
        throw new Error(`Dropdown element ${id} not found`);
      }
    });

    // Clear existing options
    Object.values(dropdowns).forEach(dropdown => {
      dropdown.innerHTML = '<option value="">Select a zone...</option>';
    });

    // Add zone options
    zones.forEach((zone) => {
      const zoneName = zone.anchor_name || `Zone ${zone.id}`;
      const option = document.createElement("option");
      option.value = zone.id;
      option.textContent = zoneName;

      // Add to each dropdown
      Object.values(dropdowns).forEach(dropdown => {
        dropdown.appendChild(option.cloneNode(true));
      });
    });

    sortZoneDropdowns();
  } catch (err) {
    console.error("[populateZoneDropdowns] Error:", err.message);
    displayMessage("Error populating zone dropdowns: " + err.message);
  }
}

/* ------------------------------ MANAGE ACTIONS ------------------------------ */
async function manageMove() {
  const sourceZoneId = document.getElementById("manageSourceZone").value;
  const targetZoneId = document.getElementById("manageTargetZone").value;
  if (!sourceZoneId || !targetZoneId) {
    displayMessage("Please select both Source and Target zones.");
    return;
  }
  try {
    const payload = { sourceZoneId, targetZoneId };
    const resp = await postJson("/api/macros/move", payload);
    displayMessage(resp.message);
  } catch (err) {
    displayMessage(err.message);
  }
}

async function manageCopy() {
  const sourceZoneId = document.getElementById("manageSourceZone").value;
  const targetZoneId = document.getElementById("manageTargetZone").value;
  if (!sourceZoneId || !targetZoneId) {
    displayMessage("Please select both Source and Target zones.");
    return;
  }
  try {
    const payload = { sourceZoneId, targetZoneId };
    console.log("manageCopy() => payload:", payload);
    const resp = await postJson("/api/macros/copy", payload);
    displayMessage(resp.message);
  } catch (err) {
    displayMessage(err.message);
  }
}

async function manageDelete() {
  const zoneId = document.getElementById("manageSourceZone").value;
  if (!zoneId) {
    displayMessage("Please select a Source zone for deletion.");
    return;
  }
  try {
    const payload = { zoneId };
    console.log("manageDelete() => payload:", payload);
    const resp = await postJson("/api/macros/delete", payload);
    displayMessage(resp.message);
  } catch (err) {
    displayMessage(err.message);
  }
}

/* ------------------------------------------------------------------------- */
/* Undelete + Delete History List                                           */
/* ------------------------------------------------------------------------- */

async function refreshDeletedRecords() {
  const listEl = document.getElementById("deletedRecordsList");
  listEl.innerHTML = "Loading deleted records...";
  
  try {
    const res = await fetch("/api/macros/deleted-records");
    if (!res.ok) {
      throw new Error("Failed to retrieve deleted records.");
    }
    const data = await res.json();
    if (!data.success || !data.records) {
      throw new Error("Invalid response retrieving deleted records.");
    }
    
    if (data.records.length === 0) {
      listEl.innerHTML = "No deleted records found.";
      return;
    }

    // Sort records by timestamp (most recent first)
    data.records.sort((a, b) => {
      if (a.timestamp < b.timestamp) return 1;
      if (a.timestamp > b.timestamp) return -1;
      return 0;
    });

    // Clear and populate the list
    listEl.innerHTML = "";
    data.records.forEach(record => {
      const div = document.createElement("div");
      div.classList.add("deletion-record");
      
      // Format the timestamp
      const date = new Date(record.timestamp);
      const formattedDate = date.toLocaleString('en-GB', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      });
      
      div.textContent = `Deleted at: ${formattedDate}, ID: ${record.recordId}`;
      div.dataset.recordId = record.recordId;
      div.dataset.timestamp = record.timestamp;
      
      // Add click handler
      div.addEventListener("click", handleDeletionRecordClick);
      listEl.appendChild(div);
    });
  } catch (err) {
    displayMessage(err.message);
    listEl.innerHTML = "Error loading records.";
  }
}

async function handleDeletionRecordClick(e) {
  const div = e.currentTarget;
  const recordId = div.dataset.recordId;
  const targetZoneId = document.getElementById("undeleteTargetZone").value;

  if (!targetZoneId) {
    displayMessage("Please select a target zone first.");
    return;
  }

  if (!recordId) {
    displayMessage("Invalid record selected.");
    return;
  }

  try {
    // Fetch details about the deleted record
    const res = await fetch(`/api/macros/deleted-details?recordId=${recordId}`);
    if (!res.ok) {
      throw new Error("Failed to fetch record details");
    }
    
    const data = await res.json();
    if (!data.success) {
      throw new Error(data.error || "Failed to load record details");
    }

    // Show confirmation dialog with details
    const msgEl = document.getElementById("undeleteConfirmMessage");
    const listEl = document.getElementById("undeleteWidgetList");
    
    msgEl.textContent = `Restore deleted record [${recordId}]?`;
    
    // Format the widget count and types
    let detailsHtml = `Will restore ${data.count} widgets:<br/>`;
    if (Object.keys(data.types).length > 0) {
      detailsHtml += "<ul>";
      for (const [type, count] of Object.entries(data.types)) {
        detailsHtml += `<li>${type}: ${count}</li>`;
      }
      detailsHtml += "</ul>";
    }
    
    listEl.innerHTML = detailsHtml;

    // Store the selected record ID for the confirmation handler
    selectedRecordId = recordId;

    // Show the modal
    document.getElementById("undeleteConfirmOverlay").style.display = "block";
  } catch (err) {
    displayMessage(err.message);
  }
}

async function confirmUndelete() {
  if (!selectedRecordId) {
    displayMessage("No record selected to undelete.");
    hideUndeleteModal();
    return;
  }

  const targetZoneId = document.getElementById("undeleteTargetZone").value;
  if (!targetZoneId) {
    displayMessage("Please select a target zone.");
    hideUndeleteModal();
    return;
  }

  try {
    const payload = { recordId: selectedRecordId, targetZoneId };
    const resp = await postJson("/api/macros/undelete", payload);
    displayMessage(resp.message || "Undelete successful.");
    
    // Refresh the deleted records list
    refreshDeletedRecords();
  } catch (err) {
    displayMessage(err.message);
  }
  
  hideUndeleteModal();
}

function hideUndeleteModal() {
  selectedRecordId = null;
  document.getElementById("undeleteConfirmOverlay").style.display = "none";
}

/* ------------------------------------------------------------------------- */
/* Grouping: AutoGrid, Color, Title                                         */
/* ------------------------------------------------------------------------- */

async function autoGrid() {
  const zoneId = document.getElementById("groupingSourceZone").value;
  if (!zoneId) {
    displayMessage("Please select a Source zone for Auto-Grid.");
    return;
  }
  try {
    // We now only send { zoneId }
    const payload = { zoneId };
    const resp = await postJson("/api/macros/auto-grid", payload);
    displayMessage(resp.message);
  } catch (err) {
    displayMessage(err.message);
  }
}

async function groupByColor() {
  const zoneId = document.getElementById("groupingSourceZone").value;
  const tolerance = document.getElementById("colorToleranceSlider").value;
  if (!zoneId) {
    displayMessage("Please select a Source zone for Group by Color.");
    return;
  }
  try {
    const payload = { zoneId, tolerance };
    const resp = await postJson("/api/macros/group-color", payload);
    displayMessage(resp.message);
  } catch (err) {
    displayMessage(err.message);
  }
}

async function groupByTitle() {
  const zoneId = document.getElementById("groupingSourceZone").value;
  if (!zoneId) {
    displayMessage("Please select a Source zone for Group by Title.");
    return;
  }
  try {
    // The server might do alphabetical grouping or sorting
    const payload = { zoneId };
    const resp = await postJson("/api/macros/group-title", payload);
    displayMessage(resp.message);
  } catch (err) {
    displayMessage(err.message);
  }
}

/* ------------------------------------------------------------------------- */
/* Color Tolerance Slider                                                   */
/* ------------------------------------------------------------------------- */

function setupColorToleranceSlider() {
  const slider = document.getElementById("colorToleranceSlider");
  const valSpan = document.getElementById("colorToleranceValue");
  if (!slider || !valSpan) return;
  slider.addEventListener("input", () => {
    valSpan.textContent = slider.value + "%";
  });
}

/* ------------------------------------------------------------------------- */
/* Generic Helpers                                                           */
/* ------------------------------------------------------------------------- */

async function postJson(url, bodyObj) {
  const resp = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(bodyObj),
  });
  if (!resp.ok) {
    const msg = await resp.text();
    throw new Error(msg || "Server error");
  }
  return resp.json();
}

function displayMessage(msg) {
  const msgEl = document.getElementById("message");
  if (msgEl) {
    msgEl.textContent = msg;
    msgEl.style.display = "block"; // Make sure message is visible
    
    // Optional: Hide message after 5 seconds
    setTimeout(() => {
      msgEl.style.display = "none";
    }, 5000);
  } else {
    console.log("[displayMessage]", msg);
  }
}

/* ------------------------------------------------------------------------- */
/* Sorting Functions                                                          */
/* ------------------------------------------------------------------------- */

// Sort zone dropdowns
function sortZoneDropdowns() {
  const dropdowns = [
    'manageSourceZone',
    'manageTargetZone',
    'undeleteTargetZone',
    'groupingSourceZone',
    'pinningSourceZone'
  ];

  dropdowns.forEach(id => {
    const dropdown = document.getElementById(id);
    if (dropdown) {
      const options = Array.from(dropdown.options);
      options.sort((a, b) => a.text.localeCompare(b.text, undefined, { numeric: true }));
      
      dropdown.innerHTML = '';
      options.forEach(option => dropdown.appendChild(option));
    }
  });
}

// Sort deleted records list
function sortDeletedRecords() {
  const list = document.getElementById('deletedRecordsList');
  if (!list) return;

  const records = Array.from(list.children);
  records.sort((a, b) => {
    const aTime = new Date(a.dataset.timestamp || 0);
    const bTime = new Date(b.dataset.timestamp || 0);
    return bTime - aTime; // Most recent first
  });

  list.innerHTML = '';
  records.forEach(record => list.appendChild(record));
}

// Sort widgets by title
function sortWidgetsByTitle(widgets) {
  return widgets.sort((a, b) => {
    const aTitle = a.title || '';
    const bTitle = b.title || '';
    return aTitle.localeCompare(bTitle, undefined, { numeric: true });
  });
}

/* ------------------------------------------------------------------------- */
/* Pinning Functions                                                         */
/* ------------------------------------------------------------------------- */

async function pinAll() {
  const zoneId = document.getElementById("pinningSourceZone").value;
  if (!zoneId) {
    displayMessage("Please select a Source zone for pinning.");
    return;
  }
  try {
    const payload = { zoneId };
    const resp = await postJson("/api/macros/pin-all", payload);
    displayMessage(resp.message);
  } catch (err) {
    displayMessage(err.message);
  }
}

async function unpinAll() {
  const zoneId = document.getElementById("pinningSourceZone").value;
  if (!zoneId) {
    displayMessage("Please select a Source zone for unpinning.");
    return;
  }
  try {
    const payload = { zoneId };
    const resp = await postJson("/api/macros/unpin-all", payload);
    displayMessage(resp.message);
  } catch (err) {
    displayMessage(err.message);
  }
}
