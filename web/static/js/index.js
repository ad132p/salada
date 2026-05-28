import React from 'react';
import { Collapse } from 'flowbite';
import { createRoot } from 'react-dom/client';
import CommentSection from './components/CommentSection';
import LikeButton from './components/LikeButton';
import HeroImageCropper from './components/HeroImageCropper';

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
