(function () {
    class Browser {
        #windowId = {windowId};
        #frameId = {frameId};

        windowId() {
            return this.#windowId;
        }

        frameId() {
            return this.#frameId;
        }
    }

    window.browser = new Browser();
})();