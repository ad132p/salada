// web/static/js/components/CommentSection.jsx
import React, { useState, useEffect } from 'react';
import CommentList from './CommentList'; 
import CommentForm from './CommentForm'; 

/**
 * Main container component for the entire comments section.
 * @param {object} props
 * @param {string} props.initialPostID - The unique slug of the blog post, passed from the server template.
 * @param {string} props.currentUserName - The current user's name.
 * @param {Array<object>} props.initialComments - The initial comments list, likely passed from the server template.
 */
function CommentSection({ initialPostID, initialPostSlug, currentUserName, initialComments }) {
    // 1. Initialize local state for comments using the prop's value
    const [comments, setComments] = useState(initialComments || []); 
    
    // 2. Local state to track loading
    const [isLoading, setIsLoading] = useState(false); // Set to false initially if comments are pre-loaded
    const postID = initialPostID;
    const postSlug = initialPostSlug;
    
    // Function to fetch and refresh comments from the API
    const fetchComments = async () => {
        // Only set loading if we're actually fetching (i.e., not the initial render if pre-loaded)
        setIsLoading(true);
        
        try {
            // Replace with your actual API endpoint: `/api/posts/${postID}/comments`
            const response = await fetch(`/blog/comment/${postSlug}`);
            const data = await response.json();

            // 3. Update the local state
            setComments(data.comments);
        } catch (error) {
            console.error("Failed to fetch comments:", error);
        } finally {
            setIsLoading(false);
        }
    };

    // Handler to refresh comments after a successful submission
    const handleCommentSubmitted = () => {
        // Re-run the fetch to include the new comment
        fetchComments(); 
    };

    if (isLoading) {
        return <div className="p-4 text-center text-gray-500">Loading comments...</div>;
    }

    return (
        <div className="Comment-ApplicationRoot">
            <CommentForm postID={postID} currentUserName={currentUserName} onSubmit={handleCommentSubmitted} />

           <CommentList 
               comments={comments} // Pass the local state to the list
           />
        </div>
    );
}

export default CommentSection;