// web/static/js/components/CommentForm.jsx
import React, { useState } from 'react';

/**
 * A form component for submitting a new blog comment.
 *
 * @param {object} props
 * @param {string} props.postSlug - The unique slug of the blog post.
 * @param {function} props.onSubmit - Function to call on successful submission (e.g., to refresh comments).
 */
function CommentForm({ postSlug, onSubmit }) {
    // State to hold the content of the comment textarea
    const [commentText, setCommentText] = useState('');
    // State to handle loading/submission status
    const [isSubmitting, setIsSubmitting] = useState(false);
    // State to hold any submission errors
    const [error, setError] = useState(null);

    // Determines the action URL based on the postSlug prop
    const formAction = `/comments/${postSlug}`;

    /**
     * Handles the form submission logic.
     * @param {object} e - The form submission event.
     */
    const handleSubmit = async (e) => {
        e.preventDefault(); // Prevent the default browser form submission

        if (!commentText.trim()) {
            setError('Comment cannot be empty.');
            return;
        }

        setIsSubmitting(true);
        setError(null);

        // --- Simulated API Submission ---
        // In a real application, you would replace this block with your
        // actual fetch() or Axios call to the backend API.
        try {
            const response = await fetch(formAction, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    // Often include a security token (CSRF token, JWT, etc.) here
                },
                // Send the comment data to the server
                body: JSON.stringify({ comment: commentText }),
            });

            if (!response.ok) {
                // Handle non-200 responses (e.g., 400, 500)
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to submit comment.');
            }

            // Success! Clear the form and run the parent callback
            setCommentText('');
            // The onSubmit prop is typically used to tell the parent component
            // (e.g., BlogPostPage) to refresh the list of comments.
            if (onSubmit) {
                onSubmit();
            }

        } catch (err) {
            console.error('Submission Error:', err);
            setError(err.message || 'An unexpected error occurred.');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="mt-12 pt-6 border-t border-gray-200">
            <h3 className="text-2xl font-bold text-gray-800 mb-6">Leave a Comment</h3>
            
            {/* The action is handled by JavaScript, not the browser, so we use an onSubmit handler */}
            <form onSubmit={handleSubmit} className="space-y-4">
                {/* Display error message if present */}
                {error && (
                    <div className="p-3 text-sm font-medium text-red-700 bg-red-100 rounded-lg" role="alert">
                        {error}
                    </div>
                )}

                <div>
                    <label htmlFor="comment-text" className="sr-only">Your Comment</label>
                    <textarea 
                        id="comment-text" 
                        name="comment" 
                        rows="4" 
                        required
                        value={commentText}
                        onChange={(e) => setCommentText(e.target.value)}
                        // Reusing Tailwind classes from your template
                        className="w-full p-4 text-gray-700 bg-gray-50 border border-gray-300 rounded-lg focus:ring-blue-500 focus:border-blue-500 transition duration-150 ease-in-out resize-none shadow-sm"
                        placeholder="Share your thoughts on this post..."
                        disabled={isSubmitting} // Disable during submission
                    ></textarea>
                </div>
                
                <div className="flex justify-end">
                    <button 
                        type="submit"
                        disabled={isSubmitting} // Disable button while submitting
                        // Reusing Tailwind classes from your template
                        className={`px-6 py-3 text-lg font-medium text-white rounded-lg focus:ring-4 focus:ring-blue-300 focus:outline-none transition duration-150 ease-in-out shadow-md ${
                            isSubmitting 
                                ? 'bg-blue-400 cursor-not-allowed' // Style when disabled
                                : 'bg-blue-600 hover:bg-blue-700' // Style when active
                        }`}
                    >
                        {/* Dynamic text based on submission state */}
                        {isSubmitting ? 'Submitting...' : 'Submit Comment'}
                    </button>
                </div>
            </form>
        </div>
    );
}

export default CommentForm;