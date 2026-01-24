// Chat functionality
function addChatMessage(message, isUser = false) {
    const chatHistory = document.getElementById('chatHistory');
    if (!chatHistory) {
        console.error('chatHistory element not found');
        return;
    }
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${isUser ? 'user-message' : 'ai-message'}`;
    messageDiv.textContent = message;
    chatHistory.appendChild(messageDiv);
    chatHistory.scrollTop = chatHistory.scrollHeight;
}

async function askGemini(question) {
    console.log('Making API call to /ask-gemini with question:', question);
    try {
        const response = await fetch('/ask-gemini', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ question })
        });
        
        console.log('Response status:', response.status);
        
        if (!response.ok) {
            const errorText = await response.text();
            console.error('Response error:', errorText);
            throw new Error(`Server error: ${response.status} - ${errorText}`);
        }
        
        const data = await response.json();
        console.log('Response data:', data);
        
        if (data.error) {
            throw new Error(data.error);
        }
        
        // Handle different response formats
        if (data.explanation) {
            return data.explanation;
        } else if (data.answer) {
            return data.answer;
        } else if (data.data) {
            // If we have structured data, format it nicely
            return JSON.stringify(data.data, null, 2);
        } else {
            return 'I received your question but got an unexpected response format.';
        }
    } catch (error) {
        console.error('Error asking Gemini:', error);
        throw error;
    }
}

async function handleQuestion(question) {
    console.log('handleQuestion called with:', question);
    const askBtn = document.getElementById('askBtn');
    const questionInput = document.getElementById('questionInput');
    
    if (!askBtn || !questionInput) {
        console.error('Missing elements:', { askBtn: !!askBtn, questionInput: !!questionInput });
        return;
    }
    
    // Disable input during processing
    askBtn.disabled = true;
    askBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
    questionInput.disabled = true;
    
    // Add user message
    addChatMessage(question, true);
    
    try {
        const answer = await askGemini(question);
        addChatMessage(answer, false);
    } catch (error) {
        addChatMessage('Sorry, I encountered an error: ' + error.message, false);
        console.error('Error in chat:', error);
    } finally {
        // Re-enable input
        askBtn.disabled = false;
        askBtn.innerHTML = '<i class="fas fa-paper-plane"></i>';
        questionInput.disabled = false;
        questionInput.value = '';
        questionInput.focus();
    }
}

// Function to handle send button click (exposed globally immediately)
function handleSendClick() {
    console.log('=== handleSendClick called ===');
    const questionInput = document.getElementById('questionInput');
    if (!questionInput) {
        console.error('questionInput not found');
        alert('Input field not found!');
        return;
    }
    const question = questionInput.value.trim();
    console.log('Question value:', question);
    if (question) {
        handleQuestion(question).catch(err => {
            console.error('Error in handleQuestion:', err);
        });
    } else {
        console.log('No question entered');
        alert('Please enter a question');
    }
}

// Make it globally available immediately
window.handleSendClick = handleSendClick;

// Event listeners - set up when DOM is ready
(function() {
    function init() {
        console.log('Initializing chat interface...');
        const askBtn = document.getElementById('askBtn');
        const questionInput = document.getElementById('questionInput');
        
        if (!askBtn) {
            console.error('Could not find askBtn element');
            return;
        }
        if (!questionInput) {
            console.error('Could not find questionInput element');
            return;
        }
        
        console.log('Elements found, attaching event listeners...');
        
        // Remove any existing onclick to avoid conflicts
        askBtn.removeAttribute('onclick');
        
        // Add click event listener
        askBtn.addEventListener('click', function(e) {
            console.log('Button clicked via addEventListener');
            e.preventDefault();
            e.stopPropagation();
            handleSendClick();
        });
        
        // Add Enter key handler
        questionInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                console.log('Enter key pressed');
                e.preventDefault();
                handleSendClick();
            }
        });
        
        // Focus on input
        questionInput.focus();
        console.log('Chat interface initialized successfully');
    }
    
    // Run when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        // DOM is already ready
        init();
    }
})();
