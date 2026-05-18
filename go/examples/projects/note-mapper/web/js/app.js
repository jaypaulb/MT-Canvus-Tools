window.addEventListener('DOMContentLoaded', () => {
    // --- Tab Switching ---
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabSections = document.querySelectorAll('.tab-section');
    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            tabBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            tabSections.forEach(sec => sec.style.display = 'none');
            document.getElementById(`${btn.dataset.tab}-section`).style.display = '';
        });
    });
    document.getElementById('tab-config').classList.add('active');

    // --- Dark Mode Toggle ---
    const body = document.body;
    const darkmodeToggle = document.getElementById('darkmode-toggle');
    const darkmodeIcon = document.getElementById('darkmode-icon');
    function setDarkMode(isDark) {
        if (isDark) {
            body.classList.remove('light');
            body.classList.add('dark');
            darkmodeIcon.textContent = '🌙';
        } else {
            body.classList.remove('dark');
            body.classList.add('light');
            darkmodeIcon.textContent = '☀️';
        }
        localStorage.setItem('darkmode', isDark ? '1' : '0');
    }
    darkmodeToggle.addEventListener('click', () => {
        setDarkMode(!body.classList.contains('dark'));
    });
    // Default to dark mode
    setDarkMode(localStorage.getItem('darkmode') !== '0');

    // --- Persist and Restore Credentials ---
    const mcsServer = localStorage.getItem('mcsServer');
    const apiKey = localStorage.getItem('apiKey');
    if (mcsServer) document.getElementById('mcs-server').value = mcsServer;
    if (apiKey) document.getElementById('api-key').value = apiKey;

    // --- Credentials ---
    const credentialsForm = document.getElementById('credentials-form');
    const credentialsStatus = document.getElementById('credentials-status');
    credentialsForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const mcsServer = document.getElementById('mcs-server').value;
        const apiKey = document.getElementById('api-key').value;
        localStorage.setItem('mcsServer', mcsServer);
        localStorage.setItem('apiKey', apiKey);
        const res = await fetch('/api/set-credentials', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ mcsServer, apiKey })
        });
        const data = await res.json();
        credentialsStatus.textContent = data.status === 'ok' ? 'Credentials saved.' : (data.error || 'Error');
        if (data.status === 'ok') {
            await fetchCanvases();
            // Switch to Scan tab
            const scanTabBtn = document.getElementById('tab-scan');
            if (scanTabBtn) {
                scanTabBtn.classList.add('active');
                // Remove active from other tab buttons
                tabBtns.forEach(b => { if (b !== scanTabBtn) b.classList.remove('active'); });
                // Hide all tab sections
                tabSections.forEach(sec => sec.style.display = 'none');
                // Show scan section
                document.getElementById('scan-section').style.display = '';
            }
        }
    });

    // --- Fetch Canvases and Anchors ---
    const canvasSelect = document.getElementById('canvas-select');
    const anchorSelect = document.getElementById('zone-select');
    const anchorStatus = document.getElementById('zone-status');
    const canvasAnchorInfo = document.getElementById('canvas-anchor-info');

    async function fetchCanvases() {
        try {
            const res = await fetch('/api/get-canvases');
            if (!res.ok) {
                let data = {};
                try { data = await res.json(); } catch (e) {}
                const msg = data.error || 'Error fetching canvases. Please check your credentials or server connection.';
                anchorStatus.textContent = msg;
                console.error('[fetchCanvases] Server error:', msg);
                return;
            }
            const data = await res.json();
            canvasSelect.innerHTML = '';
            if (data.canvases) {
                data.canvases.forEach(c => {
                    const opt = document.createElement('option');
                    opt.value = c.id;
                    opt.textContent = c.name;
                    canvasSelect.appendChild(opt);
                });
            }
            if (canvasSelect.options.length > 0) {
                await fetchAnchors(canvasSelect.value);
            } else {
                anchorStatus.textContent = 'No canvases found. Please check your credentials or server connection.';
                console.warn('[fetchCanvases] No canvases found in response.');
            }
        } catch (err) {
            anchorStatus.textContent = 'Network or server error while fetching canvases. Please check your connection or server logs.';
            console.error('[fetchCanvases] Network or server error:', err);
            setTimeout(() => { anchorStatus.textContent = ''; }, 5000);
        }
    }

    async function fetchAnchors(canvasID) {
        anchorSelect.innerHTML = '';
        anchorStatus.textContent = 'Loading anchors...';
        try {
            const res = await fetch(`/api/get-anchors?canvasID=${encodeURIComponent(canvasID)}`);
            if (!res.ok) {
                const data = await res.json();
                anchorStatus.textContent = data.error || 'Error fetching anchors.';
                return;
            }
            const data = await res.json();
            if (Array.isArray(data.anchors)) {
                anchorData = data.anchors.map(a => ({
                    id: a.id,
                    name: a.name
                }));
                anchorSelect.innerHTML = '';
                anchorData.forEach(a => {
                    const opt = document.createElement('option');
                    opt.value = a.id;
                    opt.textContent = a.name;
                    anchorSelect.appendChild(opt);
                });
                anchorStatus.textContent = '';
            } else {
                anchorStatus.textContent = 'No anchors found for this canvas.';
            }
        } catch (err) {
            anchorStatus.textContent = 'Error fetching anchors.';
        }
        updateCanvasAnchorInfo();
    }

    async function fetchAnchorInfo(canvasID, anchorID) {
        anchorStatus.textContent = 'Refreshing anchor details...';
        try {
            const url = `/api/get-anchor-info?canvasID=${encodeURIComponent(canvasID)}&anchorID=${encodeURIComponent(anchorID)}`;
            const res = await fetch(url);
            if (!res.ok) {
                let errorText = '';
                try { errorText = await res.text(); } catch (e) {}
                console.error('[fetchAnchorInfo] Error response:', errorText);
                anchorStatus.textContent = 'Error fetching anchor details.';
                return null;
            }
            const anchor = await res.json();
            if (!anchor.id || !anchor.name) {
                console.error('Anchor details missing required fields:', anchor);
                anchorStatus.textContent = 'Anchor details missing required fields.';
                return null;
            }
            anchorStatus.textContent = '';
            return anchor;
        } catch (err) {
            console.error('[fetchAnchorInfo] Network or server error:', err);
            anchorStatus.textContent = 'Network or server error while fetching anchor details.';
            return null;
        }
    }

    // updateCanvasAnchorInfo now fetches anchor info for the selected anchor
    async function updateCanvasAnchorInfo() {
        const anchorOption = anchorSelect.options[anchorSelect.selectedIndex];
        if (!anchorOption) return;
        const canvasID = canvasSelect.value;
        const anchorID = anchorOption.value;
        const anchor = await fetchAnchorInfo(canvasID, anchorID);
        if (anchor) {
            const x = Math.floor(anchor.x);
            const y = Math.floor(anchor.y);
            anchorOption.textContent = `${anchor.name} [${anchor.width}x${anchor.height} @ ${x}, ${y}]`;
        } else {
            anchorOption.textContent = anchorOption.value;
        }
    }

    if (canvasSelect) {
        canvasSelect.addEventListener('change', async () => {
            await fetchAnchors(canvasSelect.value); // repopulate anchor dropdown
            if (anchorSelect.options.length > 0) {
                anchorSelect.selectedIndex = 0; // select first anchor
                await updateCanvasAnchorInfo(); // fetch info for new anchor
            } else {
                // No anchors for this canvas, clear anchor info display
                // Optionally, set anchorStatus or clear anchorSelect
                anchorStatus.textContent = 'No anchors found for this canvas.';
            }
        });
    }

    // --- Image Capture/Upload & Auto-Scan ---
    const imageInput = document.getElementById('image-input');
    const imageStatus = document.getElementById('image-status');
    const preview = document.getElementById('preview');
    const uploadBtn = document.getElementById('upload-btn');
    let uploadedImage = null;
    let lastScanData = null;
    let selectedNotes = [];
    const imageInputLabel = document.getElementById('image-input-label');

    // Debug: Check if elements are found
    console.log('[DOM Check] imageInput:', imageInput);
    console.log('[DOM Check] uploadBtn:', uploadBtn);
    console.log('[DOM Check] imageInputLabel:', imageInputLabel);
    console.log('[DOM Check] preview:', preview);

    // Make Choose File label trigger file input
    imageInput.addEventListener('change', async (e) => {
        console.log('[imageInput] File selection changed');
        const file = e.target.files[0];
        console.log('[imageInput] Selected file:', file);
        
        if (file) {
            console.log('[imageInput] File details:', {
                name: file.name,
                size: file.size,
                type: file.type
            });
            
            const reader = new FileReader();
            reader.onload = function(evt) {
                console.log('[imageInput] FileReader loaded, setting preview');
                preview.src = evt.target.result;
                preview.style.display = 'block';
            };
            reader.readAsDataURL(file);
            uploadedImage = file;
            
            // Show upload button
            console.log('[imageInput] Showing upload button');
            uploadBtn.style.display = 'inline-block';
            imageStatus.textContent = 'File selected. Click Upload to process.';
            console.log('[imageInput] Upload button display style:', uploadBtn.style.display);
        } else {
            console.log('[imageInput] No file selected');
        }
    });

    // Handle upload button click
    uploadBtn.addEventListener('click', async () => {
        if (!uploadedImage) {
            imageStatus.textContent = 'No file selected';
            return;
        }

        imageStatus.textContent = 'Uploading and processing...';
        const formData = new FormData();
        formData.append('image', uploadedImage);
        
        // Get selected anchor info for zone parameters
        const anchorOption = anchorSelect.options[anchorSelect.selectedIndex];
        let zoneWidth = 640, zoneHeight = 480, zoneX = 0, zoneY = 0, zoneScale = 1;
        if (anchorOption) {
            const anchor = anchorData.find(a => a.id === anchorOption.value);
            if (anchor) {
                zoneWidth = Math.round(anchor.width);
                zoneHeight = Math.round(anchor.height);
                zoneX = Math.round(anchor.x);
                zoneY = Math.round(anchor.y);
                zoneScale = anchor.scale !== undefined ? Math.round(anchor.scale * 100) / 100 : 1;
            }
        }
        
        // Add zone parameters to the upload request
        formData.append('zoneDimensions', JSON.stringify([zoneWidth, zoneHeight]));
        formData.append('zoneLocation', JSON.stringify([zoneX, zoneY]));
        formData.append('zoneScale', JSON.stringify(zoneScale));
        
        try {
            console.log('[uploadBtn] Starting upload and processing...');
            const res = await fetch('/api/upload-image', {
                method: 'POST',
                body: formData
            });
            console.log('[uploadBtn] Response status:', res.status);
            
            if (!res.ok) {
                const errorText = await res.text();
                console.error('[uploadBtn] HTTP error response:', errorText);
                throw new Error(`HTTP error! status: ${res.status}, response: ${errorText}`);
            }
            
            const data = await res.json();
            console.log('[uploadBtn] Response data:', data);
            
            if (data.status === 'complete' && data.notes) {
                console.log('[uploadBtn] Upload successful, processing notes:', data.notes);
                console.log('[uploadBtn] Number of notes found:', data.notes.length);
                
                // Process the extracted notes
                lastScanData = data;
                console.log('[uploadBtn] Set lastScanData:', lastScanData);
                
                console.log('[uploadBtn] Calling renderThumbnails with notes:', data.notes);
                renderThumbnails(data.notes || []);
                
                imageStatus.textContent = `Processing complete. Found ${data.notes.length} notes.`;
                console.log('[uploadBtn] Updated image status');
            } else if (data.error) {
                console.log('[uploadBtn] Upload failed with error:', data.error);
                imageStatus.textContent = 'Processing failed: ' + data.error;
            } else {
                console.log('[uploadBtn] Unexpected response:', data);
                imageStatus.textContent = data.message || 'Processing complete';
            }
        } catch (err) {
            console.error('[uploadBtn] Upload/processing error:', err);
            imageStatus.textContent = 'Upload/processing failed: ' + err.message;
        }
    });

    // --- Thumbnails, Selection, Select All/Deselect All ---
    const scanResultsContainer = document.getElementById('scan-results-container');
    const thumbnailsDiv = document.getElementById('thumbnails');
    const selectAllBtn = document.getElementById('select-all');
    const deselectAllBtn = document.getElementById('deselect-all');
    const createBtn = document.getElementById('create-notes');
    const createStatus = document.getElementById('create-status');

    function renderThumbnails(notes) {
        console.log('[renderThumbnails] Called with notes:', notes);
        console.log('[renderThumbnails] Notes is array:', Array.isArray(notes));
        console.log('[renderThumbnails] Notes length:', notes ? notes.length : 'null');
        
        if (!Array.isArray(notes) || notes.length === 0) {
            console.log('[renderThumbnails] No notes to render, hiding interface');
            if (scanResultsContainer) scanResultsContainer.style.display = 'none';
            if (createBtn) createBtn.style.display = 'none';
            return;
        }
        
        console.log('[renderThumbnails] Showing scan results container');
        if (scanResultsContainer) scanResultsContainer.style.display = '';
        if (createBtn) createBtn.style.display = '';
        if (thumbnailsDiv) thumbnailsDiv.innerHTML = '';
        
        // If selectedNotes is empty, select all by default
        if (!selectedNotes || selectedNotes.length === 0) {
            selectedNotes = notes.map((n, i) => i);
            console.log('[renderThumbnails] Selected all notes by default:', selectedNotes);
        }
        
        console.log('[renderThumbnails] Creating thumbnails for', notes.length, 'notes');
        notes.forEach((note, idx) => {
            console.log(`[renderThumbnails] Processing note ${idx}:`, note);
            const thumb = document.createElement('div');
            thumb.className = 'thumbnail';
            if (note.background_color) thumb.style.background = note.background_color;
            thumb.innerHTML = `
                <input type="checkbox" class="note-checkbox" data-idx="${idx}" ${selectedNotes.includes(idx) ? 'checked' : ''}>
                <div><strong>${note.text || 'Note'}</strong></div>
                <div style="font-size:0.8em;">${note.size?.width || 0}x${note.size?.height || 0}</div>
            `;
            if (thumbnailsDiv) {
                thumbnailsDiv.appendChild(thumb);
                console.log(`[renderThumbnails] Added thumbnail ${idx} to DOM`);
            } else {
                console.error('[renderThumbnails] thumbnailsDiv not found!');
            }
        });
        
        console.log('[renderThumbnails] Setting up checkbox listeners');
        // Checkbox listeners
        if (thumbnailsDiv) {
            thumbnailsDiv.querySelectorAll('.note-checkbox').forEach(cb => {
                cb.addEventListener('change', (e) => {
                    const idx = parseInt(e.target.dataset.idx);
                    if (e.target.checked) {
                        if (!selectedNotes.includes(idx)) selectedNotes.push(idx);
                    } else {
                        selectedNotes = selectedNotes.filter(i => i !== idx);
                    }
                    console.log('[renderThumbnails] Updated selectedNotes:', selectedNotes);
                });
            });
        }
        
        console.log('[renderThumbnails] Calling updateCanvasAnchorInfo');
        updateCanvasAnchorInfo();
        console.log('[renderThumbnails] Function complete');
    }

    if (selectAllBtn) {
        selectAllBtn.addEventListener('click', () => {
            if (!thumbnailsDiv) return;
            const checkboxes = thumbnailsDiv.querySelectorAll('.note-checkbox');
            selectedNotes = (lastScanData?.notes || []).map((_, i) => i);
            checkboxes.forEach(cb => { cb.checked = true; });
        });
    } else {
        console.error('selectAllBtn not found in DOM!');
    }

    if (deselectAllBtn) {
        deselectAllBtn.addEventListener('click', () => {
            if (!thumbnailsDiv) return;
            const checkboxes = thumbnailsDiv.querySelectorAll('.note-checkbox');
            selectedNotes = [];
            checkboxes.forEach(cb => { cb.checked = false; });
        });
    } else {
        console.error('deselectAllBtn not found in DOM!');
    }

    // --- Camera Capture ---
    const captureBtn = document.getElementById('capture-btn');
    const cameraStream = document.getElementById('camera-stream');
    const captureCanvas = document.getElementById('capture-canvas');
    const cameraContainer = document.getElementById('camera-container');
    const closeCameraBtn = document.getElementById('close-camera');
    const switchCameraBtn = document.getElementById('switch-camera');
    let stream = null;
    let currentFacingMode = 'environment'; // default to rear camera
    let videoInputDevices = [];
    let currentDeviceIndex = 0;

    // Check if getUserMedia is supported
    function checkCameraSupport() {
        if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
            console.error('[Camera] getUserMedia not supported');
            imageStatus.textContent = '❌ Camera not supported on this browser/device.';
            return false;
        }
        
        // Check HTTPS requirement (except localhost)
        if (location.protocol !== 'https:' && location.hostname !== 'localhost' && location.hostname !== '127.0.0.1') {
            console.error('[Camera] HTTPS required for camera access');
            imageStatus.textContent = '❌ Camera requires HTTPS connection (except localhost).';
            return false;
        }
        
        console.log('[Camera] Camera support check passed');
        imageStatus.textContent = '✅ Camera support check passed...';
        return true;
    }

    // Simple camera test function
    async function testCameraAccess() {
        console.log('[Camera] Testing basic camera access...');
        imageStatus.textContent = '🔍 Testing basic camera access...';
        try {
            const testStream = await navigator.mediaDevices.getUserMedia({ video: true });
            console.log('[Camera] Test successful - camera access granted');
            imageStatus.textContent = '✅ Camera test successful!';
            testStream.getTracks().forEach(track => track.stop());
            return true;
        } catch (err) {
            console.error('[Camera] Test failed:', err);
            imageStatus.textContent = '❌ Camera test failed: ' + err.message;
            return false;
        }
    }

    async function getVideoInputDevices() {
        try {
            console.log('[Camera] Enumerating video devices...');
            imageStatus.textContent = '📹 Looking for cameras...';
            const devices = await navigator.mediaDevices.enumerateDevices();
            const videoDevices = devices.filter(device => device.kind === 'videoinput');
            console.log('[Camera] Found video devices:', videoDevices.length, videoDevices);
            imageStatus.textContent = `📹 Found ${videoDevices.length} camera(s)`;
            return videoDevices;
        } catch (err) {
            console.error('[Camera] Error enumerating devices:', err);
            imageStatus.textContent = '❌ Error accessing camera devices: ' + err.message;
            return [];
        }
    }

    async function startCamera(constraints) {
        console.log('[Camera] Starting camera with constraints:', constraints);
        imageStatus.textContent = '📷 Starting camera...';
        
        if (stream) {
            console.log('[Camera] Stopping existing stream');
            stream.getTracks().forEach(track => track.stop());
        }

        imageStatus.textContent = '🔐 Requesting camera permission...';
        
        try {
            console.log('[Camera] Calling getUserMedia...');
            imageStatus.textContent = '⏳ Waiting for permission popup...';
            
            // Try with the provided constraints first
            try {
                stream = await navigator.mediaDevices.getUserMedia(constraints);
            } catch (err) {
                console.log('[Camera] First attempt failed:', err.name);
                
                // If first attempt fails, try with basic constraints
                if (err.name === 'OverconstrainedError' || err.name === 'ConstraintNotSatisfiedError') {
                    console.log('[Camera] Trying with basic constraints...');
                    stream = await navigator.mediaDevices.getUserMedia({ video: true });
                } else {
                    throw err; // Re-throw if it's not a constraint error
                }
            }
            
            console.log('[Camera] Camera access granted, stream:', stream);
            
            cameraStream.srcObject = stream;
            cameraContainer.style.display = 'flex';
            cameraStream.style.display = 'block';
            preview.style.display = 'none';
            
            // Log video track info
            const videoTrack = stream.getVideoTracks()[0];
            if (videoTrack) {
                const settings = videoTrack.getSettings();
                console.log('[Camera] Video track settings:', settings);
                imageStatus.textContent = `✅ Camera ready! (${settings.width}x${settings.height}) Click to capture.`;
            } else {
                imageStatus.textContent = '✅ Camera ready! Click to capture.';
            }
            
        } catch (err) {
            console.error('[Camera] getUserMedia error:', err);
            console.error('[Camera] Error name:', err.name);
            console.error('[Camera] Error message:', err.message);
            
            let userMessage = '❌ Camera failed: ';
            let debugInfo = `\n🔧 Debug: ${err.name} - ${err.message}`;
            
            switch (err.name) {
                case 'NotFoundError':
                case 'DevicesNotFoundError':
                    userMessage += 'No camera found on this device.' + debugInfo;
                    break;
                case 'NotReadableError':
                case 'TrackStartError':
                    userMessage += 'Camera is being used by another app.' + debugInfo;
                    break;
                case 'OverconstrainedError':
                case 'ConstraintNotSatisfiedError':
                    userMessage += 'Camera settings not supported.' + debugInfo;
                    break;
                case 'NotAllowedError':
                case 'PermissionDeniedError':
                    userMessage += 'Permission denied! Check browser settings or try refreshing.' + debugInfo;
                    break;
                case 'SecurityError':
                    userMessage += 'Security blocked. HTTPS required?' + debugInfo;
                    break;
                default:
                    userMessage += (err.message || 'Unknown error.') + debugInfo;
            }
            
            imageStatus.textContent = userMessage;
            stream = null;
        }
    }

    if (captureBtn) {
        captureBtn.addEventListener('click', async () => {
            console.log('[Camera] Capture button clicked');
            imageStatus.textContent = '🎯 Capture button clicked...';
            
            if (!checkCameraSupport()) {
                return;
            }
            
            if (!stream) {
                console.log('[Camera] No active stream, requesting camera access...');
                imageStatus.textContent = '🚀 Starting camera setup...';
                
                try {
                    videoInputDevices = await getVideoInputDevices();
                    
                    if (videoInputDevices.length === 0) {
                        imageStatus.textContent = '❌ No camera devices found on this device.';
                        return;
                    }
                    
                    console.log('[Camera] Using', videoInputDevices.length, 'available camera(s)');
                    imageStatus.textContent = `🎥 Setting up camera (${videoInputDevices.length} available)...`;
                    
                    // Start with basic constraints
                    let constraints = { 
                        video: { 
                            facingMode: currentFacingMode
                        } 
                    };
                    
                    // If only one camera, use deviceId
                    if (videoInputDevices.length === 1) {
                        console.log('[Camera] Single camera detected, using deviceId');
                        imageStatus.textContent = '📱 Single camera detected, configuring...';
                        constraints = { 
                            video: { 
                                deviceId: { exact: videoInputDevices[0].deviceId }
                            } 
                        };
                    }
                    
                    await startCamera(constraints);
                    
                } catch (err) {
                    console.error('[Camera] Error in capture flow:', err);
                    imageStatus.textContent = '❌ Camera setup failed: ' + err.message + '\n🔧 Try refreshing the page or check browser settings.';
                }
                
            } else {
                console.log('[Camera] Closing active camera');
                imageStatus.textContent = '📷 Closing camera...';
                closeCamera();
            }
        });
    } else {
        console.error('[Camera] captureBtn not found in DOM!');
        imageStatus.textContent = '❌ ERROR: Capture button not found in page!';
    }

    if (switchCameraBtn) {
        switchCameraBtn.addEventListener('click', async () => {
            console.log('[Camera] Switch camera button clicked');
            
            try {
                videoInputDevices = await getVideoInputDevices();
                
                if (videoInputDevices.length > 1) {
                    // Cycle to next camera
                    currentDeviceIndex = (currentDeviceIndex + 1) % videoInputDevices.length;
                    const deviceId = videoInputDevices[currentDeviceIndex].deviceId;
                    console.log('[Camera] Switching to device:', deviceId);
                    await startCamera({ 
                        video: { 
                            deviceId: { exact: deviceId },
                            width: { ideal: 1280 },
                            height: { ideal: 720 }
                        } 
                    });
                } else {
                    // Toggle facingMode if only one camera
                    currentFacingMode = currentFacingMode === 'environment' ? 'user' : 'environment';
                    console.log('[Camera] Toggling facingMode to:', currentFacingMode);
                    await startCamera({ 
                        video: { 
                            facingMode: currentFacingMode,
                            width: { ideal: 1280 },
                            height: { ideal: 720 }
                        } 
                    });
                }
            } catch (err) {
                console.error('[Camera] Error switching camera:', err);
                imageStatus.textContent = 'Error switching camera: ' + err.message;
            }
        });
    }

    // Always show close button when camera is open
    closeCameraBtn.addEventListener('click', closeCamera);
    function closeCamera() {
        console.log('[Camera] Closing camera');
        if (stream) {
            stream.getTracks().forEach(track => {
                console.log('[Camera] Stopping track:', track.kind);
                track.stop();
            });
            stream = null;
        }
        cameraStream.style.display = 'none';
        cameraContainer.style.display = 'none';
        imageStatus.textContent = '';
    }

    // Capture image from video stream
    cameraStream.addEventListener('click', () => {
        if (!stream) return;
        captureCanvas.width = cameraStream.videoWidth;
        captureCanvas.height = cameraStream.videoHeight;
        captureCanvas.getContext('2d').drawImage(cameraStream, 0, 0);
        captureCanvas.toBlob(async (blob) => {
            preview.src = captureCanvas.toDataURL('image/png');
            preview.style.display = 'block';
            preview.style.margin = '1em auto 0 auto'; // Center preview
            cameraStream.style.display = 'none';
            cameraContainer.style.display = 'none';
            imageStatus.textContent = 'Uploading and processing photo...';
            if (stream) {
                stream.getTracks().forEach(track => track.stop());
                stream = null;
            }
            
            // Get selected anchor info for zone parameters
            const anchorOption = anchorSelect.options[anchorSelect.selectedIndex];
            let zoneWidth = 640, zoneHeight = 480, zoneX = 0, zoneY = 0, zoneScale = 1;
            if (anchorOption) {
                const anchor = anchorData.find(a => a.id === anchorOption.value);
                if (anchor) {
                    zoneWidth = Math.round(anchor.width);
                    zoneHeight = Math.round(anchor.height);
                    zoneX = Math.round(anchor.x);
                    zoneY = Math.round(anchor.y);
                    zoneScale = anchor.scale !== undefined ? Math.round(anchor.scale * 100) / 100 : 1;
                }
            }
            
            const formData = new FormData();
            formData.append('image', blob, 'capture.png');
            formData.append('zoneDimensions', JSON.stringify([zoneWidth, zoneHeight]));
            formData.append('zoneLocation', JSON.stringify([zoneX, zoneY]));
            formData.append('zoneScale', JSON.stringify(zoneScale));
            
            try {
                const res = await fetch('/api/upload-image', {
                    method: 'POST',
                    body: formData
                });
                
                if (!res.ok) {
                    const errorText = await res.text();
                    throw new Error(`HTTP error! status: ${res.status}, response: ${errorText}`);
                }
                
                const data = await res.json();
                if (data.status === 'complete' && data.notes) {
                    lastScanData = data;
                    renderThumbnails(data.notes || []);
                    imageStatus.textContent = `Processing complete. Found ${data.notes.length} notes.`;
                } else if (data.error) {
                    imageStatus.textContent = 'Processing failed: ' + data.error;
                } else {
                    imageStatus.textContent = data.message || 'Processing complete';
                }
            } catch (err) {
                console.error('[camera] Upload/processing error:', err);
                imageStatus.textContent = 'Upload/processing failed: ' + err.message;
            }
        }, 'image/png');
    });

    // --- Create Notes ---
    async function createNotes() {
        if (!lastScanData || selectedNotes.length === 0) {
            createStatus.textContent = 'Select at least one note.';
            return;
        }
        const canvasID = canvasSelect.value;
        const zoneID = anchorSelect.value;
        if (!canvasID || !zoneID) {
            createStatus.textContent = 'Please select both a canvas and a zone before creating notes.';
            return;
        }
        const anchor = await fetchAnchorInfo(canvasID, zoneID);
        if (!anchor) {
            createStatus.textContent = 'Failed to refresh anchor details.';
            return;
        }
        // Send notes as detected (raw), backend will handle transformation
        const notesToSend = selectedNotes.map(i => (lastScanData.notes ? lastScanData.notes[i] : lastScanData[i]));
        try {
            const res = await fetch('/api/create-notes', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    canvasID,
                    zoneID,
                    notes: notesToSend,
                    imageWidth: lastScanData.imageWidth,
                    imageHeight: lastScanData.imageHeight
                })
            });
            const data = await res.json();
            if (res.ok && data.status) {
                createStatus.textContent = data.status;
            } else {
                createStatus.textContent = data.error || 'Error creating notes. Please check your canvas/zone selection and try again.';
            }
        } catch (err) {
            console.error('[Create Notes] Network or server error:', err);
            createStatus.textContent = 'Network or server error while creating notes.';
        }
    }
    if (!createBtn._bound) {
        createBtn.addEventListener('click', createNotes);
        createBtn._bound = true;
    }
});
