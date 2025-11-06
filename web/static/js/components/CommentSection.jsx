// web/static/js/components/CommentSection.jsx
import React, { useState, useEffect } from 'react';
import CommentList from './CommentList'; 
import CommentForm from './CommentForm'; 

/**
 * Main container component for the entire comments section.
 * @param {object} props
 * @param {string} props.initialPostSlug - The unique slug of the blog post, passed from the server template.
 */
function CommentSection({ initialPostSlug, currentUserName }) {
    const [comments, setComments] = useState([]);
    const [isLoading, setIsLoading] = useState(true);
    const postSlug = initialPostSlug;
    
    // Function to fetch comments from the API
    const fetchComments = async () => {
        setIsLoading(true);
        // Replace with your actual API endpoint: `/api/posts/${postSlug}/comments`
        // const response = await fetch(`/api/comments?slug=${postSlug}`);
        // const data = await response.json();
        
        // --- Placeholder Data ---
        const data = [
            { id: 'c3', authorName: 'Carla', authorAvatarUrl: '...', text: 'Nice component!', timestamp: new Date().toISOString() },
            // ... more comments
        ];
        
        setComments(data);
        setIsLoading(false);
    };

    // Fetch comments when the component mounts or postSlug changes
    useEffect(() => {
        if (postSlug) {
            fetchComments();
        }
    }, [postSlug]); 

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
            <CommentList 
                comments={comments} 
                currentUserName={currentUserName} 
            />
            <CommentForm postSlug={postSlug} onSubmit={handleCommentSubmitted} />
        </div>
    );
}

export default CommentSection;