import React, { useState, useEffect, useRef } from 'react';

/**
 * VideoViewer component for watching a stream.
 */
function VideoViewer({ wsUrl, streamer, initialViewerCount }) {
    const [isConnected, setIsConnected] = useState(false);
    const [isConnecting, setIsConnecting] = useState(false);
    const [error, setError] = useState(null);
    const [viewerCount, setViewerCount] = useState(initialViewerCount || 0);

    const videoRef = useRef(null);
    const wsRef = useRef(null);
    const pcRef = useRef(null);

    useEffect(() => {
        connectToStream();
        return () => cleanup();
    }, []);

    const cleanup = () => {
        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }
        if (pcRef.current) {
            pcRef.current.close();
            pcRef.current = null;
        }
    };

    const connectToStream = async () => {
        setIsConnecting(true);
        setError(null);

        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const ws = new WebSocket(`${protocol}//${window.location.host}${wsUrl}`);
            wsRef.current = ws;

            ws.onopen = () => {
                console.log('WebSocket connected');
                setIsConnected(true);
                setIsConnecting(false);
                createPeerConnection();
            };

            ws.onmessage = async (event) => {
                const msg = JSON.parse(event.data);
                await handleSignalingMessage(msg);
            };

            ws.onerror = (err) => {
                console.error('WebSocket error:', err);
                setError('Connection error. The streamer may have ended the stream.');
                setIsConnecting(false);
            };

            ws.onclose = () => {
                setIsConnected(false);
            };
        } catch (err) {
            console.error('Failed to connect:', err);
            setError('Failed to connect to stream.');
            setIsConnecting(false);
        }
    };

    const createPeerConnection = async () => {
        const pc = new RTCPeerConnection({
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' }
            ]
        });
        pcRef.current = pc;

        pc.ontrack = (event) => {
            console.log('Received remote track');
            if (videoRef.current) {
                videoRef.current.srcObject = event.streams[0];
            }
        };

        pc.onicecandidate = (event) => {
            if (event.candidate && wsRef.current?.readyState === WebSocket.OPEN) {
                wsRef.current.send(JSON.stringify({
                    type: 'ice-candidate',
                    candidate: event.candidate
                }));
            }
        };

        pc.onconnectionstatechange = () => {
            console.log('Connection state:', pc.connectionState);
        };
    };

    const handleSignalingMessage = async (msg) => {
        const pc = pcRef.current;
        if (!pc) return;

        try {
            switch (msg.type) {
                case 'offer':
                    await pc.setRemoteDescription(new RTCSessionDescription(msg));
                    const answer = await pc.createAnswer();
                    await pc.setLocalDescription(answer);
                    if (wsRef.current?.readyState === WebSocket.OPEN) {
                        wsRef.current.send(JSON.stringify({
                            type: 'answer',
                            sdp: pc.localDescription.sdp
                        }));
                    }
                    break;

                case 'ice-candidate':
                    await pc.addIceCandidate(new RTCIceCandidate(msg.candidate));
                    break;

                case 'error':
                    setError(msg.message || 'Stream error');
                    break;

                default:
                    break;
            }
        } catch (err) {
            console.error('Error handling signaling message:', err);
        }
    };

    return (
        <div>
            <div className="mb-6 flex items-center justify-between">
                <div>
                    <div className="flex items-center gap-3 mb-2">
                        <h1 className="text-3xl font-bold text-gray-900">{streamer}'s Stream</h1>
                        {isConnected && (
                            <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-600 text-white animate-pulse">
                                LIVE
                            </span>
                        )}
                    </div>
                    <div className="flex items-center text-gray-600">
                        <svg className="w-5 h-5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                        </svg>
                        <span id="viewer-count">{viewerCount}</span> viewers
                    </div>
                </div>
                <a
                    href="/rooms"
                    className="inline-flex items-center px-4 py-2 bg-gray-200 hover:bg-gray-300 text-gray-700 font-semibold rounded-lg transition-colors"
                >
                    <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                    </svg>
                    Back to Rooms
                </a>
            </div>

            <div className="bg-black rounded-lg shadow-lg overflow-hidden aspect-video relative">
                <video
                    ref={videoRef}
                    autoPlay
                    playsInline
                    muted
                    controls
                    className="w-full h-full object-contain"
                />

                {!isConnected && (
                    <div className="absolute inset-0 flex items-center justify-center text-white bg-gray-900">
                        <div className="text-center">
                            {isConnecting ? (
                                <>
                                    <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-white mx-auto mb-4"></div>
                                    <p className="text-xl">Connecting to stream...</p>
                                </>
                            ) : error ? (
                                <>
                                    <svg className="w-16 h-16 mx-auto mb-4 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                    </svg>
                                    <p className="text-xl text-red-400 mb-4">{error}</p>
                                    <button
                                        onClick={connectToStream}
                                        className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg transition-colors"
                                    >
                                        Retry
                                    </button>
                                </>
                            ) : (
                                <>
                                    <svg className="w-16 h-16 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                    </svg>
                                    <p className="text-xl">Waiting for streamer...</p>
                                </>
                            )}
                        </div>
                    </div>
                )}
            </div>

            <div className="mt-6 bg-white rounded-lg shadow p-6">
                <h2 className="text-xl font-bold text-gray-900 mb-2">About this stream</h2>
                <p className="text-gray-600">
                    You are watching <strong>{streamer}</strong>'s live stream.
                    This is a WebRTC-based peer-to-peer stream.
                </p>
            </div>
        </div>
    );
}

export default VideoViewer;
