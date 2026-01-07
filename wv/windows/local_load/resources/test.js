function myAlertFunction() {
    alert("Hello! I am an alert box!!");
}

(function () {
    console.log('test.js')
    console.log(energy)
    console.log(ipc)
    console.log(window)
})()

window.setInterval(function () {
    if (ipc) {
        console.log('ipc.emit-test')
        ipc.emit("test", [new Date().toString()], function (res1, res2, res3, res4) {
            console.log("result:", res1, res2, res3, res4)
        })
    }
}, 1000)

ipc.on('test', function (data) {
    console.log('ipc.on-test:', data)
})