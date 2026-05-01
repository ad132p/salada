import { Collapse } from 'flowbite';

// Wait for the DOM to be fully loaded before initializing components
document.addEventListener('DOMContentLoaded', () => {
    try {
        // 1. Get the target element (the menu content that collapses)
        // This ID now matches the <div> with id="navbar-hamburger"
        const $targetEl = document.getElementById('navbar-hamburger');

        // 2. Get the trigger element (the hamburger button)
        const $triggerEl = document.querySelector('[data-collapse-toggle="navbar-hamburger"]');

        // Check if both elements exist
        if ($targetEl && $triggerEl) {
            const collapse = new Collapse($targetEl, $triggerEl);
            // Optional: Expose the collapse object globally for debugging
            window.mobileMenuCollapse = collapse;
        }

    } catch (e) {
        console.error('Error during Flowbite Collapse setup:', e);
    }

    // Dynamic syntax highlighting
    if (document.querySelector('pre code')) {
        import('highlight.js').then((module) => {
            module.default.highlightAll();
        }).catch(err => console.error('Failed to load highlight.js:', err));
    }

    // Copy to clipboard functionality
    document.addEventListener('click', (e) => {
        if (e.target && e.target.classList.contains('copy-button')) {
            const button = e.target;
            const container = button.closest('.relative.group');
            const code = container ? container.querySelector('code') : null;

            if (code) {
                const text = code.innerText.replace(/\u200B/g, '');
                navigator.clipboard.writeText(text).then(() => {
                    const originalText = button.textContent;
                    button.textContent = 'Copied!';
                    button.classList.add('bg-green-600', 'text-white');
                    button.classList.remove('bg-gray-700', 'text-gray-300');

                    setTimeout(() => {
                        button.textContent = originalText;
                        button.classList.remove('bg-green-600', 'text-white');
                        button.classList.add('bg-gray-700', 'text-gray-300');
                    }, 2000);
                }).catch(err => {
                    console.error('Could not copy text: ', err);
                });
            }
        }
    });
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
    const postID = container.getAttribute('data-id');
    const postSlug = container.getAttribute('data-slug');
    const currentUserName = container.getAttribute('data-current-user');
    const comments = container.getAttribute('data-comments');
    const likes = container.getAttribute('data-like');

    // 3. Create a root with the modern React 18+ API
    const root = createRoot(container);

    // 4. Render the CommentSection component into the root
    root.render(
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
    const iceServers = JSON.parse(videoStreamerContainer.getAttribute('data-ice-servers') || '[]');

    root.render(
        <React.StrictMode>
            <VideoStreamer wsUrl={wsUrl} iceServers={iceServers} />
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
    const iceServers = JSON.parse(videoViewerContainer.getAttribute('data-ice-servers') || '[]');

    root.render(
        <React.StrictMode>
            <VideoViewer
                wsUrl={wsUrl}
                streamer={streamer}
                initialViewerCount={viewerCount}
                iceServers={iceServers}
            />
        </React.StrictMode>
    );
}

// Mount BlogForm component with dynamic import
const blogFormContainer = document.getElementById('blog-form-container');
if (blogFormContainer) {
    const isEditing = blogFormContainer.getAttribute('data-is-editing') === 'true';
    const categories = JSON.parse(blogFormContainer.getAttribute('data-categories') || '[]');
    const initialPost = JSON.parse(blogFormContainer.getAttribute('data-post') || 'null');
    const username = blogFormContainer.getAttribute('data-username');

    import('./components/BlogForm').then(({ default: BlogForm }) => {
        const root = createRoot(blogFormContainer);
        root.render(
            <React.StrictMode>
                <BlogForm
                    isEditing={isEditing}
                    categories={categories}
                    initialPost={initialPost}
                    username={username}
                />
            </React.StrictMode>
        );
    }).catch(err => console.error('Failed to load BlogForm:', err));
}
