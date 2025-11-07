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



// web/static/js/index.js (Your Webpack Entry Point)
import React from 'react';
import { createRoot } from 'react-dom/client';
import CommentSection from './components/CommentSection';

// 1. Find the DOM element where React will be mounted
const container = document.getElementById('comment');

// Check if the container element exists on the page
if (container) {
    // 2. Get the necessary data (like the post slug) from the DOM element itself.
    // This is how React communicates with your server-rendered template.
    // We assume your HTML template renders a 'data-slug' attribute on the #comment element.
    const postID = container.getAttribute('data-id');
    const currentUserName = container.getAttribute('data-current-user');
    const comments = container.getAttribute('data-comments');

    // just checking
    console.log(currentUserName, 'hi')

    // 3. Create a root with the modern React 18+ API
    const root = createRoot(container);

    // 4. Render the CommentSection component into the root
    root.render(
        // Use React.StrictMode for development checks
        <React.StrictMode>
            <CommentSection
                initialPostID={postID}
                currentUserName={currentUserName}
                initialComments={comments}
            />
        </React.StrictMode>
    );

} else {
    // This is useful for debugging if the element is missing
    console.warn('React root element with id="comment" not found. The comment section will not render.');
}