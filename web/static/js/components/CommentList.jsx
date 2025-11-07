// web/static/js/components/CommentList.jsx
import React from 'react';
import Comment from './Comment';

function CommentList({ comments, currentUserName }) { // Receive the prop here
    if (!comments || comments.length === 0) {
        return <div className="Comment-EmptyState">No comments yet. Be the first to post!</div>;
    }

    const commentsJSON = JSON.parse(comments);

    return (
        <div className="CommentList-Container">
            <h2 className="text-2xl font-bold mb-4 border-b pb-2 text-gray-700">
                Comments ({commentsJSON.length})
            </h2>

            {/* --- FIX: Map over the array and render the Comment component --- */}
            <section>
                {commentsJSON.map(comment => (
                    <Comment
                        key={comment.id} // IMPORTANT: Every mapped element needs a unique key
                        comment={comment}
                        currentUserName={currentUserName}
                    />
                ))}
            </section>
        </div>
    );
}

export default CommentList;