// web/static/js/components/VideoStreamer.jsx
import React, { useState, useEffect, useRef, useCallback } from 'react';

/**
 * VideoStreamer component for camera selection and streaming control.
 * Uses WebRTC for peer-to-peer video streaming with WebSocket signaling.
 *
 * @param {object} props
 * @param {string} props.wsUrl - WebSocket URL for signaling
 */
function VideoStreamer({ wsUrl }) {
    const [devices, setDevices] = useState([]);
    const [selectedDeviceId, setSelectedDeviceId] = useState('');
    const [isStreaming, setIsStreaming] = useState(false);
    const [isConnecting, setIsConnecting] = useState(false);
    const [error, setError] = useState(null);
    const [connectedClients, setConnectedClients] = useState(0);
    const [isSecureContext, setIsSecureContext] = useState(true);

    const videoRef = useRef(null);
    const streamRef = useRef(null);
    const wsRef = useRef(null);
    const peerConnectionRef = useRef(null);

    // Check for secure context and mediaDevices support
    useEffect(() => {
        // Check if we're in a secure context
        if (!window.isSecureContext) {
            setIsSecureContext(false);
            setError(
                'Camera access requires a secure context (HTTPS or localhost). ' +
                'Please access this page via https:// or http://localhost:8080'
            );
            return;
        }

        // Check for mediaDevices support
        if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
            setError(
                'Your browser does not support camera access. ' +
                'Please use a modern browser like Chrome, Firefox, or Safari.'
            );
            return;
        }
    }, []);

    // Fetch available video input devices
    useEffect(() => {
        if (!isSecureContext || !navigator.mediaDevices) return;

        const getDevices = async () => {
            try {
                // Request permission first to get device labels
                await navigator.mediaDevices.getUserMedia({ video: true });

                const deviceList = await navigator.mediaDevices.enumerateDevices();
                const videoDevices = deviceList.filter(device => device.kind === 'videoinput');

                setDevices(videoDevices);

                // Select first device by default
                if (videoDevices.length > 0 && !selectedDeviceId) {
                    setSelectedDeviceId(videoDevices[0].deviceId);
                }
            } catch (err) {
                console.error('Error getting devices:', err);
                if (err.name === 'NotAllowedError' || err.name === 'PermissionDeniedError') {
                    setError('Camera access denied. Please allow camera access in your browser settings and refresh the page.');
                } else if (err.name === 'NotFoundError') {
                    setError('No camera found. Please connect a camera and refresh the page.');
                } else {
                    setError('Unable to access camera: ' + err.message);
                }
            }
        };

        getDevices();

        // Listen for device changes (e.g., camera plugged in/out)
        const handleDeviceChange = () => getDevices();
        navigator.mediaDevices.addEventListener('devicechange', handleDeviceChange);
        return () => navigator.mediaDevices.removeEventListener('devicechange', handleDeviceChange);
    }, [isSecureContext, selectedDeviceId]);

    // Fetch connected client count
    const fetchClientCount = useCallback(async () => {
        try {
            const response = await fetch('/stream/clients');
            const data = await response.json();
            setConnectedClients(data.connected_clients);
        } catch (err) {
            console.error('Error fetching client count:', err);
        }
    }, []);

    // Poll for client count when streaming
    useEffect(() => {
        if (!isStreaming) return;

        fetchClientCount();
        const interval = setInterval(fetchClientCount, 5000);
        return () => clearInterval(interval);
    }, [isStreaming, fetchClientCount]);

    // Initialize WebSocket connection for signaling
    const initializeWebSocket = useCallback(() => {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const ws = new WebSocket(`${protocol}//${window.location.host}${wsUrl}`);
        wsRef.current = ws;

        ws.onopen = () => {
            console.log('WebSocket connected');
        };

        ws.onmessage = async (event) => {
            const message = JSON.parse(event.data);
            await handleSignalingMessage(message);
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            setError('Connection error. Please try again.');
        };

        ws.onclose = () => {
            console.log('WebSocket disconnected');
        };

        return ws;
    }, [wsUrl]);

    // Handle incoming signaling messages
    const handleSignalingMessage = async (message) => {
        const pc = peerConnectionRef.current;
        if (!pc) return;

        try {
            switch (message.type) {
                case 'answer':
                    await pc.setRemoteDescription(new RTCSessionDescription(message));
                    break;

                case 'ice-candidate':
                    await pc.addIceCandidate(new RTCIceCandidate(message.candidate));
                    break;

                default:
                    break;
            }
        } catch (err) {
            console.error('Error handling signaling message:', err);
        }
    };

    // Send signaling message via WebSocket
    const sendSignalingMessage = (message) => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify(message));
        }
    };

    // Start streaming
    const startStreaming = async () => {
        if (!selectedDeviceId) {
            setError('Please select a video device first.');
            return;
        }

        if (!navigator.mediaDevices) {
            setError('Camera access not available in this context.');
            return;
        }

        setIsConnecting(true);
        setError(null);

        try {
            // Get video stream from selected device
            const constraints = {
                video: {
                    deviceId: { exact: selectedDeviceId },
                    width: { ideal: 1280 },
                    height: { ideal: 720 }
                },
                audio: false
            };

            const stream = await navigator.mediaDevices.getUserMedia(constraints);
            streamRef.current = stream;

            // Display stream in video element
            if (videoRef.current) {
                videoRef.current.srcObject = stream;
            }

            // Create peer connection
            const pc = new RTCPeerConnection({
                iceServers: [
                    { urls: 'stun:stun.l.google.com:19302' },
                    { urls: 'stun:stun1.l.google.com:19302' }
                ]
            });
            peerConnectionRef.current = pc;

            // Add stream tracks to peer connection
            stream.getTracks().forEach(track => {
                pc.addTrack(track, stream);
            });

            // Handle ICE candidates
            pc.onicecandidate = (event) => {
                if (event.candidate) {
                    sendSignalingMessage({
                        type: 'ice-candidate',
                        candidate: event.candidate
                    });
                }
            };

            // Initialize WebSocket for signaling
            initializeWebSocket();

            // Create and send offer
            const offer = await pc.createOffer();
            await pc.setLocalDescription(offer);

            // Wait a moment for ICE gathering, then send the offer
            setTimeout(() => {
                sendSignalingMessage({
                    type: 'offer',
                    sdp: pc.localDescription.sdp
                });
            }, 1000);

            setIsStreaming(true);
        } catch (err) {
            console.error('Error starting stream:', err);
            setError('Failed to start streaming: ' + err.message);
        } finally {
            setIsConnecting(false);
        }
    };

    // Stop streaming
    const stopStreaming = () => {
        // Close peer connection
        if (peerConnectionRef.current) {
            peerConnectionRef.current.close();
            peerConnectionRef.current = null;
        }

        // Close WebSocket
        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }

        // Stop all tracks in the stream
        if (streamRef.current) {
            streamRef.current.getTracks().forEach(track => track.stop());
            streamRef.current = null;
        }

        // Clear video element
        if (videoRef.current) {
            videoRef.current.srcObject = null;
        }

        setIsStreaming(false);
        setConnectedClients(0);
    };

    // Cleanup on unmount
    useEffect(() => {
        return () => {
            stopStreaming();
        };
    }, []);

    // If not in secure context, show warning and disable controls
    if (!isSecureContext) {
        return (
            <div className="VideoStreamer-Root">
                <div className="p-6 bg-yellow-50 border border-yellow-200 rounded-lg">
                    <div className="flex items-start gap-4">
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-yellow-600 flex-shrink-0 mt-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                        </svg>
                        <div>
                            <h3 className="text-lg font-medium text-yellow-800 mb-2">Secure Context Required</h3>
                            <p className="text-yellow-700 mb-4">
                                Camera access requires a secure context. This page must be accessed via:
                            </p>
                            <ul className="list-disc list-inside text-yellow-700 space-y-1 mb-4">
                                <li><code className="bg-yellow-100 px-2 py-0.5 rounded">https://</code> (recommended for production)</li>
                                <li><code className="bg-yellow-100 px-2 py-0.5 rounded">http://localhost:8080</code> (for local development)</li>
                            </ul>
                            <p className="text-sm text-yellow-600">
                                Please update your URL and refresh the page.
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="VideoStreamer-Root">
            {/* Device Selection */}
            <div className="mb-6">
                <label htmlFor="device-select" className="block text-sm font-medium text-gray-700 mb-2">
                    Select Video Input Device
                </label>
                <div className="flex gap-4">
                    <select
                        id="device-select"
                        value={selectedDeviceId}
                        onChange={(e) => setSelectedDeviceId(e.target.value)}
                        disabled={isStreaming || isConnecting}
                        className="flex-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm disabled:bg-gray-100 disabled:cursor-not-allowed"
                    >
                        <option value="">-- Select a camera --</option>
                        {devices.map((device) => (
                            <option key={device.deviceId} value={device.deviceId}>
                                {device.label || `Camera ${devices.indexOf(device) + 1}`}
                            </option>
                        ))}
                    </select>

                    <button
                        onClick={async () => {
                            const deviceList = await navigator.mediaDevices.enumerateDevices();
                            const videoDevices = deviceList.filter(device => device.kind === 'videoinput');
                            setDevices(videoDevices);
                        }}
                        disabled={isStreaming || isConnecting || !navigator.mediaDevices}
                        className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                        title="Refresh device list"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                        </svg>
                    </button>
                </div>
            </div>

            {/* Control Buttons */}
            <div className="flex gap-4 mb-6">
                {!isStreaming ? (
                    <button
                        onClick={startStreaming}
                        disabled={!selectedDeviceId || isConnecting}
                        className="flex items-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
                    >
                        {isConnecting ? (
                            <>
                                <svg className="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                </svg>
                                Connecting...
                            </>
                        ) : (
                            <>
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                </svg>
                                Start Streaming
                            </>
                        )}
                    </button>
                ) : (
                    <button
                        onClick={stopStreaming}
                        className="flex items-center gap-2 px-6 py-3 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors font-medium"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
                        </svg>
                        Stop Streaming
                    </button>
                )}

                {isStreaming && (
                    <div className="flex items-center gap-2 text-sm text-gray-600">
                        <span className="relative flex h-3 w-3">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                        </span>
                        Live • {connectedClients} viewer{connectedClients !== 1 ? 's' : ''}
                    </div>
                )}
            </div>

            {/* Error Message */}
            {error && (
                <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md">
                    <p className="text-red-700 text-sm">{error}</p>
                    <button
                        onClick={() => setError(null)}
                        className="mt-2 text-red-600 hover:text-red-800 text-sm underline"
                    >
                        Dismiss
                    </button>
                </div>
            )}

            {/* Video Preview */}
            <div className="relative bg-gray-900 rounded-lg overflow-hidden aspect-video">
                <video
                    ref={videoRef}
                    autoPlay
                    playsInline
                    muted
                    className="w-full h-full object-cover"
                />

                {!isStreaming && (
                    <div className="absolute inset-0 flex items-center justify-center bg-gray-800">
                        <div className="text-center">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 text-gray-400 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                            </svg>
                            <p className="text-gray-400">Camera preview will appear here</p>
                        </div>
                    </div>
                )}
            </div>

            {/* Device Info */}
            {devices.length > 0 && (
                <div className="mt-6 p-4 bg-gray-50 rounded-md">
                    <h3 className="text-sm font-medium text-gray-700 mb-2">Available Cameras ({devices.length})</h3>
                    <ul className="text-sm text-gray-600 space-y-1">
                        {devices.map((device) => (
                            <li key={device.deviceId} className="flex items-center gap-2">
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
                                </svg>
                                {device.label || `Camera ${devices.indexOf(device) + 1}`}
                                {device.deviceId === selectedDeviceId && <span className="text-blue-600 font-medium">(selected)</span>}
                            </li>
                        ))}
                    </ul>
                </div>
            )}

            {devices.length === 0 && !error && (
                <div className="mt-6 p-4 bg-yellow-50 border border-yellow-200 rounded-md">
                    <p className="text-yellow-800 text-sm">No video input devices found. Please connect a camera and refresh the page.</p>
                </div>
            )}
        </div>
    );
}

export default VideoStreamer;
