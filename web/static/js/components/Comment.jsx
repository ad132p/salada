// web/static/js/components/Comment.jsx
import React from 'react';

/**
 * Renders a single, styled blog comment.
 * @param {object} props
 * @param {object} props.comment - The single comment object.
 * @param {string} props.currentUserName - Name of the current authenticated user (optional for highlighting).
 */
function Comment({ comment, currentUserName }) {
    // Determine the author's name, falling back to a generic label if needed
    const authorName = currentUserName || 'Anon';
    
    // Format the date for better readability
    const formattedDate = new Date(comment.created_at).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
    
    // Highlight the current user's comment (assuming authorName is available)
    const isCurrentUser = currentUserName && authorName === currentUserName;
    const baseClasses = "p-4 my-4 rounded-lg shadow-md transition duration-150 ease-in-out";
    const commentClasses = isCurrentUser 
        ? `${baseClasses} bg-blue-50 border-2 border-blue-400`
        : `${baseClasses} bg-white border border-gray-200`;

    return (
        <article className={commentClasses}>
            {/* Header: Author and Timestamp */}
            <header className="flex justify-between items-start mb-2 border-b pb-2">
                <div className="flex items-center">
                    <span className="font-semibold text-gray-800 mr-2">
                        {comment.author_name}
                    </span>
                    {isCurrentUser && (
                        <span className="text-xs bg-blue-500 text-white px-2 py-0.5 rounded-full">
                            You
                        </span>
                    )}
                </div>
                <time 
                    dateTime={comment.created_at} 
                    className="text-sm text-gray-500"
                >
                    {formattedDate}
                </time>
            </header>

            {/* Body: Comment Content */}
            <p className="text-gray-700 whitespace-pre-wrap">
                {comment.content}
            </p>
        </article>
    );
}

export default Comment;