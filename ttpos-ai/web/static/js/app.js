// TTPOS AI Chat Application
class ChatApp {
    constructor() {
        this.STORAGE_KEY_CHAT = 'ttpos_ai_chat_messages';
        this.STORAGE_KEY_AGENT = 'ttpos_ai_agent_messages';
        this.STORAGE_KEY_MODE = 'ttpos_ai_mode';

        this.messages = [];
        this.mode = 'chat'; // 'chat' or 'agent'
        this.isStreaming = true;
        this.isLoading = false;

        this.initElements();
        this.initEventListeners();
        this.loadFromStorage();
    }

    initElements() {
        this.messagesContainer = document.getElementById('messages');
        this.userInput = document.getElementById('userInput');
        this.sendBtn = document.getElementById('sendBtn');
        this.streamToggle = document.getElementById('streamToggle');
        this.statusEl = document.getElementById('status');
        this.chatModeBtn = document.getElementById('chatMode');
        this.agentModeBtn = document.getElementById('agentMode');
        this.clearBtn = document.getElementById('clearBtn');
    }

    initEventListeners() {
        // Send button
        this.sendBtn.addEventListener('click', () => this.sendMessage());

        // Enter to send
        this.userInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        // Auto-resize textarea
        this.userInput.addEventListener('input', () => {
            this.userInput.style.height = 'auto';
            this.userInput.style.height = Math.min(this.userInput.scrollHeight, 150) + 'px';
        });

        // Stream toggle
        this.streamToggle.addEventListener('change', (e) => {
            this.isStreaming = e.target.checked;
        });

        // Mode switch
        this.chatModeBtn.addEventListener('click', () => this.setMode('chat'));
        this.agentModeBtn.addEventListener('click', () => this.setMode('agent'));

        // Clear button
        if (this.clearBtn) {
            this.clearBtn.addEventListener('click', () => this.clearChat());
        }
    }

    // Load messages from localStorage
    loadFromStorage() {
        // Load saved mode
        const savedMode = localStorage.getItem(this.STORAGE_KEY_MODE);
        if (savedMode === 'chat' || savedMode === 'agent') {
            this.mode = savedMode;
            this.chatModeBtn.classList.toggle('active', this.mode === 'chat');
            this.agentModeBtn.classList.toggle('active', this.mode === 'agent');
        }

        // Load saved messages
        const storageKey = this.mode === 'chat' ? this.STORAGE_KEY_CHAT : this.STORAGE_KEY_AGENT;
        const savedMessages = localStorage.getItem(storageKey);

        if (savedMessages) {
            try {
                const parsed = JSON.parse(savedMessages);
                if (Array.isArray(parsed) && parsed.length > 0) {
                    // Restore messages
                    for (const msg of parsed) {
                        this.messages.push(msg);
                        this.renderMessage(msg.role, msg.content);
                    }
                    this.scrollToBottom();
                    return;
                }
            } catch (e) {
                console.error('Failed to parse saved messages:', e);
            }
        }

        // No saved messages, show welcome
        this.addWelcomeMessage();
    }

    // Save messages to localStorage
    saveToStorage() {
        const storageKey = this.mode === 'chat' ? this.STORAGE_KEY_CHAT : this.STORAGE_KEY_AGENT;
        localStorage.setItem(storageKey, JSON.stringify(this.messages));
        localStorage.setItem(this.STORAGE_KEY_MODE, this.mode);
    }

    // Clear current chat
    clearChat() {
        if (!confirm('Are you sure you want to clear chat history?')) {
            return;
        }

        this.messages = [];
        this.messagesContainer.textContent = '';

        // Clear from storage
        const storageKey = this.mode === 'chat' ? this.STORAGE_KEY_CHAT : this.STORAGE_KEY_AGENT;
        localStorage.removeItem(storageKey);

        this.addWelcomeMessage();
        this.setStatus('Chat cleared', 'success');
    }

    setMode(mode) {
        if (this.mode === mode) return;

        this.mode = mode;
        this.chatModeBtn.classList.toggle('active', mode === 'chat');
        this.agentModeBtn.classList.toggle('active', mode === 'agent');

        // Save current mode
        localStorage.setItem(this.STORAGE_KEY_MODE, mode);

        // Clear UI and load messages for new mode
        this.messages = [];
        this.messagesContainer.textContent = '';
        this.loadFromStorage();
    }

    addWelcomeMessage() {
        const welcomeText = this.mode === 'chat'
            ? 'Hello! I\'m TTPOS AI Assistant. How can I help you?'
            : 'Hello! I\'m TTPOS Data Assistant. You can ask me questions about your data, such as "How many orders today?" or "What are the best-selling items?"';

        this.addMessage('assistant', welcomeText);
    }

    async sendMessage() {
        const text = this.userInput.value.trim();
        if (!text || this.isLoading) return;

        // Add user message
        this.addMessage('user', text);
        this.userInput.value = '';
        this.userInput.style.height = 'auto';

        // Show loading
        this.setLoading(true);
        const loadingEl = this.addLoadingIndicator();

        try {
            if (this.mode === 'agent') {
                await this.sendAgentQuery(text, loadingEl);
            } else if (this.isStreaming) {
                await this.sendStreamMessage(text, loadingEl);
            } else {
                await this.sendNormalMessage(text, loadingEl);
            }
        } catch (error) {
            this.removeElement(loadingEl);
            this.addMessage('assistant', 'Error: ' + error.message);
            this.setStatus('Request failed', 'error');
        } finally {
            this.setLoading(false);
        }
    }

    async sendNormalMessage(text, loadingEl) {
        const response = await fetch('/api/v1/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                messages: this.getApiMessages(),
                stream: false
            })
        });

        const data = await response.json();
        this.removeElement(loadingEl);

        if (data.code === 1 && data.data) {
            this.addMessage('assistant', data.data.content);
            if (data.data.usage) {
                this.setStatus('Tokens: ' + data.data.usage.total_tokens, 'success');
            }
        } else {
            this.addMessage('assistant', 'Error: ' + data.message);
        }
    }

    async sendStreamMessage(text, loadingEl) {
        const response = await fetch('/api/v1/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                messages: this.getApiMessages(),
                stream: true
            })
        });

        this.removeElement(loadingEl);

        const messageEl = this.addMessage('assistant', '');
        const contentEl = messageEl.querySelector('.message-content');
        let fullContent = '';

        const reader = response.body.getReader();
        const decoder = new TextDecoder();

        while (true) {
            const { done, value } = await reader.read();
            if (done) break;

            const chunk = decoder.decode(value);
            const lines = chunk.split('\n');

            for (const line of lines) {
                if (line.startsWith('data:')) {
                    try {
                        const data = JSON.parse(line.slice(5));
                        if (data.content) {
                            fullContent += data.content;
                            this.safeSetContent(contentEl, fullContent);
                            this.scrollToBottom();
                        }
                    } catch (e) {
                        // Ignore parse errors
                    }
                }
            }
        }

        // Update messages array and save
        this.messages[this.messages.length - 1].content = fullContent;
        this.saveToStorage();
    }

    async sendAgentQuery(text, loadingEl) {
        const response = await fetch('/api/v1/agent/query', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ question: text })
        });

        const data = await response.json();
        this.removeElement(loadingEl);

        if (data.code === 1 && data.data) {
            let content = data.data.answer;

            // Add SQL info if available
            if (data.data.sql_used) {
                content += '\n\n---\n**SQL Query:**\n```sql\n' + data.data.sql_used + '\n```';
            }

            this.addMessage('assistant', content);
            this.setStatus('Query completed', 'success');
        } else {
            this.addMessage('assistant', 'Error: ' + data.message);
            this.setStatus('Query failed', 'error');
        }
    }

    getApiMessages() {
        return this.messages.map(m => ({
            role: m.role,
            content: m.content
        }));
    }

    addMessage(role, content) {
        this.messages.push({ role, content });
        this.saveToStorage();
        return this.renderMessage(role, content);
    }

    // Render message without saving (used for restoring from storage)
    renderMessage(role, content) {
        const messageEl = document.createElement('div');
        messageEl.className = 'message ' + role;

        const avatar = document.createElement('div');
        avatar.className = 'avatar';
        avatar.textContent = role === 'user' ? 'U' : 'AI';

        const contentEl = document.createElement('div');
        contentEl.className = 'message-content';
        this.safeSetContent(contentEl, content);

        messageEl.appendChild(avatar);
        messageEl.appendChild(contentEl);
        this.messagesContainer.appendChild(messageEl);

        this.scrollToBottom();
        return messageEl;
    }

    // Safe content rendering using marked library (which sanitizes by default)
    safeSetContent(el, content) {
        if (typeof marked !== 'undefined') {
            // marked.parse with default options sanitizes HTML
            el.innerHTML = marked.parse(content, { breaks: true });
        } else {
            // Fallback: use textContent for safety
            el.textContent = content;
        }
    }

    addLoadingIndicator() {
        const messageEl = document.createElement('div');
        messageEl.className = 'message assistant';

        const avatar = document.createElement('div');
        avatar.className = 'avatar';
        avatar.textContent = 'AI';

        const contentEl = document.createElement('div');
        contentEl.className = 'message-content';

        // Create typing indicator using DOM methods
        const typingDiv = document.createElement('div');
        typingDiv.className = 'typing-indicator';
        for (let i = 0; i < 3; i++) {
            typingDiv.appendChild(document.createElement('span'));
        }
        contentEl.appendChild(typingDiv);

        messageEl.appendChild(avatar);
        messageEl.appendChild(contentEl);
        this.messagesContainer.appendChild(messageEl);

        this.scrollToBottom();
        return messageEl;
    }

    removeElement(el) {
        if (el && el.parentNode) {
            el.parentNode.removeChild(el);
        }
    }

    setLoading(loading) {
        this.isLoading = loading;
        this.sendBtn.disabled = loading;
        this.userInput.disabled = loading;
    }

    setStatus(text, type) {
        this.statusEl.textContent = text;
        this.statusEl.className = 'status ' + (type || '');

        // Auto clear after 5 seconds
        const currentText = text;
        setTimeout(() => {
            if (this.statusEl.textContent === currentText) {
                this.statusEl.textContent = '';
                this.statusEl.className = 'status';
            }
        }, 5000);
    }

    scrollToBottom() {
        this.messagesContainer.scrollTop = this.messagesContainer.scrollHeight;
    }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.chatApp = new ChatApp();
});
