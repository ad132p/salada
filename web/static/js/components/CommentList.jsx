// web/static/js/components/CommentList.jsx
import React from 'react';
import Comment from './Comment';

function CommentList({ comments, currentUserName }) { // Receive the prop here
    if (!comments || comments.length === 0) {
        return <div className="Comment-EmptyState">No comments yet. Be the first to post!</div>;
    }

    return (
        <div className="CommentList-Container">
            <h2>Comments ({comments.length})</h2>
            {comments.map((comment) => (
                <Comment 
                    key={comment.id} 
                    comment={comment} 
                    // Pass the current user's name to the child
                    currentUserName={currentUserName} 
                />
            ))}
        </div>
    );
}

export default CommentList;