import { Collapse } from 'flowbite';
import hljs from 'highlight.js';

// Initialize syntax highlighting
hljs.highlightAll();

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

        const collapse = new Collapse($targetEl, $triggerEl);

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
import LikeButton from './components/LikeButton';
import HeroImageCropper from './components/HeroImageCropper';
import VideoStreamer from './components/VideoStreamer';
import RoomsList from './components/RoomsList';
import VideoViewer from './components/VideoViewer';

// 1. Find the DOM element where React will be mounted
const container = document.getElementById('comment');

// Check if the container element exists on the page
if (container) {
    // 2. Get the necessary data (like the post slug) from the DOM element itself.
    // This is how React communicates with your server-rendered template.
    // We assume your HTML template renders a 'data-slug' attribute on the #comment element.
    const postID = container.getAttribute('data-id');
    const postSlug = container.getAttribute('data-slug');
    const currentUserName = container.getAttribute('data-current-user');
    const comments = container.getAttribute('data-comments');
    const likes = container.getAttribute('data-like');

    // 3. Create a root with the modern React 18+ API
    const root = createRoot(container);

    // 4. Render the CommentSection component into the root
    root.render(
        // Use React.StrictMode for development checks
        <React.StrictMode>
            <LikeButton initialPostID={postID} initialLikes={likes}></LikeButton>
            <CommentSection
                initialPostID={postID}
                initialPostSlug={postSlug}
                currentUserName={currentUserName}
                initialComments={comments}
            />
        </React.StrictMode>
    );
}

const heroCropperContainer = document.getElementById('hero-cropper');
if (heroCropperContainer) {
    const root = createRoot(heroCropperContainer);
    const hiddenInput = document.getElementById('thumbnail_position');

    const handleCropChange = (pos) => {
        if (hiddenInput) {
            // Format: "center 50%"
            hiddenInput.value = `center ${pos}%`;
        }
    };

    const initialImage = heroCropperContainer.getAttribute('data-initial-image');
    const initialPosition = heroCropperContainer.getAttribute('data-initial-position');

    root.render(
        <React.StrictMode>
            <HeroImageCropper
                initialImage={initialImage}
                initialPosition={initialPosition}
                onCropChange={handleCropChange}
            />
        </React.StrictMode>
    );
}

// Mount VideoStreamer component on the stream page
const videoStreamerContainer = document.getElementById('video-streamer');
if (videoStreamerContainer) {
    const root = createRoot(videoStreamerContainer);
    const wsUrl = videoStreamerContainer.getAttribute('data-ws-url') || '/stream/ws';

    root.render(
        <React.StrictMode>
            <VideoStreamer wsUrl={wsUrl} />
        </React.StrictMode>
    );
}

// Mount RoomsList component on the rooms page
const roomsListContainer = document.getElementById('rooms-list');
if (roomsListContainer) {
    const root = createRoot(roomsListContainer);
    const isLoggedIn = roomsListContainer.getAttribute('data-is-logged-in') === 'true';

    root.render(
        <React.StrictMode>
            <RoomsList isLoggedIn={isLoggedIn} />
        </React.StrictMode>
    );
}

// Mount VideoViewer component on the watch page
const videoViewerContainer = document.getElementById('video-viewer');
if (videoViewerContainer) {
    const root = createRoot(videoViewerContainer);
    const wsUrl = videoViewerContainer.getAttribute('data-ws-url');
    const streamer = videoViewerContainer.getAttribute('data-streamer');
    const viewerCount = parseInt(videoViewerContainer.getAttribute('data-viewer-count') || '0', 10);

    root.render(
        <React.StrictMode>
            <VideoViewer
                wsUrl={wsUrl}
                streamer={streamer}
                initialViewerCount={viewerCount}
            />
        </React.StrictMode>
    );
}
