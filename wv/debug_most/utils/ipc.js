// render process send process message
(function () {
    // Listener
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

    // Energy
    class Energy {
        // js ipc.on event listener
        // @key {string} event name
        // @value {Listener} listener object
        #eventListeners;

        // js ipc.emit callbacks
        // @key {number} executionID
        // @value {function} callback
        #emitCallbacks;

        // js ipc.emit callback executionID, global accumulation
        #executionID;

        /**
         * Creates an instance of Energy.
         * @memberof Energy
         */
        constructor() {
            this.#eventListeners = {};
            this.#emitCallbacks = {};
            this.#executionID = 0;
        }

        /**
         * @param {object} message
         * type ProcessMessage struct {
         * 	Name string        `json:"n"`
         * 	Data []interface{} `json:"d"`
         * 	Id   int           `json:"i"`
         * }
         */
        #notifyListeners(message) {
            let id = message.i;
            let name = message.n;
            let callback;
            if (!name && id !== 0) {
                callback = this.#emitCallbacks[id];
                if (callback) {
                    delete this.#emitCallbacks[id];
                }
            } else {
                callback = this.#eventListeners[name];
            }
            if (callback) {
                return callback.apply(null, message.d);
            }
        };

        /**
         * @param {string} name
         * @param {function} callback
         */
        setEventListener(name, callback) {
            this.#eventListeners[name] = callback;
        }

        /**
         * @param {number} executionID
         * @param {function} callback
         */
        setJSEmitCallback(executionID, callback) {
            this.#emitCallbacks[executionID] = callback;
        }

        /**
         * @param {string} messageData
         */
        executeEvent(messageData) {
            try {
                const result = this.#notifyListeners(JSON.parse(messageData));
                if (result) {
                    return result
                }
            } catch (e) {
                throw new Error('Invalid JSON passed to Notify: ' + messageData);
            }
        };

        nextExecutionID() {
            this.#executionID++;
            return this.#executionID;
        };
    }

    // IPC
    class IPC {
        /**
         * @param {string} name
         * @param {function} callback
         */
        on(name, callback) {
            if (name && typeof callback === 'function') {
                // __energyEventListeners[name] = __energyEventListeners[name] || [];
                // __energyEventListeners[name].push(thisListener);
                window.energy.setEventListener(name, callback);
            }
        }

        /**
         * @param {string} name
         * @param {argument} args
         */
        emit(name, ...args) {
            if (!name) {
                throw new Error('ipc.emit call event name is null');
            } else if (args.length > 2) {
                throw new Error('Invalid ipc.emit call arguments');
            }
            let data = [];
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
                d:  [].slice.apply(data), // data
                i: executionID, // executionID
            };
            // call js event

            // call go event
            ProcessMessage(JSON.stringify(payload));
        }
    }

    window.energy = new Energy();
    window.ipc = new IPC();

    let deepTest = function (s) {
        let obj = window[s.shift()];
        while (obj && s.length) obj = obj[s.shift()];
        return obj;
    };
    if (deepTest(["chrome", "webview", "postMessage"])) {
        // webview2
        let webview = window.chrome.webview;
        // render process send message => go
        window.ProcessMessage = (message) => webview.postMessage(message);
        // render process receive browser process string message
        webview.addEventListener("message", event => {
            console.log("message:", event);
            const result = window.energy.executeEvent(event.data);

        });
        // render process receive browser process buffer message
        webview.addEventListener("sharedbufferreceived", event => {
            let buffer = event.getBuffer();
            let bufferData = new TextDecoder().decode(new Uint8Array(buffer));
            console.log("buffer:", bufferData);
        });
    } else if (deepTest(["webkit", "messageHandlers", "external", "postMessage"])) {
        // webkit
        // render process send message => go
        window.ProcessMessage = (message) => window.webkit.messageHandlers.external.postMessage(message);
    } else {
        console.error("Unsupported Platform");
    }
})();
