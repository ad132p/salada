import React, { useState, useEffect, useRef } from 'react';
import EasyMDE from 'easymde';
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
    const [htmlContent, setHtmlContent] = useState('');
    const [viewMode, setViewMode] = useState('edit'); // 'edit', 'html'
    const [copied, setCopied] = useState(false);

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

        const easyMDE = new EasyMDE({
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
            showIcons: ["code", "table"],
            renderingConfig: {
                singleLineBreaks: false,
            },
            toolbar: [
                "bold", "italic", "heading", "|",
                "unordered-list", "ordered-list", "code", "quote", "|",
                "link", "upload-image", "|",
                "preview", "side-by-side", "fullscreen", "|",
                "guide"
            ],
        });

        // Set initial HTML content
        setHtmlContent(easyMDE.markdown(easyMDE.value()));

        // Configure CodeMirror to show zero-width spaces
        easyMDE.codemirror.setOption("specialChars", /[\u200b]/g);
        easyMDE.codemirror.setOption("specialCharPlaceholder", (char) => {
            const el = document.createElement("span");
            el.style.color = "red";
            el.textContent = "ZWSP";
            el.title = "Zero Width Space";
            el.className = "zwsp-placeholder";
            return el;
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

            // Update HTML content in real-time
            setHtmlContent(easyMDE.markdown(content));
        });

        easyMDERef.current = easyMDE;

        return () => {
            if (easyMDERef.current) {
                easyMDERef.current.toTextArea();
                easyMDERef.current = null;
            }
        };
    }, []);

    useEffect(() => {
        if (easyMDERef.current && viewMode === 'edit') {
            setTimeout(() => {
                easyMDERef.current.codemirror.refresh();
            }, 50);
        }
    }, [viewMode]);

    const handleCopyHTML = () => {
        navigator.clipboard.writeText(htmlContent);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setIsSubmitting(true);
        setMessage(isEditing ? 'Updating post...' : 'Publishing post...');

        const content = easyMDERef.current ? easyMDERef.current.value() : '';
        const cleanContent = content.replace(/\u200B/g, '');
        
        const submissionData = {
            id: initialPost?.id,
            title: title,
            content: cleanContent,
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

                    <div className="space-y-3">
                        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 bg-gray-50 p-3 rounded-lg border border-gray-200">
                            <span className="text-sm font-semibold text-gray-700">Post Content</span>
                            
                            <div className="flex flex-wrap items-center gap-3">
                                {/* View Mode Controls */}
                                <div className="inline-flex rounded-md shadow-sm" role="group">
                                    <button
                                        type="button"
                                        onClick={() => setViewMode('edit')}
                                        className={`px-4 py-1.5 text-xs font-medium rounded-l-lg border transition-all duration-150 cursor-pointer ${
                                            viewMode === 'edit'
                                                ? 'bg-blue-600 border-blue-600 text-white shadow-inner'
                                                : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
                                        }`}
                                    >
                                        Markdown Editor
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => setViewMode('html')}
                                        className={`px-4 py-1.5 text-xs font-medium rounded-r-lg border-t border-b border-r transition-all duration-150 cursor-pointer ${
                                            viewMode === 'html'
                                                ? 'bg-blue-600 border-blue-600 text-white shadow-inner'
                                                : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
                                        }`}
                                    >
                                        HTML Equivalent
                                    </button>
                                </div>

                                {/* Always Accessible Easy Copy HTML Button */}
                                <button
                                    type="button"
                                    onClick={handleCopyHTML}
                                    className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-gray-300 bg-white hover:bg-gray-50 text-gray-700 text-xs font-medium transition-all duration-200 active:scale-95 cursor-pointer shadow-sm"
                                    title="Copy HTML output to clipboard"
                                >
                                    {copied ? (
                                        <>
                                            <svg className="w-3.5 h-3.5 text-green-500 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M5 13l4 4L19 7" />
                                            </svg>
                                            <span className="text-green-600 font-semibold">HTML Copied!</span>
                                        </>
                                    ) : (
                                        <>
                                            <svg className="w-3.5 h-3.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                                            </svg>
                                            <span>Copy HTML</span>
                                        </>
                                    )}
                                </button>
                            </div>
                        </div>

                        {/* Editor and HTML Panes Container */}
                        <div className="grid grid-cols-1">
                            {/* Editor Column */}
                            <div className={viewMode === 'html' ? 'hidden' : 'block'}>
                                <label htmlFor="content" className="sr-only">Write your post content here:</label>
                                <textarea
                                    id="content"
                                    ref={textareaRef}
                                    className="block w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm min-h-[300px]"
                                ></textarea>
                            </div>

                            {/* HTML Equivalent Column */}
                            <div className={`flex flex-col border border-gray-300 rounded-md bg-gray-900 text-gray-100 overflow-hidden ${viewMode === 'edit' ? 'hidden' : 'block'}`}>
                                <div className="flex items-center justify-between px-4 py-2 bg-gray-800 border-b border-gray-700">
                                    <span className="text-xs font-semibold uppercase tracking-wider text-gray-400 font-mono">HTML Output (Realtime)</span>
                                </div>
                                <div className="p-4 overflow-auto font-mono text-sm leading-relaxed flex-1 min-h-[300px] max-h-[500px]" style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                                    <code dangerouslySetInnerHTML={{ __html: highlightHTML(htmlContent) || '<span class="text-gray-500 italic">&lt;!-- HTML equivalent will appear here in real time... --&gt;</span>' }} />
                                </div>
                            </div>
                        </div>
                    </div>
                </fieldset>

                <div className="flex space-x-4 justify-end">
                    <button
                        type="submit"
                        disabled={isSubmitting}
                        className={`bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors duration-200 cursor-pointer ${isSubmitting ? 'opacity-50 cursor-not-allowed' : ''}`}
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
                                setHtmlContent('');
                                setThumbnailPosition('center 50%');
                            }}
                            className="bg-gray-400 hover:bg-gray-500 text-white font-medium py-2 px-6 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 transition-colors duration-200 cursor-pointer"
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
                .html-tag { color: #60a5fa; font-weight: 600; }
                .html-attr { color: #f59e0b; }
                .html-val { color: #34d399; }
                .html-comment { color: #9ca3af; font-style: italic; }
            `}</style>
        </div>
    );
};

const highlightHTML = (html) => {
    if (!html) return '';
    let escaped = html
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');

    escaped = escaped.replace(/&lt;(\/?[a-zA-Z0-9:-]+)(.*?)&gt;/g, (match, tagName, attrs) => {
        const highlightedAttrs = attrs.replace(/([a-zA-Z0-9:-]+)=(['"].*?['"])/g, ' <span class="html-attr">$1</span>=<span class="html-val">$2</span>');
        return `<span class="html-tag">&lt;${tagName}</span>${highlightedAttrs}<span class="html-tag">&gt;</span>`;
    });
    
    escaped = escaped.replace(/&lt;!--(.*?)--&gt;/g, '<span class="html-comment">&lt;!--$1--&gt;</span>');
    return escaped;
};

export default BlogForm;
