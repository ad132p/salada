import React, { useState, useEffect, useRef } from 'react';
import HeroImageCropper from './HeroImageCropper';

const BlogForm = ({ initialPost, categories, isEditing, username }) => {
    const [title, setTitle] = useState(initialPost?.title || '');
    const [category, setCategory] = useState(initialPost?.category || '');
    // Handle tags as a string for the input
    const initialTags = Array.isArray(initialPost?.tags) ? initialPost.tags.join(',') : (initialPost?.tags || '');
    const [tags, setTags] = useState(initialTags);
    const [thumbnailPosition, setThumbnailPosition] = useState(initialPost?.thumbnail_position || 'center 50%');
    const [uploadedImageUUIDs, setUploadedImageUUIDs] = useState([]);
    const [message, setMessage] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);

    const easyMDERef = useRef(null);
    const textareaRef = useRef(null);

    useEffect(() => {
        if (!textareaRef.current) return;

        const imageMonitorStatus = {
            className: "image-uuid-monitor",
            defaultValue: "Image Uploads: None"
        };

        const updateImageMonitor = (newUUID) => {
            setUploadedImageUUIDs(prev => {
                const updated = [...prev, newUUID];
                const monitorElement = document.querySelector('.' + imageMonitorStatus.className);
                if (monitorElement) {
                    monitorElement.textContent = `Image Uploads: ${updated.join(', ')}`;
                }
                return updated;
            });
        };

        const customImageUpload = async (file, onSuccess, onError) => {
            const formData = new FormData();
            formData.append("image", file);

            try {
                const response = await fetch('/blog/uploads/', {
                    method: 'POST',
                    body: formData,
                });

                if (!response.ok) {
                    const errorText = await response.text();
                    onError("Upload Failed: " + errorText);
                    return;
                }

                const payload = await response.json();
                updateImageMonitor(payload.data.image_id);
                onSuccess(payload.data.filepath);
            } catch (error) {
                onError("Network Error: " + error.message);
            }
        };

        const easyMDE = new window.EasyMDE({
            element: textareaRef.current,
            autofocus: true,
            imageMaxSize: 1024 * 1024 * 8,
            imageAccept: "image/png, image/jpeg, image/jpg, image/webp",
            status: ["autosave", "lines", "words", "cursors", imageMonitorStatus],
            previewImagesInEditor: true,
            previewClass: ["h-200", "w-96", "object-scale-down"],
            imageUploadFunction: customImageUpload,
            imagePathAbsolute: true,
            placeholder: "Start writing your blog post in Markdown...",
            spellChecker: true,
            initialValue: initialPost?.content || '',
            toolbar: [
                "bold", "italic", "heading", "|",
                "unordered-list", "ordered-list", "code", "quote", "|",
                "link", "upload-image", "|",
                "preview", "side-by-side", "fullscreen", "|",
                "guide"
            ],
        });

        easyMDE.codemirror.on("change", () => {
            const content = easyMDE.value();
            const match = content.match(/!\[.*?\]\((.*?)\)/);
            let firstImage = match ? match[1] : null;

            if (firstImage) {
                const titleIndex = firstImage.indexOf(' "');
                if (titleIndex !== -1) {
                    firstImage = firstImage.substring(0, titleIndex);
                }
            }
            window.dispatchEvent(new CustomEvent('hero-image-change', { detail: { filepath: firstImage } }));
        });

        easyMDERef.current = easyMDE;

        return () => {
            if (easyMDERef.current) {
                easyMDERef.current.toTextArea();
                easyMDERef.current = null;
            }
        };
    }, []);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setIsSubmitting(true);
        setMessage(isEditing ? 'Updating post...' : 'Publishing post...');

        const content = easyMDERef.current ? easyMDERef.current.value() : '';
        
        const submissionData = {
            id: initialPost?.id,
            title: title,
            content: content,
            tags: tags,
            category: category,
            image_ids: uploadedImageUUIDs,
            thumbnail_position: thumbnailPosition,
            author: username // Only used for new posts in server side usually, but kept for parity
        };

        const url = isEditing ? `/blog/${initialPost.id}` : '/blog/';
        const method = isEditing ? 'PUT' : 'POST';

        try {
            const response = await fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(submissionData),
            });

            if (response.ok) {
                setMessage(isEditing ? 'Post updated successfully!' : 'Post published successfully!');
                setTimeout(() => {
                    window.location.replace(isEditing ? `/blog/${initialPost.slug || initialPost.id}` : "/thankyou");
                }, 1500);
            } else {
                const errorText = await response.text();
                setMessage(`Error: ${errorText}`);
            }
        } catch (error) {
            setMessage(`Network error: ${error.message}`);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="bg-white rounded-lg shadow-xl p-8 lg:p-12">
            <h1 className="text-3xl lg:text-4xl font-extrabold text-gray-900 leading-tight mb-6">
                {isEditing ? `Editing: ${initialPost?.title}` : 'Create New Post'}
            </h1>
            <form className="space-y-8" onSubmit={handleSubmit}>
                <fieldset className="border border-gray-200 rounded-lg p-6 space-y-4">
                    <legend className="text-lg font-semibold text-gray-800 px-2">Post Details</legend>
                    
                    <div>
                        <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-1">Title:</label>
                        <input
                            type="text"
                            id="title"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            required
                            placeholder="Enter your post title here"
                            className="block w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                        />
                    </div>

                    <div>
                        <label htmlFor="category" className="block text-sm font-medium text-gray-700 mb-1">Category (Optional):</label>
                        <select
                            id="category"
                            value={category}
                            onChange={(e) => setCategory(e.target.value)}
                            className="block w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                        >
                            <option value="">Please choose an option</option>
                            {categories.map(cat => (
                                <option key={cat} value={cat}>{cat}</option>
                            ))}
                        </select>
                    </div>

                    <div>
                        <label htmlFor="tags" className="block text-sm font-medium text-gray-700 mb-1">Tags (Comma-separated, Optional):</label>
                        <input
                            type="text"
                            id="tags"
                            value={tags}
                            onChange={(e) => setTags(e.target.value)}
                            placeholder="e.g., html, css, javascript"
                            className="block w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                        />
                    </div>
                </fieldset>

                <fieldset className="border border-gray-200 rounded-lg p-6 space-y-4">
                    <legend className="text-lg font-semibold text-gray-800 px-2">Post Content</legend>

                    <HeroImageCropper
                        initialImage={initialPost?.thumbnail}
                        initialPosition={thumbnailPosition}
                        onCropChange={(pos) => setThumbnailPosition(`center ${pos}%`)}
                    />

                    <div>
                        <label htmlFor="content" className="block text-sm font-medium text-gray-700 mb-1">Write your post content here:</label>
                        <textarea
                            id="content"
                            ref={textareaRef}
                            className="block w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm min-h-[300px]"
                        ></textarea>
                    </div>
                </fieldset>

                <div className="flex space-x-4 justify-end">
                    <button
                        type="submit"
                        disabled={isSubmitting}
                        className={`bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors duration-200 ${isSubmitting ? 'opacity-50 cursor-not-allowed' : ''}`}
                    >
                        {isEditing ? 'Update Post' : 'Publish Post'}
                    </button>
                    {!isEditing && (
                        <button
                            type="reset"
                            onClick={() => {
                                setTitle('');
                                setCategory('');
                                setTags('');
                                if (easyMDERef.current) easyMDERef.current.value('');
                                setThumbnailPosition('center 50%');
                            }}
                            className="bg-gray-400 hover:bg-gray-500 text-white font-medium py-2 px-6 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 transition-colors duration-200"
                        >
                            Clear Form
                        </button>
                    )}
                </div>
                {message && (
                    <p className={`mt-4 text-center text-sm italic ${message.includes('Error') ? 'text-red-600' : 'text-gray-600'}`}>
                        {message}
                    </p>
                )}
            </form>
            <style>{`
                .CodeMirror span[role="presentation"][data-img-src] {
                    width: 400px !important;
                    display: inline-block !important;
                    background-size: contain !important;
                    background-repeat: no-repeat !important;
                    background-position: center !important;
                    border: 1px solid #ccc;
                }
            `}</style>
        </div>
    );
};

export default BlogForm;
