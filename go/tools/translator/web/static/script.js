document.addEventListener('DOMContentLoaded', () => {
    const flagImage = document.getElementById('flag-image');
    const languageDropdown = document.getElementById('language-dropdown');
    const languageOptions = languageDropdown.querySelectorAll('li');
    const messageArea = document.getElementById('message-area');

    // Helper function to display messages
    function showMessage(message, type = 'info') {
        messageArea.textContent = message;
        messageArea.className = type; // 'info', 'success', 'error'
        messageArea.classList.remove('hidden');
    }

    function hideMessage() {
        messageArea.classList.add('hidden');
    }

    // FR.1.5: Map language names to flag image URLs from Wikimedia Commons
    const flagImageMap = {
        "Mandarin Chinese": "/flags/Flag_of_the_People's_Republic_of_China.svg",
        "Spanish": "/flags/Flag_of_Spain.svg",
        "English": "/flags/Flag_of_the_United_States.svg", // Using US flag as a common English flag
        "Hindi": "/flags/Flag_of_India.svg",
        "Arabic": "/flags/Flag_of_the_Arab_League_(1-2).svg",
        "Klingon": "/flags/Klingon_Empire_Flag.svg"
    };

    // FR.1.3: Toggle visibility of the language dropdown on flag click
    flagImage.addEventListener('click', () => {
        languageDropdown.classList.toggle('hidden');
    });

    languageOptions.forEach(option => {
        option.addEventListener('click', async () => {
            const selectedLanguage = option.dataset.lang;
            console.log(`Selected language: ${selectedLanguage}`);

            // FR.1.5: Change the src of the main flag image
            const newFlagSrc = flagImageMap[selectedLanguage];
            if (newFlagSrc) {
                flagImage.src = newFlagSrc;
            } else {
                console.warn(`No flag image found for ${selectedLanguage}`);
            }

            // Hide the dropdown after selection
            languageDropdown.classList.add('hidden');

            // NFR.5.3.2: Implement a basic loading indicator
            showMessage('Translating... Please wait.', 'info');

            try {
                // Send POST request to the /translate endpoint
                const response = await fetch('/translate', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ language: selectedLanguage }),
                });

                if (response.ok) {
                    const result = await response.text();
                    console.log('Translation request successful:', result);
                    showMessage('Translation initiated successfully!', 'success');
                } else {
                    console.error('Translation request failed:', response.statusText);
                    showMessage(`Translation failed! Error: ${response.statusText}`, 'error');
                }
            } catch (error) {
                console.error('Error sending translation request:', error);
                showMessage('An error occurred during translation.', 'error');
            } finally {
                // Hide loading indicator after a short delay
                setTimeout(hideMessage, 3000);
            }
        });
    });
}); 