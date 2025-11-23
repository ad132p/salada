import React, { useState, useEffect, useRef } from 'react';

const HeroImageCropper = ({ initialImage, initialPosition, onCropChange }) => {
    const [imageSrc, setImageSrc] = useState(initialImage);

    // Parse initial position (e.g., "center 50%") to get the number
    let startPosVal = 50;
    if (initialPosition) {
        const match = initialPosition.match(/(\d+(\.\d+)?)/);
        if (match) {
            startPosVal = parseFloat(match[0]);
        }
    }

    const [position, setPosition] = useState(startPosVal); // 0 to 100
    const [isDragging, setIsDragging] = useState(false);
    const startY = useRef(0);
    const startPos = useRef(50);

    useEffect(() => {
        // Listen for image changes from EasyMDE (based on content order)
        const handleImageChange = (e) => {
            // e.detail.filepath can be null if no image exists
            setImageSrc(e.detail.filepath);
        };

        window.addEventListener('hero-image-change', handleImageChange);
        return () => window.removeEventListener('hero-image-change', handleImageChange);
    }, []);

    const handleMouseDown = (e) => {
        setIsDragging(true);
        startY.current = e.clientY;
        startPos.current = position;
    };

    const handleMouseMove = (e) => {
        if (!isDragging) return;
        const deltaY = e.clientY - startY.current;
        // Sensitivity factor.
        const sensitivity = 0.5;
        let newPos = startPos.current - (deltaY * sensitivity);
        newPos = Math.max(0, Math.min(100, newPos));
        setPosition(newPos);
        onCropChange(newPos);
    };

    const handleMouseUp = () => {
        setIsDragging(false);
    };

    useEffect(() => {
        if (isDragging) {
            window.addEventListener('mousemove', handleMouseMove);
            window.addEventListener('mouseup', handleMouseUp);
        } else {
            window.removeEventListener('mousemove', handleMouseMove);
            window.removeEventListener('mouseup', handleMouseUp);
        }
        return () => {
            window.removeEventListener('mousemove', handleMouseMove);
            window.removeEventListener('mouseup', handleMouseUp);
        };
    }, [isDragging]);

    if (!imageSrc) return null;

    return (
        <div className="mb-6 border border-gray-200 rounded-lg p-4 bg-gray-50">
            <label className="block text-sm font-medium text-gray-700 mb-2">
                Hero Image Preview (Drag to adjust vertical position)
            </label>
            <div
                className="w-full h-[20vh] min-h-[200px] relative overflow-hidden rounded-lg cursor-ns-resize border border-gray-300 shadow-sm group bg-gray-200"
                onMouseDown={handleMouseDown}
            >
                <img
                    src={imageSrc}
                    alt="Hero Preview"
                    className="w-full h-full object-cover pointer-events-none select-none"
                    style={{ objectPosition: `center ${position}%` }}
                />
                <div className="absolute inset-0 bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white font-medium pointer-events-none">
                    Drag to Reposition
                </div>
            </div>
            <div className="mt-2 flex items-center space-x-4">
                <input
                    type="range"
                    min="0"
                    max="100"
                    value={position}
                    onChange={(e) => {
                        const val = Number(e.target.value);
                        setPosition(val);
                        onCropChange(val);
                    }}
                    className="flex-1 h-2 bg-gray-300 rounded-lg appearance-none cursor-pointer"
                />
                <span className="text-xs text-gray-600 font-mono min-w-[3rem]">
                    {Math.round(position)}%
                </span>
            </div>
        </div>
    );
};

export default HeroImageCropper;
