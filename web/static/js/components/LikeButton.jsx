import React, { useState } from 'react';

// You would typically use an icon library like Heroicons or Font Awesome
// For simplicity, we'll use inline SVG heart icons.
const HeartIcon = ({ filled, className }) => (
    <svg
        className={className}
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill={filled ? "currentColor" : "none"}
        stroke={filled ? "currentColor" : "currentColor"}
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
    >
        <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
    </svg>
);

/**
 * A like button component that manages state and posts a like/unlike action.
 * @param {object} props
 * @param {string} props.postID - Unique identifier for the blog post.
 * @param {number} props.initialLikes - The starting number of likes.
 * @param {boolean} props.hasUserLiked - True if the current user has already liked the post.
 */
function LikeButton({ initialPostID, initialLikes = 0, hasUserLiked = false }) {
    const [likes, setLikes] = useState(initialLikes);
    const [isLiked, setIsLiked] = useState(hasUserLiked);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState(null);

    const handleLikeToggle = async () => {
        if (isLoading) return;

        // Capture state BEFORE optimistic update logic
        const previousIsLiked = isLiked;
        const previousLikes = likes;

        setIsLoading(true);
        setError(null);

        const action = isLiked ? 'unlike' : 'like';
        const endpoint = '/blog/like';

        // APPLY OPTIMISTIC UPDATE *NOW*
        setLikes(action === 'like' ? likes + 1 : likes - 1);
        setIsLiked(!isLiked);

        try {
            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ post_id: initialPostID, action: action }),
            });

            if (!response.ok) {
                // Throwing the error moves execution to the catch block
                throw new Error(`Failed to ${action} post. Status: ${response.status}`);
            }

            // Success is implicitly handled: state stays at the optimistic value.

        } catch (err) {
            // --- FAILURE: REVERT STATE ---
            console.error(err.message);
            setError("Could not complete action. Please try again.");

            // Revert to the state captured before the optimistic update
            setLikes(previousLikes);
            setIsLiked(previousIsLiked);

        } finally {
            setIsLoading(false);
        }
    };
    // Conditional classes for button style
    const baseClasses = 'flex items-center space-x-2 py-2 px-4 rounded-full transition-colors duration-200 focus:outline-none focus:ring-4';
    const activeClasses = 'bg-red-500 hover:bg-red-600 focus:ring-red-300';
    const inactiveClasses = 'bg-gray-100 text-gray-600 hover:bg-gray-200 focus:ring-gray-300';
    const loadingClasses = 'bg-gray-300 text-gray-500 cursor-not-allowed';

    const buttonClass = isLoading
        ? `${baseClasses} ${loadingClasses}`
        : `${baseClasses} ${isLiked ? activeClasses : inactiveClasses}`;

    return (
        <div className="flex flex-col items-start">
            <button
                className={buttonClass}
                onClick={handleLikeToggle}
                disabled={isLoading}
                aria-label={isLiked ? "Unlike post" : "Like post"}
            >
                {isLoading ? (
                    // Simple loading spinner using Tailwind
                    <span className="w-5 h-5 border-2 border-current border-t-transparent rounded-full animate-spin"></span>
                ) : (
                    <HeartIcon
                        filled={isLiked}
                        className="w-5 h-5"
                    />
                )}

                <span className="font-semibold select-none">
                    {likes} {likes === 1 ? 'Like' : 'Likes'}
                </span>
            </button>
            {error && (
                <p className="mt-1 text-sm text-red-600">
                    {error}
                </p>
            )}
        </div>
    );
};




export default LikeButton;