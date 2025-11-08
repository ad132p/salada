import React, { useState } from 'react';

// The main application component
const Share = () => {
  // State to manage the visual feedback after copying
  const [isCopied, setIsCopied] = useState(false);

  // Get the current URL. We wrap this in a try-catch in case window is not available,
  // though it should be in a browser environment.
  let currentUrl = 'Loading URL...';
  try {
    currentUrl = window.location.href;
  } catch (e) {
    currentUrl = 'Error: Cannot access location.';
  }

  /**
   * Handles copying the current URL to the clipboard using the recommended
   * document.execCommand('copy') method for sandboxed environments.
   */
  const handleCopy = () => {
    // 1. Create a temporary textarea element
    const tempInput = document.createElement('textarea');
    tempInput.value = currentUrl;
    
    // 2. Hide the element off-screen
    tempInput.style.position = 'fixed';
    tempInput.style.left = '-9999px';
    tempInput.style.top = '0';
    document.body.appendChild(tempInput);

    // 3. Select the text and execute the copy command
    tempInput.focus();
    tempInput.select();

    try {
      const successful = document.execCommand('copy');
      if (successful) {
        // Set state for visual feedback
        setIsCopied(true);
        // Reset the message after 2 seconds
        setTimeout(() => setIsCopied(false), 2000);
      } else {
        console.error("Fallback: document.execCommand('copy') failed.");
      }
    } catch (err) {
      console.error('Failed to copy text: ', err);
    } finally {
      // 4. Clean up by removing the temporary element
      document.body.removeChild(tempInput);
    }
  };

  return (
      <div className="bg-white p-8 rounded-2xl shadow-xl w-full max-w-sm text-center border border-gray-100">
        <h1 className="text-2xl font-bold text-gray-800 mb-4 tracking-tight">
          Copy Link Tool
        </h1>
        
        <p className="text-sm text-gray-500 mb-6 px-3 py-2 bg-gray-100 rounded-lg overflow-x-auto whitespace-nowrap border border-gray-200">
          {currentUrl}
        </p>

        <button
          onClick={handleCopy}
          className={`
            w-full py-3 px-6 rounded-xl font-semibold text-lg transition duration-300 ease-in-out
            flex items-center justify-center space-x-2
            ${isCopied
              ? 'bg-green-500 hover:bg-green-600 text-white shadow-lg shadow-green-300/50'
              : 'bg-indigo-600 hover:bg-indigo-700 text-white shadow-lg shadow-indigo-300/50'
            }
            transform hover:scale-[1.01] active:scale-[0.99] focus:outline-none focus:ring-4 focus:ring-indigo-500 focus:ring-opacity-50
          `}
          aria-live="polite"
        >
          {/* Icon: Checkmark or Clipboard */}
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            {isCopied ? (
              // Checkmark icon
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
            ) : (
              // Clipboard icon
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            )}
          </svg>


          <span>{isCopied ? 'Copied Successfully!' : 'Copy Current Link'}</span>
        </button>
        
        {/* Helper text for the action */}
        {!isCopied && (
            <p className="mt-4 text-xs text-gray-400">
                Click the button to save the URL above to your clipboard.
            </p>
        )}
      </div>
  );
};

export default Share;