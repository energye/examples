class Listener {
    /**
     * Creates an instance of Listener.
     * @param {string} eventName
     * @param {function} callback
     * @param {number} maxCallbacks
     * @memberof Listener
     */
    constructor(eventName, callback, maxCallbacks) {
        this.eventName = eventName;
        // Default of -1 means infinite
        this.maxCallbacks = maxCallbacks || -1;
        // Callback invokes the callback with the given data
        // Returns true if this listener should be destroyed
        this.Callback = (data) => {
            callback.apply(null, data);
            // If maxCallbacks is infinite, return false (do not destroy)
            if (this.maxCallbacks === -1) {
                return false;
            }
            // Decrement maxCallbacks. Return true if now 0, otherwise false
            this.maxCallbacks -= 1;
            return this.maxCallbacks === 0;
        };
    }
}


// render process send process message
(function () {
    // Credit: https://stackoverflow.com/a/2631521
    let _deeptest = function (s) {
        var obj = window[s.shift()];
        while (obj && s.length) obj = obj[s.shift()];
        return obj;
    };
    let windows = _deeptest(["chrome", "webview", "postMessage"]);
    let mac_linux = _deeptest(["webkit", "messageHandlers", "external", "postMessage"]);
    if (!windows && !mac_linux) {
        console.error("Unsupported Platform");
        return;
    }
    if (windows) {
        window.ProcessMessage = (message) => window.chrome.webview.postMessage(message);
    }
    if (mac_linux) {
        window.ProcessMessage = (message) => window.webkit.messageHandlers.external.postMessage(message);
    }
})();

// js ipc.on event listener
// @key {string} event name
// @value {function} listener object
const __energyEventListeners = {};

// js ipc.emit event callbacks
// @key {number} messageId
// @value {function} callback
const __energyJSEmitCallbacks = {};

// render process receive process message
(function () {
    window.energy = {
        messageId: 0,
        // go: execute js event
        executeEvent: (messageData) => {
            let jsonObject;
            try {
                jsonObject = JSON.parse(messageData);
            } catch (e) {
                throw new Error('Invalid JSON passed to Notify: ' + messageData);
            }
            this.__notifyListeners(jsonObject);
        },
        // go: execute js event
        // internal
        __notifyListeners: (jsonObject) => {
            let eventName = jsonObject.name;
            const listener = __energyEventListeners[eventName]
            if (listener) {
                listener.Callback(jsonObject.data);
            }
        },
        __nextMessageId: () => {
            this.messageId++;
            return this.messageId;
        }
    };
})();

// js ipc
const ipc = {
    // js ipc.on
    on: (name, callback) => {
        if (name !== '' && typeof callback === 'function') {
            // __energyEventListeners[name] = __energyEventListeners[name] || [];
            // __energyEventListeners[name].push(thisListener);
            __energyEventListeners[name] = new Listener(name, callback, -1);
        }
    },
    // js ipc.emit
    emit: (name, ...arguments) => {
        if (!name) {
            throw new Error('ipc.emit call event name is null');
        } else if (arguments.length > 2) {
            throw new Error('Invalid ipc.emit call arguments');
        }
        let data = null;
        let callback = null;
        let messageId = 0;
        if (arguments.length === 1) {
            let arg0 = arguments[0];
            if (Array.isArray(arg0)) {
                data = arg0;
            } else if (typeof arg0 === 'function') {
                callback = arg0;
            } else {
                throw new Error('Invalid ipc.emit call parameter');
            }
        } else if (arguments.length === 2) {
            let arg0 = arguments[0];
            let arg1 = arguments[1];
            if (Array.isArray(arg0) && typeof arg1 === 'function') {
                data = arg0;
                callback = arg1;
            } else {
                throw new Error('Invalid ipc.emit call arguments');
            }
        }
        if (callback !== null) {
            messageId = window.energy.__nextMessageId();
            __energyJSEmitCallbacks[messageId] = callback;
        }
        const payload = {
            name: name,
            data: [].slice.apply(data),
            messageId: messageId,
        };
        // call js event

        // call go event
        ProcessMessage(JSON.stringify(payload));
    }
};
