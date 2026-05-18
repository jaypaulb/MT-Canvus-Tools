// Utility functions for sorting lists
const sortUtils = {
    /**
     * Sort an array of items alphabetically by a specific property
     * @param {Array} items - Array of items to sort
     * @param {string} property - Property to sort by
     * @param {boolean} ascending - Sort direction
     * @returns {Array} Sorted array
     */
    sortAlphabetically: function(items, property, ascending = true) {
        return [...items].sort((a, b) => {
            const aValue = this.getPropertyValue(a, property);
            const bValue = this.getPropertyValue(b, property);
            
            if (ascending) {
                return aValue.localeCompare(bValue, undefined, { sensitivity: 'base' });
            } else {
                return bValue.localeCompare(aValue, undefined, { sensitivity: 'base' });
            }
        });
    },

    /**
     * Get nested property value using dot notation
     * @param {Object} obj - Object to get property from
     * @param {string} path - Property path (e.g. "user.name")
     * @returns {string} Property value
     */
    getPropertyValue: function(obj, path) {
        return path.split('.').reduce((acc, part) => {
            if (acc && typeof acc === 'object') {
                return acc[part] || '';
            }
            return '';
        }, obj).toString().toLowerCase();
    },

    /**
     * Initialize sorting for a list element
     * @param {string} listId - ID of the list element
     * @param {string} property - Property to sort by
     * @param {Function} renderItem - Function to render a single item
     */
    initializeSorting: function(listId, property, renderItem) {
        const list = document.getElementById(listId);
        if (!list) return;

        // Create header with sort button
        const header = document.createElement('div');
        header.className = 'sortable-list-header';
        
        const sortButton = document.createElement('button');
        sortButton.className = 'sort-button';
        sortButton.innerHTML = `Sort by ${property.split('.').pop()} <span class="sort-icon"></span>`;
        
        header.appendChild(sortButton);
        list.parentNode.insertBefore(header, list);

        let ascending = true;
        let items = Array.from(list.children);
        
        sortButton.addEventListener('click', () => {
            ascending = !ascending;
            sortButton.querySelector('.sort-icon').className = `sort-icon ${ascending ? 'asc' : 'desc'}`;
            
            const sortedItems = this.sortAlphabetically(items, property, ascending);
            
            // Clear and repopulate list
            list.innerHTML = '';
            sortedItems.forEach(item => {
                list.appendChild(renderItem(item));
            });
        });

        // Initial sort
        const sortedItems = this.sortAlphabetically(items, property);
        list.innerHTML = '';
        sortedItems.forEach(item => {
            list.appendChild(renderItem(item));
        });
    }
}; 