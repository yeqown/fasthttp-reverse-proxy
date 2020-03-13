const WebSocket = require('ws')

// host = "ws://localhost:8080/echo" // backend 
host = "ws://localhost:8081/echo" // proxy

function benchmark() {
    setInterval(() => {
        console.log("current avg delta(ms): ", avgDelta.toFixed(3));
    }, 2000)

    for (let i = 0; i < 100; i++) {
        delay_test()
    }
}


var testCnt = 0
var avgDelta = 0

function count_delay(delta) {
    testCnt++
    if (testCnt === 0) {
        avgDelta = delta
        return
    }

    avgDelta = avgDelta * ((testCnt - 1) / testCnt) + delta * (1 / testCnt)
}

function delay_test() {
    var c = new WebSocket(host)
    c.onmessage = (evt) => {
        let now = (new Date()).getTime()
        count_delay(now - (+evt.data))
        // console.log("recv:", evt.data)
    }
    c.onerror = (evt) => {
        console.error("error:", evt.data)
    }
    c.onopen = (e) => {
        setInterval(() => {
            c.send((new Date()).getTime().toString())
        }, 100)
    }
}

benchmark()
