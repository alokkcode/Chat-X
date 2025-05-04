// chat.js - WebSocket Chat Implementation

// Core variables
let socket;
const currentUserID = parseInt(document.getElementById("current-user-id").value);
const currentUserRole = document.getElementById("current-user-role").value;
const roomID = document.getElementById("room-id").value;

// Connect to WebSocket
function connectWebSocket() {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsURL = `${wsProtocol}//${window.location.host}/ws?room=${roomID}&user=${encodeURIComponent(currentUserID)}`;
    
    console.log("Connecting to WebSocket:", wsURL);
    socket = new WebSocket(wsURL);

    socket.onopen = function() {
        console.log("WebSocket connected successfully");
        // Enable input field once connected
        document.getElementById("messageInput").disabled = false;
        document.getElementById("messageInput").focus();
    };

    socket.onclose = function(event) {
        console.log("WebSocket disconnected with code:", event.code);
        document.getElementById("messageInput").disabled = true;
        
        // Attempt to reconnect unless it was a normal closure
        if (event.code !== 1000) {
            console.log("Attempting to reconnect in 3 seconds...");
            setTimeout(connectWebSocket, 3000);
        }
    };

    socket.onerror = function(error) {
        console.error("WebSocket error occurred:", error);
    };

    socket.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            handleIncomingMessage(data);
        } catch (e) {
            console.error("Error parsing message:", e);
            appendSystemMessage({ text: "Error: Failed to parse incoming message" });
        }
    };
}

// Handle different types of incoming messages
function handleIncomingMessage(data) {
    switch (data.type) {
        case "chat":
            appendMessage(data);
            break;
        case "system":
            appendSystemMessage(data);
            break;
        case "delete":
            if (data.message_id) removeMessage(data.message_id);
            break;
        case "user_joined":
            appendSystemMessage({ text: `${data.username} joined the chat` });
            break;
        case "user_left":
            appendSystemMessage({ text: `${data.username} left the chat` });
            break;
        default:
            console.warn("Unknown message type:", data.type);
    }
}

// Append a chat message to the conversation
function appendMessage(message) {
    const messageContainer = document.getElementById("chatMessages");
    const messageDiv = document.createElement("div");
    messageDiv.className = "message";
    messageDiv.dataset.messageId = message.message_id;

    // Determine if current user can delete this message
    const canDelete = (currentUserID === message.user_id) || (currentUserRole === "admin");

    // Create message HTML with appropriate formatting
    messageDiv.innerHTML = `
        <span class="username ${message.role || ''}">${escapeHtml(message.username)}: </span>
        <span class="message-text">${escapeHtml(message.text)}</span>
        <span class="message-time">${formatTimestamp(message.timestamp)}</span>
        ${canDelete ? `<button class="delete-btn" onclick="deleteMessage(${message.message_id})">🗑️ Delete</button>` : ''}
    `;

    messageContainer.appendChild(messageDiv);
    scrollToBottom(messageContainer);
}

// Append system messages (notifications, errors, etc.)
function appendSystemMessage(message) {
    const messageContainer = document.getElementById("chatMessages");
    const messageDiv = document.createElement("div");
    messageDiv.className = "system-message";
    messageDiv.innerHTML = `<em>${escapeHtml(message.text)}</em>`;
    messageContainer.appendChild(messageDiv);
    scrollToBottom(messageContainer);
}

// Remove a message from the DOM when deleted
function removeMessage(messageId) {
    const message = document.querySelector(`.message[data-message-id="${messageId}"]`);
    if (message) {
        message.classList.add("deleting");
        setTimeout(() => message.remove(), 300);
    }
}

// Send a message via the WebSocket
function sendMessage() {
    const input = document.getElementById("messageInput");
    const message = input.value.trim();

    if (message && socket && socket.readyState === WebSocket.OPEN) {
        socket.send(message);
        input.value = "";
        input.focus();
    } else if (socket.readyState !== WebSocket.OPEN) {
        appendSystemMessage({ text: "Connection lost. Attempting to reconnect..." });
        connectWebSocket();
    }
}

// Delete a message via API call
function deleteMessage(messageId) {
    if (!messageId) {
        console.error("Invalid message ID for deletion");
        return;
    }

    const requestData = { message_id: messageId, room_id: parseInt(roomID) }; // Ensure `room_id` is an integer

    console.log("Sending delete request:", requestData); // Log request data before sending

    fetch("/api/delete-message", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(requestData)
    })
    .then(response => response.json())
    .then(data => {
        console.log("Message deleted successfully:", data);
    })
    .catch(error => {
        console.error("Error deleting message:", error);
        appendSystemMessage({ text: "Failed to delete message. Please try again." });
    });
}

// Helper functions
function escapeHtml(text) {
    if (!text) return "";
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

function formatTimestamp(timestamp) {
    if (!timestamp) return "";
    const date = new Date(timestamp);
    return isNaN(date) ? "" : date.toLocaleTimeString();
}

function scrollToBottom(element) {
    element.scrollTop = element.scrollHeight;
}

// Initialize everything when DOM is ready
document.addEventListener("DOMContentLoaded", function() {
    // Connect to WebSocket
    connectWebSocket();

    // Set up event listeners
    document.getElementById("messageInput").addEventListener("keypress", function(e) {
        if (e.key === "Enter") {
            e.preventDefault();
            sendMessage();
        }
    });

    document.getElementById("sendButton").addEventListener("click", function(e) {
        e.preventDefault();
        sendMessage();
    });
    
    // Prevent form submission
    document.getElementById("messageForm").addEventListener("submit", function(e) {
        e.preventDefault();
    });
});
