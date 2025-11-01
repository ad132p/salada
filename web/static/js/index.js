import { Collapse } from 'flowbite';

// Wait for the DOM to be fully loaded before initializing components
document.addEventListener('DOMContentLoaded', () => {
    try {
        // 1. Get the target element (the menu content that collapses)
        // This ID now matches the <div> with id="navbar-hamburger"
        const $targetEl = document.getElementById('navbar-hamburger');

        // 2. Get the trigger element (the hamburger button)
        // We can use the data-collapse-toggle value to find the button, 
        // or just use the ID if we assign one. Since you only have the target ID, 
        // Flowbite's Collapse constructor will use the target ID as the trigger ID 
        // if no explicit trigger element is provided, but since your button 
        // explicitly uses data-collapse-toggle, let's look for that data attribute 
        // or use an ID if available.

        // Standard approach is to select the element with the data-collapse-toggle attribute 
        // whose value is 'navbar-hamburger'.
        const $triggerEl = document.querySelector('[data-collapse-toggle="navbar-hamburger"]');

        // Check if both elements exist
        if (!$targetEl || !$triggerEl) {
            console.error('Flowbite Collapse initialization failed: Target or trigger element not found. Check if the element IDs and data attributes match "navbar-hamburger".');
            return;
        }

        // 3. Optional options with custom callback functions
        const options = {
            onCollapse: () => {
                console.log('Mobile menu has been collapsed');
            },
            onExpand: () => {
                console.log('Mobile menu has been expanded');
            },
            onToggle: () => {
                console.log('Mobile menu has been toggled');
            },
        };

        // 4. Create a new instance of the Collapse object
        // NOTE: Flowbite automatically hooks up the trigger element's click 
        // handler to the target element's collapse/expand functionality 
        // when both are passed to the Collapse constructor.
        const collapse = new Collapse($targetEl, $triggerEl, options);

        console.log('Flowbite Collapse initialized programmatically for navbar-hamburger.');

        // Optional: Expose the collapse object globally for debugging
        window.mobileMenuCollapse = collapse;

    } catch (e) {
        console.error('Error during Flowbite Collapse setup:', e);
    }
});

