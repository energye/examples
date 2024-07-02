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


// render process receive process message
(function () {
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

    class Energy {
        // js ipc.on event listener
        // @key {string} event name
        // @value {function} listener object
        #eventListeners;
        // js ipc.emit event callbacks
        // @key {number} executionID
        // @value {function} callback
        #jsEmitCallbacks;
        // callback executionID
        #executionID;

        /**
         * Creates an instance of Energy.
         * @memberof Energy
         */
        constructor() {
            this.#eventListeners = {};
            this.#jsEmitCallbacks = {};
            this.#executionID = 0;
        }

        /**
         * @param {object} jsonObject
         */
        #notifyListeners(jsonObject) {
            let eventName = jsonObject.name;
            const listener = this.#eventListeners[eventName]
            if (listener) {
                listener.Callback(jsonObject.data);
            }
        };

        /**
         * @param {string} name
         * @param {function} callback
         */
        setEventListener(name, callback) {
            this.#eventListeners[name] = new Listener(name, callback, -1);
        }

        /**
         * @param {string} name
         */
        getEventListener(name) {
            return this.#eventListeners[name];
        }

        /**
         * @param {number} executionID
         * @param {function} callback
         */
        setJSEmitCallback(executionID, callback) {
            this.#jsEmitCallbacks[executionID] = callback;
        }

        /**
         * @param {number} executionID
         */
        getEnergyEventListener(executionID) {
            return this.#jsEmitCallbacks[executionID];
        }

        /**
         * @param {string} messageData
         */
        executeEvent(messageData) {
            let jsonObject;
            try {
                jsonObject = JSON.parse(messageData);
            } catch (e) {
                throw new Error('Invalid JSON passed to Notify: ' + messageData);
            }
            this.#notifyListeners(jsonObject);
        };

        nextExecutionID() {
            this.#executionID++;
            return this.#executionID;
        };
    }

    class IPC {
        on(name, callback) {
            if (name && typeof callback === 'function') {
                // __energyEventListeners[name] = __energyEventListeners[name] || [];
                // __energyEventListeners[name].push(thisListener);
                window.energy.setEventListener(name, callback);
            }
        }

        emit(name, ...args) {
            if (!name) {
                throw new Error('ipc.emit call event name is null');
            } else if (args.length > 2) {
                throw new Error('Invalid ipc.emit call arguments');
            }
            let data = null;
            let callback = null;
            let executionID = 0;
            if (args.length === 1) {
                let arg0 = args[0];
                if (Array.isArray(arg0)) {
                    data = arg0;
                } else if (typeof arg0 === 'function') {
                    callback = arg0;
                } else {
                    throw new Error('Invalid ipc.emit call parameter');
                }
            } else if (args.length === 2) {
                let arg0 = args[0];
                let arg1 = args[1];
                if (Array.isArray(arg0) && typeof arg1 === 'function') {
                    data = arg0;
                    callback = arg1;
                } else {
                    throw new Error('Invalid ipc.emit call arguments');
                }
            }
            if (callback !== null) {
                executionID = window.energy.nextExecutionID();
                window.energy.setJSEmitCallback(executionID, callback)
            }
            const payload = {
                n: name, // name
                d: [].slice.apply(data), // data
                i: executionID, // executionID
            };
            // call js event

            // call go event
            ProcessMessage(JSON.stringify(payload));
        }
    }

    window.energy = new Energy();
    window.ipc = new IPC();
})();
